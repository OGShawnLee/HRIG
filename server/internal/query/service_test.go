package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"hrig/server/internal/db"
	"hrig/shared/model"
)

type mockQuerier struct {
	listAstronautsFn func(ctx context.Context) ([]db.ListActiveAstronautsRow, error)
	listMetricsFn    func(ctx context.Context) ([]db.ListActiveMetricsRow, error)
	listTrendsFn     func(ctx context.Context, arg db.ListTrendsByMetricParams) ([]db.ListTrendsByMetricRow, error)
	listHistoryFn    func(ctx context.Context, arg db.ListHistoryParams) ([]db.ListHistoryRow, error)
}

func (m *mockQuerier) CreateMeasurement(ctx context.Context, arg db.CreateMeasurementParams) (db.Measurement, error) {
	return db.Measurement{}, nil
}
func (m *mockQuerier) CreateMeasurementValue(ctx context.Context, arg db.CreateMeasurementValueParams) (db.MeasurementValue, error) {
	return db.MeasurementValue{}, nil
}
func (m *mockQuerier) CreateMediaAsset(ctx context.Context, arg db.CreateMediaAssetParams) (db.MediaAsset, error) {
	return db.MediaAsset{}, nil
}
func (m *mockQuerier) GetActiveDeviceByCode(ctx context.Context, code string) (db.Device, error) {
	return db.Device{}, nil
}
func (m *mockQuerier) GetActiveMetricByCode(ctx context.Context, code string) (db.GetActiveMetricByCodeRow, error) {
	return db.GetActiveMetricByCodeRow{}, nil
}
func (m *mockQuerier) GetActiveUnitByCode(ctx context.Context, code string) (db.Unit, error) {
	return db.Unit{}, nil
}
func (m *mockQuerier) GetAstronautByID(ctx context.Context, id pgtype.UUID) (db.GetAstronautByIDRow, error) {
	return db.GetAstronautByIDRow{}, nil
}
func (m *mockQuerier) GetThresholdByMetricCode(ctx context.Context, code string) (db.GetThresholdByMetricCodeRow, error) {
	return db.GetThresholdByMetricCodeRow{}, nil
}
func (m *mockQuerier) ListActiveAstronauts(ctx context.Context) ([]db.ListActiveAstronautsRow, error) {
	if m.listAstronautsFn != nil {
		return m.listAstronautsFn(ctx)
	}
	return nil, nil
}
func (m *mockQuerier) ListActiveAstronautsByFullName(ctx context.Context, fullName string) ([]db.Astronaut, error) {
	return nil, nil
}
func (m *mockQuerier) ListActiveDevices(ctx context.Context) ([]db.Device, error) {
	return nil, nil
}
func (m *mockQuerier) ListActiveMetrics(ctx context.Context) ([]db.ListActiveMetricsRow, error) {
	if m.listMetricsFn != nil {
		return m.listMetricsFn(ctx)
	}
	return nil, nil
}
func (m *mockQuerier) ListHistory(ctx context.Context, arg db.ListHistoryParams) ([]db.ListHistoryRow, error) {
	if m.listHistoryFn != nil {
		return m.listHistoryFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockQuerier) ListLatestReadingsByMetric(ctx context.Context, astronautID pgtype.UUID) ([]db.ListLatestReadingsByMetricRow, error) {
	return nil, nil
}
func (m *mockQuerier) ListTrendsByMetric(ctx context.Context, arg db.ListTrendsByMetricParams) ([]db.ListTrendsByMetricRow, error) {
	if m.listTrendsFn != nil {
		return m.listTrendsFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockQuerier) ListUnits(ctx context.Context) ([]db.Unit, error) {
	return nil, nil
}

func mustUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		t.Fatalf("scan uuid %q: %v", s, err)
	}
	return u
}

func mustNumeric(t *testing.T, s string) pgtype.Numeric {
	t.Helper()
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		t.Fatalf("scan numeric %q: %v", s, err)
	}
	return n
}

