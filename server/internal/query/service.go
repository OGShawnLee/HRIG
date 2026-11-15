package query

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"hrig/server/internal/db"
	"hrig/server/internal/realtime"
	"hrig/shared/model"
)

const (
	defaultTrendsLimit  = 200
	maxTrendsLimit      = 1000
	defaultHistoryLimit = 50
	maxHistoryLimit     = 500
)

var ErrBadRequest = errors.New("bad request")

type TrendsFilter struct {
	MetricCode  string
	AstronautID *string
	SinceUTC    *time.Time
	Limit       int
}

type HistoryFilter struct {
	AstronautID *string
	Limit       int
	Offset      int
}

type Service struct {
	querier         db.Querier
	missionEpochUTC time.Time
}

func NewService(pool *pgxpool.Pool, missionEpochUTC time.Time) *Service {
	if pool == nil {
		return NewServiceWithQuerier(nil, missionEpochUTC)
	}
	return NewServiceWithQuerier(db.New(pool), missionEpochUTC)
}

func NewServiceWithQuerier(q db.Querier, missionEpochUTC time.Time) *Service {
	return &Service{
		querier:         q,
		missionEpochUTC: missionEpochUTC.UTC(),
	}
}

func (s *Service) ListAstronauts(ctx context.Context) ([]model.Astronaut, error) {
	if s.querier == nil {
		return nil, errors.New("query service is not configured")
	}

	rows, err := s.querier.ListActiveAstronauts(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]model.Astronaut, 0, len(rows))
	for _, row := range rows {
		var birthDate, createdAt, updatedAt time.Time
		if row.BirthDate.Valid {
			birthDate = row.BirthDate.Time.UTC()
		}
		if row.CreatedAt.Valid {
			createdAt = row.CreatedAt.Time.UTC()
		}
		if row.UpdatedAt.Valid {
			updatedAt = row.UpdatedAt.Time.UTC()
		}

		out = append(out, model.Astronaut{
			ID:               formatUUID(row.ID),
			FullName:         row.FullName,
			BirthDate:        birthDate,
			Age:              row.Age,
			NationalityCode:  row.NationalityCode,
			NationalityLabel: cloneStringPtr(row.NationalityLabel),
			IsActive:         row.IsActive,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		})
	}
	return out, nil
}

func (s *Service) ListMetrics(ctx context.Context) ([]model.Metric, error) {
	if s.querier == nil {
		return nil, errors.New("query service is not configured")
	}

	rows, err := s.querier.ListActiveMetrics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]model.Metric, 0, len(rows))
	for _, row := range rows {
		metricID := formatUUID(row.ID)
		var displayOrder int16
		if row.DisplayOrder != nil {
			displayOrder = *row.DisplayOrder
		}

		var threshold *model.MetricThreshold
		if row.ClassificationMode != nil ||
			row.WarningMin.Valid ||
			row.WarningMax.Valid ||
			row.CriticalMin.Valid ||
			row.CriticalMax.Valid {
			warnMin, err := numericToFloatPtr(row.WarningMin)
			if err != nil {
				return nil, fmt.Errorf("convert warning_min for %s: %w", row.Code, err)
			}
			warnMax, err := numericToFloatPtr(row.WarningMax)
			if err != nil {
				return nil, fmt.Errorf("convert warning_max for %s: %w", row.Code, err)
			}
			critMin, err := numericToFloatPtr(row.CriticalMin)
			if err != nil {
				return nil, fmt.Errorf("convert critical_min for %s: %w", row.Code, err)
			}
			critMax, err := numericToFloatPtr(row.CriticalMax)
			if err != nil {
				return nil, fmt.Errorf("convert critical_max for %s: %w", row.Code, err)
			}

			mode := ""
			if row.ClassificationMode != nil {
				mode = *row.ClassificationMode
			}

			threshold = &model.MetricThreshold{
				MetricID:           metricID,
				MetricCode:         row.Code,
				WarningMin:         warnMin,
				WarningMax:         warnMax,
				CriticalMin:        critMin,
				CriticalMax:        critMax,
				ClassificationMode: mode,
			}
		}

		out = append(out, model.Metric{
			ID:                metricID,
			Code:              row.Code,
			Name:              row.Name,
			ShortName:         cloneStringPtr(row.ShortName),
			Category:          row.Category,
			Tier:              model.MetricTier(row.Tier),
			MetricType:        row.MetricType,
			ValueType:         model.ValueType(row.ValueType),
			BaseUnitID:        uuidPtr(row.BaseUnitID),
			BaseUnitCode:      cloneStringPtr(row.BaseUnitCode),
			BaseUnitSymbol:    cloneStringPtr(row.BaseUnitSymbol),
			BaseUnitDimension: cloneStringPtr(row.BaseUnitDimension),
			BaseUnitDecimals:  cloneInt16Ptr(row.BaseUnitDecimals),
			DisplayUnitSymbol: cloneStringPtr(row.DisplayUnitSymbol),
			ScaleMin:          cloneInt16Ptr(row.ScaleMin),
			ScaleMax:          cloneInt16Ptr(row.ScaleMax),
			DisplayOrder:      displayOrder,
			IsActive:          row.IsActive,
			Threshold:         threshold,
		})
	}
	return out, nil
}

