package model

import "time"

// IngestReading represents a single metric reading inside a POST / ingestion request.
type IngestReading struct {
	MetricCode   string   `json:"metric_code"`
	ValueNumeric *float64 `json:"value_numeric,omitempty"`
	ValueScale   *int     `json:"value_scale,omitempty"`
	ValueText    *string  `json:"value_text,omitempty"`
	UnitCode     *string  `json:"unit_code,omitempty"`
}

// IngestRequest represents a device measurement session payload posted to POST /.
type IngestRequest struct {
	DeviceCode    string          `json:"device_code"`
	Environment   EnvironmentMode `json:"environment"`
	RecordedAtUTC string          `json:"recorded_at_utc"`
	AstronautName *string         `json:"astronaut_name,omitempty"`
	Readings      []IngestReading `json:"readings"`
}

// IngestResult represents the 201 Created response returned by POST /.
type IngestResult struct {
	MeasurementID string    `json:"measurement_id"`
	RecordedAtUTC time.Time `json:"recorded_at_utc"`
	ValuesCount   int       `json:"values_count"`
}