func TestServiceListAstronauts(t *testing.T) {
	epoch := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 11, 12, 15, 30, 0, 0, time.UTC)
	birth := time.Date(1988, 4, 12, 0, 0, 0, 0, time.UTC)
	natLabel := "Mexico"

	q := &mockQuerier{
		listAstronautsFn: func(ctx context.Context) ([]db.ListActiveAstronautsRow, error) {
			return []db.ListActiveAstronautsRow{
				{
					ID:               mustUUID(t, "11111111-1111-4111-8111-111111111111"),
					FullName:         "Ana Torres",
					BirthDate:        pgtype.Date{Time: birth, Valid: true},
					Age:              38,
					NationalityCode:  "MX",
					NationalityLabel: &natLabel,
					IsActive:         true,
					CreatedAt:        pgtype.Timestamptz{Time: created, Valid: true},
					UpdatedAt:        pgtype.Timestamptz{Time: updated, Valid: true},
				},
			}, nil
		},
	}

	svc := NewServiceWithQuerier(q, epoch)
	items, err := svc.ListAstronauts(context.Background())
	if err != nil {
		t.Fatalf("ListAstronauts error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 astronaut, got %d", len(items))
	}
	a := items[0]
	if a.ID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("unexpected ID %q", a.ID)
	}
	if a.FullName != "Ana Torres" || a.Age != 38 || a.NationalityCode != "MX" {
		t.Fatalf("unexpected astronaut fields: %+v", a)
	}
	if a.NationalityLabel == nil || *a.NationalityLabel != "Mexico" {
		t.Fatalf("unexpected NationalityLabel: %v", a.NationalityLabel)
	}
	if !a.BirthDate.Equal(birth) || !a.CreatedAt.Equal(created) || !a.UpdatedAt.Equal(updated) {
		t.Fatalf("unexpected timestamps: birth=%s created=%s updated=%s", a.BirthDate, a.CreatedAt, a.UpdatedAt)
	}
}

