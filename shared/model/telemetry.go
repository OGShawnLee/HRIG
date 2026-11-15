package model

import "time"

// ReadingStatus represents the classification state of a telemetry reading.
// "stale" is runtime-only (RNF10) and must never be persisted to PostgreSQL.
type ReadingStatus string

const (
	StatusNormal   ReadingStatus = "normal"
	StatusAbnormal ReadingStatus = "abnormal"
	StatusCritical ReadingStatus = "critical"
	StatusStale    ReadingStatus = "stale"
)

// IsValid reports whether s is a recognized ReadingStatus value.
func (s ReadingStatus) IsValid() bool {
	switch s {
	case StatusNormal, StatusAbnormal, StatusCritical, StatusStale:
		return true
	default:
		return false
	}
}

// IsPersistable reports whether s is valid for storage in PostgreSQL
// (reading_status enum allows normal, abnormal, critical; stale is runtime-only).
func (s ReadingStatus) IsPersistable() bool {
	switch s {
	case StatusNormal, StatusAbnormal, StatusCritical:
		return true
	default:
		return false
	}
}

// EnvironmentMode represents the operational environment (IVA vs EVA).
type EnvironmentMode string

const (
	EnvIVA EnvironmentMode = "iva"
	EnvEVA EnvironmentMode = "eva"
)

// IsValid reports whether e is a recognized EnvironmentMode value.
func (e EnvironmentMode) IsValid() bool {
	switch e {
	case EnvIVA, EnvEVA:
		return true
	default:
		return false
	}
}

// MetricTier represents the HUD display tier of a metric.
type MetricTier string

const (
	TierPrimary   MetricTier = "primary"
	TierSecondary MetricTier = "secondary"
)

// IsValid reports whether t is a recognized MetricTier value.
func (t MetricTier) IsValid() bool {
	switch t {
	case TierPrimary, TierSecondary:
		return true
	default:
		return false
	}
}

// ValueType identifies which measurement_value column holds the reading payload.
type ValueType string

const (
	ValueNumeric ValueType = "numeric"
	ValueScale   ValueType = "scale"
	ValueText    ValueType = "text"
	ValueImage   ValueType = "image"
)

// IsValid reports whether v is a recognized ValueType value.
func (v ValueType) IsValid() bool {
	switch v {
	case ValueNumeric, ValueScale, ValueText, ValueImage:
		return true
	default:
		return false
	}
}

// DeviceKind classifies the capture device or source.
type DeviceKind string

const (
	DeviceWearable          DeviceKind = "wearable"
	DeviceEnvironmentSensor DeviceKind = "environment_sensor"
	DeviceManual            DeviceKind = "manual"
	DeviceLabAssay          DeviceKind = "lab_assay"
	DeviceImaging           DeviceKind = "imaging"
)

// IsValid reports whether d is a recognized DeviceKind value.
func (d DeviceKind) IsValid() bool {
	switch d {
	case DeviceWearable, DeviceEnvironmentSensor, DeviceManual, DeviceLabAssay, DeviceImaging:
		return true
	default:
		return false
	}
}

// MediaKind classifies diagnostic imaging assets.
type MediaKind string

const (
	MediaUltrasound MediaKind = "ultrasound"
	MediaOcular     MediaKind = "ocular"
)

// IsValid reports whether m is a recognized MediaKind value.
func (m MediaKind) IsValid() bool {
	switch m {
	case MediaUltrasound, MediaOcular:
		return true
	default:
		return false
	}
}

