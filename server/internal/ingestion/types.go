package ingestion

import (
	"errors"
	"net/http"

	"hrig/shared/model"
)

type IngestRequest = model.IngestRequest
type ReadingInput = model.IngestReading
type IngestResult = model.IngestResult

type FramePublisher interface {
	PublishFrames(frames []model.TelemetryFrame)
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