func (s *Service) ListTrends(ctx context.Context, filter TrendsFilter) ([]model.TrendPoint, error) {
	if s.querier == nil {
		return nil, errors.New("query service is not configured")
	}

	metricCode := strings.TrimSpace(filter.MetricCode)
	if metricCode == "" {
		return nil, fmt.Errorf("%w: metric_code is required", ErrBadRequest)
	}

	astroUUID, err := parseOptionalUUID(filter.AstronautID)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit == 0 {
		limit = defaultTrendsLimit
	} else if limit < 0 || limit > maxTrendsLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", ErrBadRequest, maxTrendsLimit)
	}

	var since pgtype.Timestamptz
	if filter.SinceUTC != nil && !filter.SinceUTC.IsZero() {
		since = pgtype.Timestamptz{Time: filter.SinceUTC.UTC(), Valid: true}
	}

	rows, err := s.querier.ListTrendsByMetric(ctx, db.ListTrendsByMetricParams{
		MetricCode:  metricCode,
		AstronautID: astroUUID,
		SinceUtc:    since,
	})
	if err != nil {
		return nil, err
	}

	if len(rows) > limit {
		rows = rows[len(rows)-limit:]
	}

	out := make([]model.TrendPoint, 0, len(rows))
	for _, row := range rows {
		valNumeric, err := numericToFloatPtr(row.ValueNumeric)
		if err != nil {
			return nil, fmt.Errorf("convert value_numeric for trend %s: %w", row.MetricCode, err)
		}

		var recordedAtUTC time.Time
		if row.RecordedAtUtc.Valid {
			recordedAtUTC = row.RecordedAtUtc.Time.UTC()
		}

		out = append(out, model.TrendPoint{
			ID:            formatUUID(row.ID),
			MeasurementID: formatUUID(row.MeasurementID),
			MetricID:      formatUUID(row.MetricID),
			MetricCode:    row.MetricCode,
			AstronautID:   uuidPtr(row.AstronautID),
			DeviceID:      formatUUID(row.DeviceID),
			Environment:   model.EnvironmentMode(row.Environment),
			UnitCode:      cloneStringPtr(row.UnitCode),
			UnitSymbol:    cloneStringPtr(row.UnitSymbol),
			ValueNumeric:  valNumeric,
			ValueScale:    cloneInt16Ptr(row.ValueScale),
			ValueText:     cloneStringPtr(row.ValueText),
			Status:        model.ReadingStatus(row.Status),
			RecordedAtUTC: recordedAtUTC,
			MET:           realtime.FormatMET(recordedAtUTC, s.missionEpochUTC),
		})
	}
	return out, nil
}

