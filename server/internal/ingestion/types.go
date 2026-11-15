package ingestion

import (
	"errors"
	"net/http"
	"time"
)

type IngestRequest struct {
	DeviceCode    string         `json:"device_code"`
	Environment   string         `json:"environment"`
	RecordedAtUTC string         `json:"recorded_at_utc"`
	AstronautName *string        `json:"astronaut_name,omitempty"`
	Readings      []ReadingInput `json:"readings"`
}

type ReadingInput struct {
	MetricCode   string   `json:"metric_code"`
	ValueNumeric *float64 `json:"value_numeric,omitempty"`
	ValueScale   *int     `json:"value_scale,omitempty"`
	ValueText    *string  `json:"value_text,omitempty"`
	UnitCode     *string  `json:"unit_code,omitempty"`
}

type IngestResult struct {
	MeasurementID string    `json:"measurement_id"`
	RecordedAtUTC time.Time `json:"recorded_at_utc"`
	ValuesCount   int       `json:"values_count"`
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return e.Message
}

func badRequest(msg string) error {
	return &Error{StatusCode: http.StatusBadRequest, Message: msg}
}

func notFound(msg string) error {
	return &Error{StatusCode: http.StatusNotFound, Message: msg}
}

func conflict(msg string) error {
	return &Error{StatusCode: http.StatusConflict, Message: msg}
}

func internalError(msg string) error {
	return &Error{StatusCode: http.StatusInternalServerError, Message: msg}
}

func StatusCodeFromError(err error) (int, string) {
	var ingErr *Error
	if errors.As(err, &ingErr) {
		return ingErr.StatusCode, ingErr.Message
	}
	return http.StatusInternalServerError, "internal server error"
}
