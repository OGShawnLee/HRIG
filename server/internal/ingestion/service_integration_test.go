package ingestion_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"hrig/server/internal/config"
	serverhttp "hrig/server/internal/http"
	"hrig/server/internal/ingestion"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("HRIG_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRIG_TEST_DATABASE_URL not set; skipping PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	// Reset public schema and load 001_schema.sql, 002_seed.sql, 003_demo_seed.sql cleanly.
	if _, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("reset schema: %v", err)
	}

	sqlFiles := []string{
		filepath.Join("..", "..", "..", "db", "001_schema.sql"),
		filepath.Join("..", "..", "..", "db", "002_seed.sql"),
		filepath.Join("..", "..", "..", "db", "003_demo_seed.sql"),
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

func countRows(t *testing.T, pool *pgxpool.Pool) (measurements int, values int) {
	t.Helper()
	ctx := context.Background()
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM measurement").Scan(&measurements); err != nil {
		t.Fatalf("count measurement: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM measurement_value").Scan(&values); err != nil {
		t.Fatalf("count measurement_value: %v", err)
	}
	return measurements, values
}

func postJSON(t *testing.T, router http.Handler, payload any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestIntegration_IngestAllBotSensorShapesAndClassification(t *testing.T) {
	pool := setupTestDB(t)
	router := serverhttp.NewRouter(config.Load(), ingestion.NewService(pool))
	ctx := context.Background()

	ana := "Ana Torres"
	liam := "Liam Okafor"
	sofia := "Sofia Petrov"

	f64 := func(v float64) *float64 { return &v }
	iPtr := func(v int) *int { return &v }
	sPtr := func(v string) *string { return &v }

	baseM, baseV := countRows(t, pool)

	// 1. biomonitor_01 (iva, Ana Torres, 9 metrics including mood scale {score})
	rec := postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "biomonitor_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:00Z",
		AstronautName: &ana,
		Readings: []ingestion.ReadingInput{
			{MetricCode: "spo2", ValueNumeric: f64(98), UnitCode: sPtr("%")},
			{MetricCode: "resp_rate", ValueNumeric: f64(15), UnitCode: sPtr("/min")},
			{MetricCode: "heart_rate", ValueNumeric: f64(72), UnitCode: sPtr("/min")},
			{MetricCode: "hrv_rmssd", ValueNumeric: f64(45), UnitCode: sPtr("ms")},
			{MetricCode: "core_temp", ValueNumeric: f64(36.8), UnitCode: sPtr("Cel")},
			{MetricCode: "sleep_hours", ValueNumeric: f64(8), UnitCode: sPtr("h")},
			{MetricCode: "mood", ValueScale: iPtr(4), UnitCode: sPtr("{score}")},
			{MetricCode: "hydration_intake", ValueNumeric: f64(2200), UnitCode: sPtr("mL")},
			{MetricCode: "body_mass", ValueNumeric: f64(65), UnitCode: sPtr("kg")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("biomonitor_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 2. baumanometro_01 (iva, Liam Okafor)
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "baumanometro_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:01Z",
		AstronautName: &liam,
		Readings: []ingestion.ReadingInput{
			{MetricCode: "bp_systolic", ValueNumeric: f64(115), UnitCode: sPtr("mm[Hg]")},
			{MetricCode: "bp_diastolic", ValueNumeric: f64(75), UnitCode: sPtr("mm[Hg]")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("baumanometro_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 3. dosimeter_01 (iva, Sofia Petrov)
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "dosimeter_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:02Z",
		AstronautName: &sofia,
		Readings: []ingestion.ReadingInput{
			{MetricCode: "radiation_rate", ValueNumeric: f64(0.25), UnitCode: sPtr("uSv/h")},
			{MetricCode: "radiation_dose_cum", ValueNumeric: f64(14.5), UnitCode: sPtr("mSv")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("dosimeter_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 4. cabin_sensor_01 (iva, no astronaut_name)
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "cabin_sensor_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:03Z",
		Readings: []ingestion.ReadingInput{
			{MetricCode: "cabin_o2", ValueNumeric: f64(21.0), UnitCode: sPtr("%")},
			{MetricCode: "cabin_ppco2", ValueNumeric: f64(1.2), UnitCode: sPtr("mm[Hg]")},
			{MetricCode: "cabin_pressure", ValueNumeric: f64(101.3), UnitCode: sPtr("kPa")},
			{MetricCode: "cabin_temp", ValueNumeric: f64(22.5), UnitCode: sPtr("Cel")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("cabin_sensor_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var cabinRes ingestion.IngestResult
	if err := json.Unmarshal(rec.Body.Bytes(), &cabinRes); err != nil {
		t.Fatalf("decode cabin response: %v", err)
	}
	var cabinAstroNull bool
	if err := pool.QueryRow(ctx, "SELECT astronaut_id IS NULL FROM measurement WHERE id = $1", cabinRes.MeasurementID).Scan(&cabinAstroNull); err != nil {
		t.Fatalf("query cabin measurement: %v", err)
	}
	if !cabinAstroNull {
		t.Fatal("expected cabin_sensor_01 measurement to have astronaut_id IS NULL")
	}

	// 5. suit_sensor_01 (eva, no astronaut_name)
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "suit_sensor_01",
		Environment:   "eva",
		RecordedAtUTC: "2026-11-15T12:00:04Z",
		Readings: []ingestion.ReadingInput{
			{MetricCode: "suit_pressure", ValueNumeric: f64(29.0), UnitCode: sPtr("kPa")},
			{MetricCode: "suit_temp", ValueNumeric: f64(26.0), UnitCode: sPtr("Cel")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("suit_sensor_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 6. lab_assay_01 (iva, Ana Torres)
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "lab_assay_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:05Z",
		AstronautName: &ana,
		Readings: []ingestion.ReadingInput{
			{MetricCode: "blood_glucose", ValueNumeric: f64(95), UnitCode: sPtr("mg/dL")},
			{MetricCode: "rbc_count", ValueNumeric: f64(4.8), UnitCode: sPtr("10*6/uL")},
			{MetricCode: "wbc_count", ValueNumeric: f64(6.5), UnitCode: sPtr("10*3/uL")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("lab_assay_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 7. ultrasound_01 (iva, Liam Okafor, value_text without unit_code, manual -> normal)
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "ultrasound_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:06Z",
		AstronautName: &liam,
		Readings: []ingestion.ReadingInput{
			{MetricCode: "ultrasound", ValueText: sPtr("simulated normal ultrasound finding")},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("ultrasound_01 status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var usRes ingestion.IngestResult
	if err := json.Unmarshal(rec.Body.Bytes(), &usRes); err != nil {
		t.Fatalf("decode ultrasound response: %v", err)
	}
	var usStatus string
	var usUnitNull bool
	if err := pool.QueryRow(ctx, "SELECT status::text, unit_id IS NULL FROM measurement_value WHERE measurement_id = $1", usRes.MeasurementID).Scan(&usStatus, &usUnitNull); err != nil {
		t.Fatalf("query ultrasound value: %v", err)
	}
	if usStatus != "normal" || !usUnitNull {
		t.Fatalf("ultrasound status=%q unit_null=%v, want normal and true", usStatus, usUnitNull)
	}

	// 8. Abnormal & Critical threshold classification persistence
	rec = postJSON(t, router, ingestion.IngestRequest{
		DeviceCode:    "biomonitor_01",
		Environment:   "iva",
		RecordedAtUTC: "2026-11-15T12:00:07Z",
		AstronautName: &ana,
		Readings: []ingestion.ReadingInput{
			{MetricCode: "heart_rate", ValueNumeric: f64(150), UnitCode: sPtr("/min")}, // abnormal
			{MetricCode: "spo2", ValueNumeric: f64(89), UnitCode: sPtr("%")},          // critical
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("threshold test status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var threshRes ingestion.IngestResult
	if err := json.Unmarshal(rec.Body.Bytes(), &threshRes); err != nil {
		t.Fatalf("decode threshold response: %v", err)
	}

	var hrStatus, spo2Status string
	if err := pool.QueryRow(ctx, `
		SELECT mv.status::text
		FROM measurement_value mv
		JOIN metric m ON m.id = mv.metric_id
		WHERE mv.measurement_id = $1 AND m.code = 'heart_rate'
	`, threshRes.MeasurementID).Scan(&hrStatus); err != nil {
		t.Fatalf("query heart_rate status: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT mv.status::text
		FROM measurement_value mv
		JOIN metric m ON m.id = mv.metric_id
		WHERE mv.measurement_id = $1 AND m.code = 'spo2'
	`, threshRes.MeasurementID).Scan(&spo2Status); err != nil {
		t.Fatalf("query spo2 status: %v", err)
	}
	if hrStatus != "abnormal" {
		t.Fatalf("heart_rate=150 status = %q, want abnormal", hrStatus)
	}
	if spo2Status != "critical" {
		t.Fatalf("spo2=89 status = %q, want critical", spo2Status)
	}

	afterM, afterV := countRows(t, pool)
	if afterM-baseM != 8 {
		t.Fatalf("expected 8 new measurement rows, got %d", afterM-baseM)
	}
	// 9 + 2 + 2 + 4 + 2 + 3 + 1 + 2 = 25
	if afterV-baseV != 25 {
		t.Fatalf("expected 25 new measurement_value rows, got %d", afterV-baseV)
	}
}

func TestIntegration_AtomicRollbackAndErrorCodes(t *testing.T) {
	pool := setupTestDB(t)
	router := serverhttp.NewRouter(config.Load(), ingestion.NewService(pool))
	ctx := context.Background()

	ana := "Ana Torres"
	unknownAstro := "Nobody Known"
	f64 := func(v float64) *float64 { return &v }
	iPtr := func(v int) *int { return &v }
	sPtr := func(v string) *string { return &v }

	cases := []struct {
		name       string
		req        ingestion.IngestRequest
		wantStatus int
	}{
		{
			name: "unknown_metric_code_rolls_back_parent_and_first_reading",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:00Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
					{MetricCode: "nonexistent_metric", ValueNumeric: f64(1), UnitCode: sPtr("%")},
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "unknown_unit_code_rolls_back_parent_and_first_reading",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:01Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
					{MetricCode: "spo2", ValueNumeric: f64(98), UnitCode: sPtr("bogus_unit")},
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "unit_dimension_mismatch_rolls_back",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:02Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
					{MetricCode: "spo2", ValueNumeric: f64(98), UnitCode: sPtr("kPa")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "metric_type_2_with_astronaut_name_rejected",
			req: ingestion.IngestRequest{
				DeviceCode:    "cabin_sensor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:03Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "cabin_pressure", ValueNumeric: f64(101.3), UnitCode: sPtr("kPa")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "metric_type_1_without_astronaut_name_rejected",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:04Z",
				Readings: []ingestion.ReadingInput{
					{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "unknown_device_code_404",
			req: ingestion.IngestRequest{
				DeviceCode:    "unknown_device",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:05Z",
				Readings: []ingestion.ReadingInput{
					{MetricCode: "cabin_pressure", ValueNumeric: f64(101.3), UnitCode: sPtr("kPa")},
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "unknown_astronaut_404",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:06Z",
				AstronautName: &unknownAstro,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "duplicate_metric_code_409",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:07Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
					{MetricCode: "heart_rate", ValueNumeric: f64(72), UnitCode: sPtr("/min")},
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "scale_out_of_range_400",
			req: ingestion.IngestRequest{
				DeviceCode:    "biomonitor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:08Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "mood", ValueScale: iPtr(9), UnitCode: sPtr("{score}")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ultrasound_with_unit_code_400",
			req: ingestion.IngestRequest{
				DeviceCode:    "ultrasound_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:09Z",
				AstronautName: &ana,
				Readings: []ingestion.ReadingInput{
					{MetricCode: "ultrasound", ValueText: sPtr("finding"), UnitCode: sPtr("%")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "recorded_at_without_timezone_400",
			req: ingestion.IngestRequest{
				DeviceCode:    "cabin_sensor_01",
				Environment:   "iva",
				RecordedAtUTC: "2026-11-15T13:00:10",
				Readings: []ingestion.ReadingInput{
					{MetricCode: "cabin_pressure", ValueNumeric: f64(101.3), UnitCode: sPtr("kPa")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_environment_400",
			req: ingestion.IngestRequest{
				DeviceCode:    "cabin_sensor_01",
				Environment:   "spacewalk",
				RecordedAtUTC: "2026-11-15T13:00:11Z",
				Readings: []ingestion.ReadingInput{
					{MetricCode: "cabin_pressure", ValueNumeric: f64(101.3), UnitCode: sPtr("kPa")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			beforeM, beforeV := countRows(t, pool)
			rec := postJSON(t, router, tc.req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			afterM, afterV := countRows(t, pool)
			if afterM != beforeM || afterV != beforeV {
				t.Fatalf("rollback failed: before (%d, %d), after (%d, %d)", beforeM, beforeV, afterM, afterV)
			}
		})
	}

	t.Run("ambiguous_astronaut_name_409", func(t *testing.T) {
		var dupID string
		err := pool.QueryRow(ctx, `
			INSERT INTO astronaut (full_name, birth_date, nationality_code, is_active)
			VALUES ('Ana Torres', '1990-01-01', 'US', true)
			RETURNING id::text
		`).Scan(&dupID)
		if err != nil {
			t.Fatalf("insert duplicate astronaut: %v", err)
		}
		defer func() {
			_, _ = pool.Exec(ctx, "DELETE FROM astronaut WHERE id = $1", dupID)
		}()

		beforeM, beforeV := countRows(t, pool)
		rec := postJSON(t, router, ingestion.IngestRequest{
			DeviceCode:    "biomonitor_01",
			Environment:   "iva",
			RecordedAtUTC: "2026-11-15T13:00:12Z",
			AstronautName: &ana,
			Readings: []ingestion.ReadingInput{
				{MetricCode: "heart_rate", ValueNumeric: f64(70), UnitCode: sPtr("/min")},
			},
		})
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409 (body: %s)", rec.Code, rec.Body.String())
		}
		afterM, afterV := countRows(t, pool)
		if afterM != beforeM || afterV != beforeV {
			t.Fatalf("rollback failed: before (%d, %d), after (%d, %d)", beforeM, beforeV, afterM, afterV)
		}
	})
}