func (s *Service) ListHistory(ctx context.Context, filter HistoryFilter) ([]model.HistoryEntry, error) {
	if s.querier == nil {
		return nil, errors.New("query service is not configured")
	}

	astroUUID, err := parseOptionalUUID(filter.AstronautID)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit == 0 {
		limit = defaultHistoryLimit
	} else if limit < 0 || limit > maxHistoryLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", ErrBadRequest, maxHistoryLimit)
	}

	if filter.Offset < 0 {
		return nil, fmt.Errorf("%w: offset must be non-negative", ErrBadRequest)
	}

	rows, err := s.querier.ListHistory(ctx, db.ListHistoryParams{
		AstronautID: astroUUID,
		OffsetRows:  int32(filter.Offset),
		LimitRows:   int32(limit),
	})
	if err != nil {
		return nil, err
	}

	out := make([]model.HistoryEntry, 0, len(rows))
	for _, row := range rows {
		valNumeric, err := numericToFloatPtr(row.ValueNumeric)
		if err != nil {
			return nil, fmt.Errorf("convert value_numeric for history %s: %w", row.MetricCode, err)
		}

		var recordedAtUTC time.Time
		if row.RecordedAtUtc.Valid {
			recordedAtUTC = row.RecordedAtUtc.Time.UTC()
		}

		deviceCode := ""
		if row.DeviceCode != nil {
			deviceCode = *row.DeviceCode
		}
		deviceName := ""
		if row.DeviceName != nil {
			deviceName = *row.DeviceName
		}

		entry := model.HistoryEntry{
			MeasurementID:     formatUUID(row.MeasurementID),
			ValueID:           formatUUID(row.ValueID),
			AstronautID:       uuidPtr(row.AstronautID),
			AstronautName:     cloneStringPtr(row.AstronautName),
			DeviceCode:        deviceCode,
			DeviceName:        deviceName,
			Environment:       model.EnvironmentMode(row.Environment),
			MetricCode:        row.MetricCode,
			MetricName:        row.MetricName,
			ValueType:         model.ValueType(row.ValueType),
			UnitCode:          cloneStringPtr(row.UnitCode),
			UnitSymbol:        cloneStringPtr(row.UnitSymbol),
			DisplayUnitSymbol: cloneStringPtr(row.DisplayUnitSymbol),
			ValueNumeric:      valNumeric,
			ValueScale:        cloneInt16Ptr(row.ValueScale),
			ValueText:         cloneStringPtr(row.ValueText),
			Status:            model.ReadingStatus(row.Status),
			RecordedAtUTC:     recordedAtUTC,
			MET:               realtime.FormatMET(recordedAtUTC, s.missionEpochUTC),
			MeasurementNote:   cloneStringPtr(row.MeasurementNote),
			ValueNote:         cloneStringPtr(row.ValueNote),
		}

		if row.MediaAssetID.Valid {
			mediaID := formatUUID(row.MediaAssetID)
			entry.MediaAssetID = &mediaID
			if row.MediaKind.Valid {
				mk := model.MediaKind(row.MediaKind.MediaKind)
				entry.MediaKind = &mk
			}
			entry.MediaFilePath = cloneStringPtr(row.MediaFilePath)
			entry.MediaMimeType = cloneStringPtr(row.MediaMimeType)
			entry.MediaFindings = cloneStringPtr(row.MediaFindings)
			if row.MediaCapturedAtUtc.Valid {
				capturedAt := row.MediaCapturedAtUtc.Time.UTC()
				entry.MediaCapturedAtUTC = &capturedAt
			}
		}

		out = append(out, entry)
	}
	return out, nil
}

func (s *Service) ListHazards(ctx context.Context, query string) ([]model.SpaceHazard, error) {
	if s.querier == nil {
		return nil, errors.New("query service is not configured")
	}

	q := strings.TrimSpace(query)
	var (
		rows []db.SpaceHazard
		err  error
	)
	if q == "" {
		rows, err = s.querier.ListSpaceHazards(ctx)
	} else {
		rows, err = s.querier.SearchSpaceHazards(ctx, q)
	}
	if err != nil {
		return nil, err
	}

	out := make([]model.SpaceHazard, 0, len(rows))
	for _, row := range rows {
		out = append(out, model.SpaceHazard{
			ID:                 formatUUID(row.ID),
			Code:               row.Code,
			Name:               row.Name,
			Summary:            row.Summary,
			ClinicalImpact:     row.ClinicalImpact,
			RelatedMetricCodes: normalizeStringSlice(row.RelatedMetricCodes),
			Source:             row.Source,
			DisplayOrder:       row.DisplayOrder,
		})
	}
	return out, nil
}

func (s *Service) ListProcedures(ctx context.Context, query string, category string, metricCode string) ([]model.MedicalProcedure, error) {
	if s.querier == nil {
		return nil, errors.New("query service is not configured")
	}

	q := strings.TrimSpace(query)
	cat := strings.TrimSpace(category)
	mc := strings.TrimSpace(metricCode)

	var (
		rows []db.MedicalProcedure
		err  error
	)
	if q == "" && cat == "" && mc == "" {
		rows, err = s.querier.ListMedicalProcedures(ctx)
	} else {
		rows, err = s.querier.SearchMedicalProcedures(ctx, db.SearchMedicalProceduresParams{
			Q:          q,
			Category:   cat,
			MetricCode: mc,
		})
	}
	if err != nil {
		return nil, err
	}

	out := make([]model.MedicalProcedure, 0, len(rows))
	for _, row := range rows {
		var env *model.EnvironmentMode
		if row.Environment.Valid {
			e := model.EnvironmentMode(row.Environment.EnvironmentMode)
			env = &e
		}
		out = append(out, model.MedicalProcedure{
			ID:                 formatUUID(row.ID),
			Code:               row.Code,
			Title:              row.Title,
			Category:           row.Category,
			BodySystem:         row.BodySystem,
			Summary:            row.Summary,
			WhenToUse:          row.WhenToUse,
			Escalation:         model.EscalationLevel(row.Escalation),
			EstimatedMinutes:   row.EstimatedMinutes,
			Environment:        env,
			Equipment:          normalizeStringSlice(row.Equipment),
			Steps:              normalizeStringSlice(row.Steps),
			Warnings:           normalizeStringSlice(row.Warnings),
			RelatedMetricCodes: normalizeStringSlice(row.RelatedMetricCodes),
			HazardCodes:        normalizeStringSlice(row.HazardCodes),
			Keywords:           normalizeStringSlice(row.Keywords),
			Source:             row.Source,
			DisplayOrder:       row.DisplayOrder,
		})
	}
	return out, nil
}