// Astronaut represents a crew member profile with derived age.
type Astronaut struct {
	ID               string    `json:"id"`
	FullName         string    `json:"fullName"`
	BirthDate        time.Time `json:"birthDate"`
	Age              int16     `json:"age"`
	NationalityCode  string    `json:"nationalityCode"`
	NationalityLabel *string   `json:"nationalityLabel,omitempty"`
	IsActive         bool      `json:"isActive"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Unit represents a UCUM-coded measurement unit.
type Unit struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Symbol    string `json:"symbol"`
	Dimension string `json:"dimension"`
	Decimals  int16  `json:"decimals"`
	IsActive  bool   `json:"isActive"`
}

// MetricThreshold represents classification limits in the metric's canonical unit.
type MetricThreshold struct {
	ID                 string    `json:"id"`
	MetricID           string    `json:"metricId"`
	MetricCode         string    `json:"metricCode,omitempty"`
	WarningMin         *float64  `json:"warningMin,omitempty"`
	WarningMax         *float64  `json:"warningMax,omitempty"`
	CriticalMin        *float64  `json:"criticalMin,omitempty"`
	CriticalMax        *float64  `json:"criticalMax,omitempty"`
	ClassificationMode string    `json:"classificationMode"`
	Source             string    `json:"source"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// Metric represents a cataloged physiological (metric_type=1) or environmental (metric_type=2) metric.
type Metric struct {
	ID                string           `json:"id"`
	Code              string           `json:"code"`
	Name              string           `json:"name"`
	ShortName         *string          `json:"shortName,omitempty"`
	Category          string           `json:"category"`
	Tier              MetricTier       `json:"tier"`
	MetricType        int16            `json:"metricType"`
	ValueType         ValueType        `json:"valueType"`
	BaseUnitID        *string          `json:"baseUnitId,omitempty"`
	BaseUnitCode      *string          `json:"baseUnitCode,omitempty"`
	BaseUnitSymbol    *string          `json:"baseUnitSymbol,omitempty"`
	BaseUnitDimension *string          `json:"baseUnitDimension,omitempty"`
	BaseUnitDecimals  *int16           `json:"baseUnitDecimals,omitempty"`
	DisplayUnitSymbol *string          `json:"displayUnitSymbol,omitempty"`
	ScaleMin          *int16           `json:"scaleMin,omitempty"`
	ScaleMax          *int16           `json:"scaleMax,omitempty"`
	DisplayOrder      int16            `json:"displayOrder"`
	IsActive          bool             `json:"isActive"`
	Threshold         *MetricThreshold `json:"threshold,omitempty"`
}

// Device represents a telemetry-producing hardware or manual capture device.
type Device struct {
	ID       string     `json:"id"`
	Code     string     `json:"code"`
	Name     string     `json:"name"`
	Kind     DeviceKind `json:"kind"`
	IsActive bool       `json:"isActive"`
}

// TelemetryFrame represents an individual captured reading with provenance and dual timestamps.
type TelemetryFrame struct {
	Source        string          `json:"source"`
	DeviceCode    string          `json:"deviceCode"`
	AstronautID   *string         `json:"astronautId,omitempty"`
	Environment   EnvironmentMode `json:"environment"`
	MetricType    int16           `json:"metricType"`
	MetricCode    string          `json:"metricCode"`
	UnitCode      *string         `json:"unitCode,omitempty"`
	ValueNumeric  *float64        `json:"valueNumeric,omitempty"`
	ValueScale    *int16          `json:"valueScale,omitempty"`
	ValueText     *string         `json:"valueText,omitempty"`
	Status        ReadingStatus   `json:"status"`
	RecordedAtUTC time.Time       `json:"recordedAtUtc"`
	MET           string          `json:"met"`
	Note          *string         `json:"note,omitempty"`
}

// LatestReading pairs a TelemetryFrame with live stream freshness metadata (RNF10).
type LatestReading struct {
	TelemetryFrame
	LastSeenUTC time.Time `json:"lastSeenUtc"`
	IsStale     bool      `json:"isStale"`
}

// CheckInMedia represents a diagnostic media attachment submitted with a manual check-in.
type CheckInMedia struct {
	Kind          MediaKind  `json:"kind"`
	FilePath      string     `json:"filePath"`
	MimeType      string     `json:"mimeType"`
	Findings      *string    `json:"findings,omitempty"`
	CapturedAtUTC *time.Time `json:"capturedAtUtc,omitempty"`
}

// CheckInRequest represents a manual check-in or spot measurement payload (POST /api/checkins).
type CheckInRequest struct {
	AstronautID   *string         `json:"astronautId,omitempty"`
	DeviceCode    *string         `json:"deviceCode,omitempty"`
	Environment   EnvironmentMode `json:"environment"`
	MetricCode    string          `json:"metricCode"`
	UnitCode      *string         `json:"unitCode,omitempty"`
	ValueNumeric  *float64        `json:"valueNumeric,omitempty"`
	ValueScale    *int16          `json:"valueScale,omitempty"`
	ValueText     *string         `json:"valueText,omitempty"`
	Status        *ReadingStatus  `json:"status,omitempty"`
	RecordedAtUTC *time.Time      `json:"recordedAtUtc,omitempty"`
	Note          *string         `json:"note,omitempty"`
	Media         []CheckInMedia  `json:"media,omitempty"`
}

// TrendQuery filters time-series trend queries (GET /api/trends).
type TrendQuery struct {
	AstronautID string `json:"astronautId,omitempty"`
	MetricCode  string `json:"metricCode"`
	Limit       int    `json:"limit"`
}

// TrendPoint represents a single chronological reading returned by GET /api/trends.
type TrendPoint struct {
	ID            string          `json:"id"`
	MeasurementID string          `json:"measurementId"`
	MetricID      string          `json:"metricId"`
	MetricCode    string          `json:"metricCode"`
	AstronautID   *string         `json:"astronautId,omitempty"`
	DeviceID      string          `json:"deviceId"`
	Environment   EnvironmentMode `json:"environment"`
	UnitCode      *string         `json:"unitCode,omitempty"`
	UnitSymbol    *string         `json:"unitSymbol,omitempty"`
	ValueNumeric  *float64        `json:"valueNumeric,omitempty"`
	ValueScale    *int16          `json:"valueScale,omitempty"`
	ValueText     *string         `json:"valueText,omitempty"`
	Status        ReadingStatus   `json:"status"`
	RecordedAtUTC time.Time       `json:"recordedAtUtc"`
	MET           string          `json:"met"`
}

// HistoryQuery filters chronological crew history queries (GET /api/history).
type HistoryQuery struct {
	AstronautID string `json:"astronautId,omitempty"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// HistoryEntry represents a persisted measurement reading (with optional media asset) returned by GET /api/history.
type HistoryEntry struct {
	MeasurementID      string          `json:"measurementId"`
	ValueID            string          `json:"valueId"`
	AstronautID        *string         `json:"astronautId,omitempty"`
	AstronautName      *string         `json:"astronautName,omitempty"`
	DeviceCode         string          `json:"deviceCode"`
	DeviceName         string          `json:"deviceName"`
	Environment        EnvironmentMode `json:"environment"`
	MetricCode         string          `json:"metricCode"`
	MetricName         string          `json:"metricName"`
	ValueType          ValueType       `json:"valueType"`
	UnitCode           *string         `json:"unitCode,omitempty"`
	UnitSymbol         *string         `json:"unitSymbol,omitempty"`
	DisplayUnitSymbol  *string         `json:"displayUnitSymbol,omitempty"`
	ValueNumeric       *float64        `json:"valueNumeric,omitempty"`
	ValueScale         *int16          `json:"valueScale,omitempty"`
	ValueText          *string         `json:"valueText,omitempty"`
	Status             ReadingStatus   `json:"status"`
	RecordedAtUTC      time.Time       `json:"recordedAtUtc"`
	MET                string          `json:"met"`
	MeasurementNote    *string         `json:"measurementNote,omitempty"`
	ValueNote          *string         `json:"valueNote,omitempty"`
	MediaAssetID       *string         `json:"mediaAssetId,omitempty"`
	MediaKind          *MediaKind      `json:"mediaKind,omitempty"`
	MediaFilePath      *string         `json:"mediaFilePath,omitempty"`
	MediaMimeType      *string         `json:"mediaMimeType,omitempty"`
	MediaFindings      *string         `json:"mediaFindings,omitempty"`
	MediaCapturedAtUTC *time.Time      `json:"mediaCapturedAtUtc,omitempty"`
}

// AlertEvent represents an in-memory crew alert triggered by a critical reading (RF08).
// Alerts are not persisted in PostgreSQL (.routing/BD.md D7).
type AlertEvent struct {
	ID                string          `json:"id"`
	AstronautID       *string         `json:"astronautId,omitempty"`
	Source            string          `json:"source"`
	DeviceCode        string          `json:"deviceCode,omitempty"`
	Environment       EnvironmentMode `json:"environment"`
	MetricCode        string          `json:"metricCode"`
	ValueNumeric      *float64        `json:"valueNumeric,omitempty"`
	ValueScale        *int16          `json:"valueScale,omitempty"`
	ValueText         *string         `json:"valueText,omitempty"`
	UnitCode          *string         `json:"unitCode,omitempty"`
	Status            ReadingStatus   `json:"status"`
	Message           string          `json:"message"`
	TriggeredAtUTC    time.Time       `json:"triggeredAtUtc"`
	MET               string          `json:"met"`
	Acknowledged      bool            `json:"acknowledged"`
	AcknowledgedBy    *string         `json:"acknowledgedBy,omitempty"`
	AcknowledgedAtUTC *time.Time      `json:"acknowledgedAtUtc,omitempty"`
}

// AckAlertRequest represents the optional request body for POST /api/alerts/{id}/ack.
type AckAlertRequest struct {
	AcknowledgedBy string `json:"acknowledgedBy"`
}

// Snapshot represents the current HUD state returned by GET /api/snapshot.
type Snapshot struct {
	Readings       []LatestReading `json:"readings"`
	Alerts         []AlertEvent    `json:"alerts"`
	GeneratedAtUTC time.Time       `json:"generatedAtUtc"`
}

// RealtimeEnvelope is the tagged WebSocket payload for /ws/realtime and bidirectional alerts.
// Type must be one of: "snapshot", "reading", "stale", "alert", "alert_ack".
type RealtimeEnvelope struct {
	Type      string          `json:"type"`
	Readings  []LatestReading `json:"readings,omitempty"`
	Alerts    []AlertEvent    `json:"alerts,omitempty"`
	Reading   *LatestReading  `json:"reading,omitempty"`
	Alert     *AlertEvent     `json:"alert,omitempty"`
	SentAtUTC time.Time       `json:"sentAtUtc"`
}

// HealthStatus represents the API health payload returned by GET /health.
type HealthStatus struct {
	Status          string    `json:"status"`
	Store           string    `json:"store"`
	TimeUTC         time.Time `json:"timeUtc"`
	ActiveAlerts    int       `json:"activeAlerts"`
	RealtimeClients int       `json:"realtimeClients"`
	IngestClients   int       `json:"ingestClients"`
}