func TestServiceListMetrics(t *testing.T) {
	epoch := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	unitCode := "Cel"
	unitSym := "°C"
	unitDim := "temperature"
	var decimals int16 = 1
	displaySym := "°C"
	var order13 int16 = 13
	modeNumeric := "numeric"

	q := &mockQuerier{
		listMetricsFn: func(ctx context.Context) ([]db.ListActiveMetricsRow, error) {
			return []db.ListActiveMetricsRow{
				{
					ID:                 mustUUID(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
					Code:               "core_temp",
					Name:               "Core body temperature",
					Category:           "temperature",
					Tier:               db.MetricTierPrimary,
					MetricType:         1,
					ValueType:          db.ValueTypeNumeric,
					BaseUnitID:         mustUUID(t, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
					BaseUnitCode:       &unitCode,
					BaseUnitSymbol:     &unitSym,
					BaseUnitDimension:  &unitDim,
					BaseUnitDecimals:   &decimals,
					DisplayUnitSymbol:  &displaySym,
					DisplayOrder:       &order13,
					IsActive:           true,
					WarningMin:         mustNumeric(t, "36.5"),
					WarningMax:         mustNumeric(t, "37.3"),
					CriticalMin:        mustNumeric(t, "35.0"),
					CriticalMax:        mustNumeric(t, "40.0"),
					ClassificationMode: &modeNumeric,
				},
				{
					ID:         mustUUID(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc"),
					Code:       "ultrasound",
					Name:       "Diagnostic ultrasound",
					Category:   "lab_imaging",
					Tier:       db.MetricTierSecondary,
					MetricType: 1,
					ValueType:  db.ValueTypeImage,
					IsActive:   true,
				},
			}, nil
		},
	}

	svc := NewServiceWithQuerier(q, epoch)
	metrics, err := svc.ListMetrics(context.Background())
	if err != nil {
		t.Fatalf("ListMetrics error: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	m0 := metrics[0]
	if m0.BaseUnitDimension == nil || *m0.BaseUnitDimension != "temperature" {
		t.Fatalf("unexpected BaseUnitDimension: %v", m0.BaseUnitDimension)
	}
	if m0.BaseUnitDecimals == nil || *m0.BaseUnitDecimals != 1 {
		t.Fatalf("unexpected BaseUnitDecimals: %v", m0.BaseUnitDecimals)
	}
	if m0.BaseUnitID == nil || *m0.BaseUnitID != "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" {
		t.Fatalf("unexpected BaseUnitID: %v", m0.BaseUnitID)
	}
	if m0.DisplayOrder != 13 {
		t.Fatalf("expected DisplayOrder 13, got %d", m0.DisplayOrder)
	}
	if m0.Threshold == nil {
		t.Fatal("expected Threshold to be populated")
	}
	if m0.Threshold.WarningMin == nil || *m0.Threshold.WarningMin != 36.5 ||
		m0.Threshold.WarningMax == nil || *m0.Threshold.WarningMax != 37.3 ||
		m0.Threshold.CriticalMin == nil || *m0.Threshold.CriticalMin != 35.0 ||
		m0.Threshold.CriticalMax == nil || *m0.Threshold.CriticalMax != 40.0 ||
		m0.Threshold.ClassificationMode != "numeric" {
		t.Fatalf("unexpected threshold: %+v", m0.Threshold)
	}

	m1 := metrics[1]
	if m1.BaseUnitID != nil || m1.DisplayOrder != 0 || m1.Threshold != nil {
		t.Fatalf("expected nil BaseUnitID, 0 DisplayOrder, nil Threshold; got %+v", m1)
	}
}

func TestServiceListTrends_ValidationAndSlicing(t *testing.T) {
	epoch := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)

	t.Run("validation_errors", func(t *testing.T) {
		svc := NewServiceWithQuerier(&mockQuerier{}, epoch)
		badUUID := "not-a-uuid"

		cases := []struct {
			name   string
			filter TrendsFilter
		}{
			{name: "empty_metric_code", filter: TrendsFilter{MetricCode: "   "}},
			{name: "invalid_astronaut_id", filter: TrendsFilter{MetricCode: "spo2", AstronautID: &badUUID}},
			{name: "negative_limit", filter: TrendsFilter{MetricCode: "spo2", Limit: -1}},
			{name: "limit_above_max", filter: TrendsFilter{MetricCode: "spo2", Limit: 1001}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := svc.ListTrends(context.Background(), tc.filter)
				if !errors.Is(err, ErrBadRequest) {
					t.Fatalf("expected ErrBadRequest, got %v", err)
				}
			})
		}
	})

	t.Run("slices_latest_limit_in_asc_order_and_computes_met", func(t *testing.T) {
		unitCode := "%"
		unitSym := "%"
		astroStr := "11111111-1111-4111-8111-111111111111"
		since := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

		var gotParams db.ListTrendsByMetricParams
		q := &mockQuerier{
			listTrendsFn: func(ctx context.Context, arg db.ListTrendsByMetricParams) ([]db.ListTrendsByMetricRow, error) {
				gotParams = arg
				return []db.ListTrendsByMetricRow{
					{
						ID:            mustUUID(t, "00000000-0000-4000-8000-000000000001"),
						MeasurementID: mustUUID(t, "10000000-0000-4000-8000-000000000001"),
						MetricID:      mustUUID(t, "20000000-0000-4000-8000-000000000001"),
						MetricCode:    "spo2",
						AstronautID:   mustUUID(t, astroStr),
						DeviceID:      mustUUID(t, "30000000-0000-4000-8000-000000000001"),
						Environment:   db.EnvironmentModeIva,
						UnitCode:      &unitCode,
						UnitSymbol:    &unitSym,
						RecordedAtUtc: pgtype.Timestamptz{Time: time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC), Valid: true},
						ValueNumeric:  mustNumeric(t, "98.0"),
						Status:        db.ReadingStatusNormal,
					},
					{
						ID:            mustUUID(t, "00000000-0000-4000-8000-000000000002"),
						MeasurementID: mustUUID(t, "10000000-0000-4000-8000-000000000002"),
						MetricID:      mustUUID(t, "20000000-0000-4000-8000-000000000001"),
						MetricCode:    "spo2",
						AstronautID:   mustUUID(t, astroStr),
						DeviceID:      mustUUID(t, "30000000-0000-4000-8000-000000000001"),
						Environment:   db.EnvironmentModeIva,
						UnitCode:      &unitCode,
						UnitSymbol:    &unitSym,
						RecordedAtUtc: pgtype.Timestamptz{Time: time.Date(2026, 11, 15, 8, 0, 0, 0, time.UTC), Valid: true},
						ValueNumeric:  mustNumeric(t, "93.5"),
						Status:        db.ReadingStatusAbnormal,
					},
					{
						ID:            mustUUID(t, "00000000-0000-4000-8000-000000000003"),
						MeasurementID: mustUUID(t, "10000000-0000-4000-8000-000000000003"),
						MetricID:      mustUUID(t, "20000000-0000-4000-8000-000000000001"),
						MetricCode:    "spo2",
						AstronautID:   mustUUID(t, astroStr),
						DeviceID:      mustUUID(t, "30000000-0000-4000-8000-000000000001"),
						Environment:   db.EnvironmentModeIva,
						UnitCode:      &unitCode,
						UnitSymbol:    &unitSym,
						RecordedAtUtc: pgtype.Timestamptz{Time: time.Date(2026, 11, 15, 10, 30, 5, 0, time.UTC), Valid: true},
						ValueNumeric:  mustNumeric(t, "89.0"),
						Status:        db.ReadingStatusCritical,
					},
				}, nil
			},
		}

		svc := NewServiceWithQuerier(q, epoch)
		points, err := svc.ListTrends(context.Background(), TrendsFilter{
			MetricCode:  " spo2 ",
			AstronautID: &astroStr,
			SinceUTC:    &since,
			Limit:       2,
		})
		if err != nil {
			t.Fatalf("ListTrends error: %v", err)
		}
		if gotParams.MetricCode != "spo2" || !gotParams.AstronautID.Valid || !gotParams.SinceUtc.Valid {
			t.Fatalf("unexpected params passed to querier: %+v", gotParams)
		}
		if len(points) != 2 {
			t.Fatalf("expected 2 points after limit slicing, got %d", len(points))
		}
		if points[0].ID != "00000000-0000-4000-8000-000000000002" || points[1].ID != "00000000-0000-4000-8000-000000000003" {
			t.Fatalf("expected trailing 2 points in ASC order, got %+v", points)
		}
		if points[0].ValueNumeric == nil || *points[0].ValueNumeric != 93.5 || points[0].Status != model.StatusAbnormal {
			t.Fatalf("unexpected first point: %+v", points[0])
		}
		if points[1].MET != "T+002d 10:30:05" {
			t.Fatalf("expected MET T+002d 10:30:05, got %q", points[1].MET)
		}
	})
}

func TestServiceListHistory_ValidationAndMediaMapping(t *testing.T) {
	epoch := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)

	t.Run("validation_errors", func(t *testing.T) {
		svc := NewServiceWithQuerier(&mockQuerier{}, epoch)
		badUUID := "bad-uuid"

		cases := []struct {
			name   string
			filter HistoryFilter
		}{
			{name: "invalid_astronaut_id", filter: HistoryFilter{AstronautID: &badUUID}},
			{name: "negative_limit", filter: HistoryFilter{Limit: -1}},
			{name: "limit_above_max", filter: HistoryFilter{Limit: 501}},
			{name: "negative_offset", filter: HistoryFilter{Offset: -1}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := svc.ListHistory(context.Background(), tc.filter)
				if !errors.Is(err, ErrBadRequest) {
					t.Fatalf("expected ErrBadRequest, got %v", err)
				}
			})
		}
	})

	t.Run("pagination_defaults_and_media_mapping", func(t *testing.T) {
		devCode := "ultrasound_01"
		devName := "Diagnostic ultrasound"
		astroName := "Liam Okafor"
		mediaPath := "media/ultrasound/scan.png"
		mediaMime := "image/png"
		mediaFindings := "Clear optic nerve sheath"
		capturedAt := time.Date(2026, 11, 15, 9, 15, 0, 0, time.UTC)
		recordedAt := time.Date(2026, 11, 15, 9, 16, 0, 0, time.UTC)

		var gotParams db.ListHistoryParams
		q := &mockQuerier{
			listHistoryFn: func(ctx context.Context, arg db.ListHistoryParams) ([]db.ListHistoryRow, error) {
				gotParams = arg
				return []db.ListHistoryRow{
					{
						MeasurementID:      mustUUID(t, "11111111-1111-4111-8111-111111111111"),
						AstronautID:        mustUUID(t, "22222222-2222-4222-8222-222222222222"),
						AstronautName:      &astroName,
						DeviceID:           mustUUID(t, "33333333-3333-4333-8333-333333333333"),
						DeviceCode:         &devCode,
						DeviceName:         &devName,
						Environment:        db.EnvironmentModeIva,
						RecordedAtUtc:      pgtype.Timestamptz{Time: recordedAt, Valid: true},
						ValueID:            mustUUID(t, "44444444-4444-4444-8444-444444444444"),
						MetricID:           mustUUID(t, "55555555-5555-4555-8555-555555555555"),
						MetricCode:         "ultrasound",
						MetricName:         "Diagnostic ultrasound",
						ValueType:          db.ValueTypeImage,
						ValueText:          &mediaFindings,
						Status:             db.ReadingStatusNormal,
						MediaAssetID:       mustUUID(t, "66666666-6666-4666-8666-666666666666"),
						MediaKind:          db.NullMediaKind{MediaKind: db.MediaKindUltrasound, Valid: true},
						MediaFilePath:      &mediaPath,
						MediaMimeType:      &mediaMime,
						MediaFindings:      &mediaFindings,
						MediaCapturedAtUtc: pgtype.Timestamptz{Time: capturedAt, Valid: true},
					},
					{
						MeasurementID: mustUUID(t, "77777777-7777-4777-8777-777777777777"),
						DeviceID:      mustUUID(t, "88888888-8888-4888-8888-888888888888"),
						Environment:   db.EnvironmentModeEva,
						RecordedAtUtc: pgtype.Timestamptz{Time: recordedAt, Valid: true},
						ValueID:       mustUUID(t, "99999999-9999-4999-8999-999999999999"),
						MetricID:      mustUUID(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
						MetricCode:    "suit_pressure",
						MetricName:    "Suit pressure",
						ValueType:     db.ValueTypeNumeric,
						ValueNumeric:  mustNumeric(t, "29.2"),
						Status:        db.ReadingStatusNormal,
					},
				}, nil
			},
		}

		svc := NewServiceWithQuerier(q, epoch)
		entries, err := svc.ListHistory(context.Background(), HistoryFilter{Offset: 10})
		if err != nil {
			t.Fatalf("ListHistory error: %v", err)
		}
		if gotParams.LimitRows != 50 || gotParams.OffsetRows != 10 {
			t.Fatalf("expected default limit 50 and offset 10, got limit=%d offset=%d", gotParams.LimitRows, gotParams.OffsetRows)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		e0 := entries[0]
		if e0.MediaAssetID == nil || *e0.MediaAssetID != "66666666-6666-4666-8666-666666666666" {
			t.Fatalf("unexpected MediaAssetID: %v", e0.MediaAssetID)
		}
		if e0.MediaKind == nil || *e0.MediaKind != model.MediaUltrasound {
			t.Fatalf("unexpected MediaKind: %v", e0.MediaKind)
		}
		if e0.MediaFilePath == nil || *e0.MediaFilePath != mediaPath {
			t.Fatalf("unexpected MediaFilePath: %v", e0.MediaFilePath)
		}
		if e0.MediaCapturedAtUTC == nil || !e0.MediaCapturedAtUTC.Equal(capturedAt) {
			t.Fatalf("unexpected MediaCapturedAtUTC: %v", e0.MediaCapturedAtUTC)
		}
		if e0.MET != "T+002d 09:16:00" {
			t.Fatalf("unexpected MET: %q", e0.MET)
		}

		e1 := entries[1]
		if e1.MediaAssetID != nil || e1.MediaKind != nil || e1.MediaFilePath != nil || e1.MediaCapturedAtUTC != nil {
			t.Fatalf("expected nil media fields when MediaAssetID invalid, got %+v", e1)
		}
		if e1.ValueNumeric == nil || *e1.ValueNumeric != 29.2 {
			t.Fatalf("unexpected ValueNumeric: %v", e1.ValueNumeric)
		}
	})
}
