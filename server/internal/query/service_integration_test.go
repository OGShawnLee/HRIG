package query_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"hrig/server/internal/config"
	serverhttp "hrig/server/internal/http"
	"hrig/server/internal/query"
	"hrig/shared/model"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("HRIG_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRIG_TEST_DATABASE_URL not set; skipping PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse test db config: %v", err)
	}
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	// Isolate query integration tests in a dedicated schema so parallel test packages
	// sharing HRIG_TEST_DATABASE_URL do not collide with public schema resets.
	cfg.ConnConfig.RuntimeParams["search_path"] = "query_test"

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, _ = pool.Exec(cleanCtx, "DROP SCHEMA IF EXISTS query_test CASCADE;")
		cleanCancel()
		pool.Close()
	})

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS query_test CASCADE; CREATE SCHEMA query_test;"); err != nil {
		t.Fatalf("reset query_test schema: %v", err)
	}

	sqlFiles := []string{
		filepath.Join("..", "..", "..", "db", "001_schema.sql"),
		filepath.Join("..", "..", "..", "db", "002_seed.sql"),
		filepath.Join("..", "..", "..", "db", "003_demo_seed.sql"),
		filepath.Join("..", "..", "..", "db", "004_medical_manual.sql"),
		// Re-run 004_medical_manual.sql to verify idempotency on an already-seeded schema.
		filepath.Join("..", "..", "..", "db", "004_medical_manual.sql"),
	}
	for _, path := range sqlFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if _, err := pool.Exec(ctx, string(raw)); err != nil {
			t.Fatalf("exec %s: %v", path, err)
		}
	}

	return pool
}

