// Package response berisi format standar untuk semua response API.
package response

import (
	"github.com/danielgtaylor/huma/v2"
)

// Success adalah format standar untuk response sukses.
type Success[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// NewSuccess membuat response sukses dengan status, message, dan data.
func NewSuccess[T any](status int, message string, data T) Success[T] {
	return Success[T]{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// Error adalah format standar untuk response error (RFC 9457 compatible).
type Error struct {
	Status  int                 `json:"status"`
	Message string              `json:"message"`
	Data    any                 `json:"data"`
	Errors  []*huma.ErrorDetail `json:"errors,omitempty"`
}

// NewError membuat response error dengan status, message, dan data null.
func NewError(status int, message string, errs ...*huma.ErrorDetail) *Error {
	return &Error{
		Status:  status,
		Message: message,
		Data:    nil,
		Errors:  errs,
	}
}

// StandardError implements huma.StatusError for consistent error format.
type StandardError struct {
	Status  int                 `json:"status"`
	Message string              `json:"message"`
	Data    any                 `json:"data"`
	Errors  []*huma.ErrorDetail `json:"errors,omitempty"`
}

func (e *StandardError) Error() string  { return e.Message }
func (e *StandardError) GetStatus() int { return e.Status }
