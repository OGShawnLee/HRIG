package ingestion

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"hrig/server/internal/db"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Ingest(ctx context.Context, req IngestRequest) (IngestResult, error) {
	if s.pool == nil {
		return IngestResult{}, internalError("database pool not configured")
	}

	deviceCode := strings.TrimSpace(req.DeviceCode)
	if deviceCode == "" {
		return IngestResult{}, badRequest("device_code is required")
	}

	env := strings.TrimSpace(req.Environment)
	if env != string(db.EnvironmentModeIva) && env != string(db.EnvironmentModeEva) {
		return IngestResult{}, badRequest("environment must be 'iva' or 'eva'")
	}

	recordedAtStr := strings.TrimSpace(req.RecordedAtUTC)
	if recordedAtStr == "" {
		return IngestResult{}, badRequest("recorded_at_utc is required")
	}
	recordedAt, err := time.Parse(time.RFC3339Nano, recordedAtStr)
	if err != nil {
		return IngestResult{}, badRequest("recorded_at_utc must be RFC3339 with timezone")
	}

	var astronautName *string
	if req.AstronautName != nil {
		trimmed := strings.TrimSpace(*req.AstronautName)
		if trimmed == "" {
			return IngestResult{}, badRequest("astronaut_name must not be empty when provided")
		}
		astronautName = &trimmed
	}

	if len(req.Readings) == 0 {
		return IngestResult{}, badRequest("readings must contain at least one reading")
	}

	seenMetrics := make(map[string]struct{}, len(req.Readings))
	for _, r := range req.Readings {
		mc := strings.TrimSpace(r.MetricCode)
		if mc == "" {
			return IngestResult{}, badRequest("reading metric_code is required")
		}
		if _, exists := seenMetrics[mc]; exists {
			return IngestResult{}, conflict("duplicate metric_code in readings")
		}
		seenMetrics[mc] = struct{}{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return IngestResult{}, internalError("failed to start database transaction")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := db.New(tx)

	device, err := q.GetActiveDeviceByCode(ctx, deviceCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return IngestResult{}, notFound("device not found or inactive")
		}
		return IngestResult{}, mapPgError(err)
	}

	var astronautID pgtype.UUID
	if astronautName != nil {
		astronauts, err := q.ListActiveAstronautsByFullName(ctx, *astronautName)
		if err != nil {
			return IngestResult{}, mapPgError(err)
		}
		if len(astronauts) == 0 {
			return IngestResult{}, notFound("astronaut not found or inactive")
		}
		if len(astronauts) > 1 {
			return IngestResult{}, conflict("multiple active astronauts match astronaut_name")
		}
		astronautID = astronauts[0].ID
	}

	meas, err := q.CreateMeasurement(ctx, db.CreateMeasurementParams{
		AstronautID:   astronautID,
		DeviceID:      device.ID,
		Environment:   db.EnvironmentMode(env),
		RecordedAtUtc: pgtype.Timestamptz{Time: recordedAt.UTC(), Valid: true},
		Note:          nil,
	})
	if err != nil {
		return IngestResult{}, mapPgError(err)
	}

	for _, r := range req.Readings {
		metricCode := strings.TrimSpace(r.MetricCode)
		metric, err := q.GetActiveMetricByCode(ctx, metricCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return IngestResult{}, notFound("metric not found or inactive")
			}
			return IngestResult{}, mapPgError(err)
		}

		if metric.MetricType == 2 && astronautID.Valid {
			return IngestResult{}, badRequest("environmental metric cannot be associated with an astronaut")
		}
		if metric.MetricType == 1 && !astronautID.Valid {
			return IngestResult{}, badRequest("personal metric requires astronaut_name")
		}

		if countProvidedValues(r) != 1 {
			return IngestResult{}, badRequest("reading must provide exactly one value field")
		}

		var (
			valNumeric  pgtype.Numeric
			valScale    *int16
			valText     *string
			unitID      pgtype.UUID
			classifyVal float64
		)

		switch metric.ValueType {
		case db.ValueTypeNumeric:
			if r.ValueNumeric == nil {
				return IngestResult{}, badRequest("numeric metric requires value_numeric")
			}
			num, err := floatToNumeric12_4(*r.ValueNumeric)
			if err != nil {
				return IngestResult{}, err
			}
			valNumeric = num
			classifyVal = *r.ValueNumeric

			unit, err := lookupAndValidateUnit(ctx, q, r.UnitCode, metric.BaseUnitDimension)
			if err != nil {
				return IngestResult{}, err
			}
			unitID = unit.ID

		case db.ValueTypeScale:
			if r.ValueScale == nil {
				return IngestResult{}, badRequest("scale metric requires value_scale")
			}
			if metric.ScaleMin == nil || metric.ScaleMax == nil {
				return IngestResult{}, internalError("scale metric is missing scale bounds")
			}
			sv := *r.ValueScale
			if sv < int(*metric.ScaleMin) || sv > int(*metric.ScaleMax) {
				return IngestResult{}, badRequest("value_scale is outside metric scale range")
			}
			s16 := int16(sv)
			valScale = &s16
			classifyVal = float64(sv)

			unit, err := lookupAndValidateUnit(ctx, q, r.UnitCode, metric.BaseUnitDimension)
			if err != nil {
				return IngestResult{}, err
			}
			unitID = unit.ID

		case db.ValueTypeText, db.ValueTypeImage:
			if r.ValueText == nil || strings.TrimSpace(*r.ValueText) == "" {
				return IngestResult{}, badRequest("text or image metric requires non-empty value_text")
			}
			if r.UnitCode != nil && strings.TrimSpace(*r.UnitCode) != "" {
				return IngestResult{}, badRequest("text or image metric must not specify unit_code")
			}
			txt := *r.ValueText
			valText = &txt

		default:
			return IngestResult{}, internalError("unsupported metric value_type")
		}

		mode := "manual"
		if metric.ClassificationMode != nil {
			mode = *metric.ClassificationMode
		}
		warnMin, err := numericToFloatPtr(metric.WarningMin)
		if err != nil {
			return IngestResult{}, internalError("invalid metric threshold configuration")
		}
		warnMax, err := numericToFloatPtr(metric.WarningMax)
		if err != nil {
			return IngestResult{}, internalError("invalid metric threshold configuration")
		}
		critMin, err := numericToFloatPtr(metric.CriticalMin)
		if err != nil {
			return IngestResult{}, internalError("invalid metric threshold configuration")
		}
		critMax, err := numericToFloatPtr(metric.CriticalMax)
		if err != nil {
			return IngestResult{}, internalError("invalid metric threshold configuration")
		}

		status := ClassifyReading(classifyVal, mode, warnMin, warnMax, critMin, critMax)

		_, err = q.CreateMeasurementValue(ctx, db.CreateMeasurementValueParams{
			MeasurementID: meas.ID,
			MetricID:      metric.ID,
			UnitID:        unitID,
			RecordedAtUtc: meas.RecordedAtUtc,
			ValueNumeric:  valNumeric,
			ValueScale:    valScale,
			ValueText:     valText,
			Note:          nil,
			Status: db.NullReadingStatus{
				ReadingStatus: status,
				Valid:         true,
			},
		})
		if err != nil {
			return IngestResult{}, mapPgError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return IngestResult{}, mapPgError(err)
	}

	return IngestResult{
		MeasurementID: formatUUID(meas.ID),
		RecordedAtUTC: meas.RecordedAtUtc.Time.UTC(),
		ValuesCount:   len(req.Readings),
	}, nil
}

func countProvidedValues(r ReadingInput) int {
	count := 0
	if r.ValueNumeric != nil {
		count++
	}
	if r.ValueScale != nil {
		count++
	}
	if r.ValueText != nil {
		count++
	}
	return count
}

func lookupAndValidateUnit(ctx context.Context, q *db.Queries, rawUnitCode *string, expectedDim *string) (db.Unit, error) {
	if rawUnitCode == nil || strings.TrimSpace(*rawUnitCode) == "" {
		return db.Unit{}, badRequest("unit_code is required for numeric and scale metrics")
	}
	unitCode := strings.TrimSpace(*rawUnitCode)
	unit, err := q.GetActiveUnitByCode(ctx, unitCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Unit{}, notFound("unit not found or inactive")
		}
		return db.Unit{}, mapPgError(err)
	}
	if expectedDim == nil || unit.Dimension != *expectedDim {
		return db.Unit{}, badRequest("unit dimension does not match metric dimension")
	}
	return unit, nil
}

func floatToNumeric12_4(v float64) (pgtype.Numeric, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return pgtype.Numeric{}, badRequest("value_numeric must be finite")
	}
	if v == 0 {
		v = 0
	}
	rounded := math.Round(v*10000) / 10000
	if math.Abs(rounded) >= 1e8 {
		return pgtype.Numeric{}, badRequest("value_numeric exceeds numeric(12,4) precision")
	}
	var num pgtype.Numeric
	if err := num.Scan(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
		return pgtype.Numeric{}, badRequest("value_numeric is invalid")
	}
	return num, nil
}

func numericToFloatPtr(n pgtype.Numeric) (*float64, error) {
	if !n.Valid {
		return nil, nil
	}
	f, err := n.Float64Value()
	if err != nil {
		return nil, err
	}
	if !f.Valid {
		return nil, nil
	}
	val := f.Float64
	return &val, nil
}

func mapPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return conflict("conflicting measurement value")
		case "23514", "23503", "22003":
			return badRequest("measurement violates database constraints")
		}
	}
	return internalError("internal database error")
}

func formatUUID(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