func (s *Service) Search(ctx context.Context, q string, tab string) (model.SearchResults, error) {
	if s.querier == nil {
		return model.SearchResults{}, errors.New("query service is not configured")
	}

	trimmedQ := strings.TrimSpace(q)
	trimmedTab := strings.TrimSpace(tab)
	if trimmedTab == "" {
		trimmedTab = "all"
	}

	switch trimmedTab {
	case "all", "crew", "procedures", "hazards_metrics":
	default:
		return model.SearchResults{}, fmt.Errorf("%w: invalid tab %q (expected all, crew, procedures, or hazards_metrics)", ErrBadRequest, trimmedTab)
	}

	res := model.SearchResults{
		Query:      trimmedQ,
		Tab:        trimmedTab,
		Astronauts: []model.Astronaut{},
		Procedures: []model.MedicalProcedure{},
		Hazards:    []model.SpaceHazard{},
		Metrics:    []model.Metric{},
	}

	if trimmedTab == "all" || trimmedTab == "crew" {
		astronauts, err := s.ListAstronauts(ctx)
		if err != nil {
			return model.SearchResults{}, err
		}
		if trimmedQ == "" {
			res.Astronauts = astronauts
		} else {
			needle := strings.ToLower(trimmedQ)
			for _, a := range astronauts {
				if matchesAstronautQuery(a, needle) {
					res.Astronauts = append(res.Astronauts, a)
				}
			}
		}
	}

	if trimmedTab == "all" || trimmedTab == "procedures" {
		procs, err := s.ListProcedures(ctx, trimmedQ, "", "")
		if err != nil {
			return model.SearchResults{}, err
		}
		res.Procedures = procs
	}

	if trimmedTab == "all" || trimmedTab == "hazards_metrics" {
		hazards, err := s.ListHazards(ctx, trimmedQ)
		if err != nil {
			return model.SearchResults{}, err
		}
		res.Hazards = hazards

		metrics, err := s.ListMetrics(ctx)
		if err != nil {
			return model.SearchResults{}, err
		}
		if trimmedQ == "" {
			res.Metrics = metrics
		} else {
			needle := strings.ToLower(trimmedQ)
			for _, m := range metrics {
				if matchesMetricQuery(m, needle) {
					res.Metrics = append(res.Metrics, m)
				}
			}
		}
	}

	return res, nil
}

func matchesAstronautQuery(a model.Astronaut, lowerNeedle string) bool {
	if strings.Contains(strings.ToLower(a.FullName), lowerNeedle) ||
		strings.Contains(strings.ToLower(a.NationalityCode), lowerNeedle) {
		return true
	}
	if a.NationalityLabel != nil && strings.Contains(strings.ToLower(*a.NationalityLabel), lowerNeedle) {
		return true
	}
	return false
}

func matchesMetricQuery(m model.Metric, lowerNeedle string) bool {
	if strings.Contains(strings.ToLower(m.Code), lowerNeedle) ||
		strings.Contains(strings.ToLower(m.Name), lowerNeedle) ||
		strings.Contains(strings.ToLower(m.Category), lowerNeedle) {
		return true
	}
	if m.ShortName != nil && strings.Contains(strings.ToLower(*m.ShortName), lowerNeedle) {
		return true
	}
	return false
}

func normalizeStringSlice(in []string) []string {
	if in == nil {
		return []string{}
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func parseOptionalUUID(raw *string) (pgtype.UUID, error) {
	if raw == nil {
		return pgtype.UUID{}, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return pgtype.UUID{}, nil
	}
	var id pgtype.UUID
	if err := id.Scan(trimmed); err != nil || !id.Valid {
		return pgtype.UUID{}, fmt.Errorf("%w: invalid astronaut_id UUID", ErrBadRequest)
	}
	return id, nil
}

func formatUUID(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func uuidPtr(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	s := formatUUID(id)
	return &s
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
	v := f.Float64
	return &v, nil
}

func cloneStringPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneInt16Ptr(p *int16) *int16 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