func TestIntegration_QueryService(t *testing.T) {
	pool := setupTestDB(t)
	epoch := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	svc := query.NewService(pool, epoch)
	ctx := context.Background()

	// 1. ListAstronauts -> 3 active seeded astronauts with valid UUIDs and derived Age
	astronauts, err := svc.ListAstronauts(ctx)
	if err != nil {
		t.Fatalf("ListAstronauts: %v", err)
	}
	if len(astronauts) != 3 {
		t.Fatalf("expected 3 active astronauts, got %d", len(astronauts))
	}
	wantNames := []string{"Ana Torres", "Liam Okafor", "Sofia Petrov"}
	for i, wantName := range wantNames {
		a := astronauts[i]
		if a.FullName != wantName {
			t.Fatalf("astronaut[%d] FullName = %q, want %q", i, a.FullName, wantName)
		}
		var parsedUUID pgtype.UUID
		if err := parsedUUID.Scan(a.ID); err != nil || !parsedUUID.Valid {
			t.Fatalf("astronaut[%d] invalid UUID %q: %v", i, a.ID, err)
		}
		if a.Age <= 0 {
			t.Fatalf("astronaut[%d] expected positive derived Age, got %d", i, a.Age)
		}
		if !a.IsActive {
			t.Fatalf("astronaut[%d] expected IsActive=true", i)
		}
	}
	anaID := astronauts[0].ID

	// 2. ListMetrics -> 23 active metrics with thresholds and unit decimals
	metrics, err := svc.ListMetrics(ctx)
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if len(metrics) != 23 {
		t.Fatalf("expected 23 active metrics, got %d", len(metrics))
	}
	byCode := make(map[string]model.Metric, len(metrics))
	for _, m := range metrics {
		byCode[m.Code] = m
		if m.Threshold == nil {
			t.Fatalf("expected metric %q to have threshold metadata", m.Code)
		}
		if m.ValueType != model.ValueImage && m.BaseUnitDecimals == nil {
			t.Fatalf("expected metric %q to have BaseUnitDecimals", m.Code)
		}
	}
	coreTemp, ok := byCode["core_temp"]
	if !ok {
		t.Fatal("missing core_temp metric")
	}
	if coreTemp.BaseUnitDecimals == nil || *coreTemp.BaseUnitDecimals != 1 {
		t.Fatalf("core_temp BaseUnitDecimals = %v, want 1", coreTemp.BaseUnitDecimals)
	}
	if coreTemp.BaseUnitDimension == nil || *coreTemp.BaseUnitDimension != "temperature" {
		t.Fatalf("core_temp BaseUnitDimension = %v, want temperature", coreTemp.BaseUnitDimension)
	}
	if coreTemp.Threshold.WarningMin == nil || *coreTemp.Threshold.WarningMin != 36.5 {
		t.Fatalf("core_temp WarningMin = %v, want 36.5", coreTemp.Threshold.WarningMin)
	}
	if us, ok := byCode["ultrasound"]; !ok || us.Threshold.ClassificationMode != "manual" || us.BaseUnitDecimals != nil {
		t.Fatalf("unexpected ultrasound metric: %+v", us)
	}

	// 3. ListTrends -> spo2 & cabin_o2 chronological ASC readings
	spo2All, err := svc.ListTrends(ctx, query.TrendsFilter{MetricCode: "spo2"})
	if err != nil {
		t.Fatalf("ListTrends(spo2): %v", err)
	}
	if len(spo2All) != 9 {
		t.Fatalf("expected 9 seeded spo2 readings (3 crew * 3 days), got %d", len(spo2All))
	}
	for i := 1; i < len(spo2All); i++ {
		if spo2All[i].RecordedAtUTC.Before(spo2All[i-1].RecordedAtUTC) {
			t.Fatalf("spo2 trends not in ASC order at index %d", i)
		}
	}

	spo2Ana, err := svc.ListTrends(ctx, query.TrendsFilter{
		MetricCode:  "spo2",
		AstronautID: &anaID,
	})
	if err != nil {
		t.Fatalf("ListTrends(spo2, Ana): %v", err)
	}
	if len(spo2Ana) != 3 {
		t.Fatalf("expected 3 spo2 readings for Ana, got %d", len(spo2Ana))
	}
	if spo2Ana[0].MET != "T+000d 08:00:00" || spo2Ana[2].MET != "T+002d 08:00:00" {
		t.Fatalf("unexpected MET sequence for Ana spo2: %q .. %q", spo2Ana[0].MET, spo2Ana[2].MET)
	}

	spo2Limited, err := svc.ListTrends(ctx, query.TrendsFilter{
		MetricCode:  "spo2",
		AstronautID: &anaID,
		Limit:       2,
	})
	if err != nil {
		t.Fatalf("ListTrends(spo2, Ana, limit=2): %v", err)
	}
	if len(spo2Limited) != 2 || spo2Limited[0].ID != spo2Ana[1].ID || spo2Limited[1].ID != spo2Ana[2].ID {
		t.Fatalf("expected trailing 2 points for limit=2, got %+v", spo2Limited)
	}

	cabinO2, err := svc.ListTrends(ctx, query.TrendsFilter{MetricCode: "cabin_o2"})
	if err != nil {
		t.Fatalf("ListTrends(cabin_o2): %v", err)
	}
	if len(cabinO2) != 3 {
		t.Fatalf("expected 3 seeded cabin_o2 readings, got %d", len(cabinO2))
	}
	for _, pt := range cabinO2 {
		if pt.AstronautID != nil {
			t.Fatalf("expected environmental cabin_o2 AstronautID=nil, got %v", *pt.AstronautID)
		}
		if pt.Environment != model.EnvIVA || pt.Status == model.StatusStale {
			t.Fatalf("unexpected cabin_o2 point: %+v", pt)
		}
	}

	// 4. ListHistory -> pagination (limit/offset), astronaut_id filter, and media asset mapping
	page1, err := svc.ListHistory(ctx, query.HistoryFilter{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListHistory page 1: %v", err)
	}
	page2, err := svc.ListHistory(ctx, query.HistoryFilter{Limit: 5, Offset: 5})
	if err != nil {
		t.Fatalf("ListHistory page 2: %v", err)
	}
	if len(page1) != 5 || len(page2) != 5 {
		t.Fatalf("expected 5 items per page, got %d and %d", len(page1), len(page2))
	}
	seen := make(map[string]struct{}, 10)
	for _, e := range page1 {
		seen[e.ValueID] = struct{}{}
	}
	for _, e := range page2 {
		if _, dup := seen[e.ValueID]; dup {
			t.Fatalf("overlapping ValueID %s across offset=0 and offset=5", e.ValueID)
		}
	}

	anaHistory, err := svc.ListHistory(ctx, query.HistoryFilter{AstronautID: &anaID, Limit: 50})
	if err != nil {
		t.Fatalf("ListHistory Ana: %v", err)
	}
	if len(anaHistory) != 33 {
		t.Fatalf("expected 33 seeded history entries for Ana (11 metrics * 3 days), got %d", len(anaHistory))
	}
	for _, e := range anaHistory {
		if e.AstronautID == nil || *e.AstronautID != anaID {
			t.Fatalf("history entry not filtered to Ana: %+v", e)
		}
	}

	// Insert one measurement with media_asset to verify optional media mapping
	var measID string
	err = pool.QueryRow(ctx, `
		INSERT INTO measurement (astronaut_id, device_id, environment, recorded_at_utc, note)
		VALUES (
			$1::uuid,
			(SELECT id FROM device WHERE code = 'ultrasound_01'),
			'iva',
			'2026-11-16T09:00:00Z',
			'Scheduled ocular ultrasound'
		)
		RETURNING id::text
	`, anaID).Scan(&measID)
	if err != nil {
		t.Fatalf("insert ultrasound measurement: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO measurement_value (measurement_id, metric_id, recorded_at_utc, value_text, status)
		VALUES (
			$1::uuid,
			(SELECT id FROM metric WHERE code = 'ultrasound'),
			'2026-11-16T09:00:00Z',
			'No optic nerve sheath distension',
			'normal'
		)
	`, measID); err != nil {
		t.Fatalf("insert ultrasound value: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO media_asset (measurement_id, kind, file_path, mime_type, findings, captured_at_utc)
		VALUES (
			$1::uuid,
			'ultrasound',
			'media/ultrasound/ana-2026-11-16.png',
			'image/png',
			'Normal posterior segment',
			'2026-11-16T08:59:30Z'
		)
	`, measID); err != nil {
		t.Fatalf("insert media_asset: %v", err)
	}

	latestHistory, err := svc.ListHistory(ctx, query.HistoryFilter{AstronautID: &anaID, Limit: 1})
	if err != nil {
		t.Fatalf("ListHistory latest Ana: %v", err)
	}
	if len(latestHistory) != 1 {
		t.Fatalf("expected 1 latest history entry, got %d", len(latestHistory))
	}
	h0 := latestHistory[0]
	if h0.MetricCode != "ultrasound" || h0.MediaAssetID == nil || h0.MediaKind == nil || *h0.MediaKind != model.MediaUltrasound {
		t.Fatalf("expected ultrasound media entry, got %+v", h0)
	}
	if h0.MediaFilePath == nil || *h0.MediaFilePath != "media/ultrasound/ana-2026-11-16.png" {
		t.Fatalf("unexpected MediaFilePath: %v", h0.MediaFilePath)
	}
	if h0.MediaCapturedAtUTC == nil || !h0.MediaCapturedAtUTC.Equal(time.Date(2026, 11, 16, 8, 59, 30, 0, time.UTC)) {
		t.Fatalf("unexpected MediaCapturedAtUTC: %v", h0.MediaCapturedAtUTC)
	}

	// 5. Medical knowledge: 5 hazards & 20 procedures covering all 23 DB metrics
	hazards, err := svc.ListHazards(ctx, "")
	if err != nil {
		t.Fatalf("ListHazards: %v", err)
	}
	if len(hazards) != 5 {
		t.Fatalf("expected 5 space hazards, got %d", len(hazards))
	}
	wantHazardCodes := map[string]bool{
		"space_radiation":       false,
		"isolation_confinement": false,
		"distance_from_earth":   false,
		"gravity_fields":        false,
		"hostile_closed_env":    false,
	}
	for _, h := range hazards {
		if _, ok := wantHazardCodes[h.Code]; !ok {
			t.Fatalf("unexpected hazard code %q", h.Code)
		}
		wantHazardCodes[h.Code] = true
		for _, mc := range h.RelatedMetricCodes {
			if _, ok := byCode[mc]; !ok {
				t.Fatalf("hazard %q references unknown metric code %q", h.Code, mc)
			}
		}
	}
	for code, seenHazard := range wantHazardCodes {
		if !seenHazard {
			t.Fatalf("missing seeded hazard %q", code)
		}
	}

	procs, err := svc.ListProcedures(ctx, "", "", "")
	if err != nil {
		t.Fatalf("ListProcedures: %v", err)
	}
	if len(procs) != 20 {
		t.Fatalf("expected 20 active medical procedures, got %d", len(procs))
	}
	coveredMetrics := make(map[string]bool, len(byCode))
	for _, p := range procs {
		if !p.Escalation.IsValid() {
			t.Fatalf("procedure %q has invalid escalation %q", p.Code, p.Escalation)
		}
		if len(p.Steps) == 0 || len(p.Warnings) == 0 || len(p.Keywords) == 0 {
			t.Fatalf("procedure %q missing steps/warnings/keywords", p.Code)
		}
		for _, mc := range p.RelatedMetricCodes {
			if _, ok := byCode[mc]; !ok {
				t.Fatalf("procedure %q references unknown metric code %q", p.Code, mc)
			}
			coveredMetrics[mc] = true
		}
		for _, hc := range p.HazardCodes {
			if _, ok := wantHazardCodes[hc]; !ok {
				t.Fatalf("procedure %q references unknown hazard code %q", p.Code, hc)
			}
		}
	}
	for mc := range byCode {
		if !coveredMetrics[mc] {
			t.Fatalf("metric %q is not referenced by any seeded medical procedure", mc)
		}
	}

	// Filtering by category, metric_code, and symptom keyword
	emergProcs, err := svc.ListProcedures(ctx, "", "emergency", "")
	if err != nil {
		t.Fatalf("ListProcedures(category=emergency): %v", err)
	}
	if len(emergProcs) != 3 {
		t.Fatalf("expected 3 emergency procedures, got %d", len(emergProcs))
	}

	spo2Procs, err := svc.ListProcedures(ctx, "", "", "spo2")
	if err != nil {
		t.Fatalf("ListProcedures(metric_code=spo2): %v", err)
	}
	if len(spo2Procs) < 3 {
		t.Fatalf("expected at least 3 procedures linked to spo2, got %d", len(spo2Procs))
	}

	solarProcs, err := svc.ListProcedures(ctx, "solar storm", "", "")
	if err != nil {
		t.Fatalf("ListProcedures(q='solar storm'): %v", err)
	}
	if len(solarProcs) != 1 || solarProcs[0].Code != "acute_radiation_spe" {
		t.Fatalf("expected acute_radiation_spe for 'solar storm', got %+v", solarProcs)
	}

	// 6. Unified Search & HTTP /api/search integration
	if _, err := svc.Search(ctx, "oxygen", "bogus"); !errors.Is(err, query.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest for invalid search tab, got %v", err)
	}

	router := serverhttp.NewRouterWithDeps(config.Load(), serverhttp.RouterDeps{
		Reader: svc,
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=ana&tab=crew", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/search?q=ana&tab=crew expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var crewRes model.SearchResults
	if err := json.Unmarshal(rec.Body.Bytes(), &crewRes); err != nil {
		t.Fatalf("decode crew search: %v", err)
	}
	if len(crewRes.Astronauts) != 1 || crewRes.Astronauts[0].FullName != "Ana Torres" ||
		len(crewRes.Procedures) != 0 || len(crewRes.Hazards) != 0 || len(crewRes.Metrics) != 0 {
		t.Fatalf("unexpected crew search payload: %+v", crewRes)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=oxygen&tab=all", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/search?q=oxygen&tab=all expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var allRes model.SearchResults
	if err := json.Unmarshal(rec.Body.Bytes(), &allRes); err != nil {
		t.Fatalf("decode all search: %v", err)
	}
	if len(allRes.Procedures) == 0 || len(allRes.Metrics) == 0 {
		t.Fatalf("expected oxygen search to match procedures and metrics, got %+v", allRes)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=oxygen&tab=invalid", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/search?tab=invalid expected 400, got %d (%s)", rec.Code, rec.Body.String())
	}
}
