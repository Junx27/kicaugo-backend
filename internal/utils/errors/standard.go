package errors

import (
	"github.com/danielgtaylor/huma/v2"
)

// StandardError implements huma.StatusError for consistent error format.
type StandardError struct {
	Status  int                 `json:"status"`
	Message string              `json:"message"`
	Data    any                 `json:"data"`
	Errors  []*huma.ErrorDetail `json:"errors,omitempty"`
}

// Error mengembalikan pesan error.
func (e *StandardError) Error() string { return e.Message }

// GetStatus mengembalikan status HTTP error.
func (e *StandardError) GetStatus() int { return e.Status }
