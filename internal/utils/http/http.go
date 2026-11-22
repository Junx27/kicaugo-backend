// Package http menyediakan helper Huma untuk registrasi route dan error HTTP.
package http

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

// API adalah alias huma.API agar handler tidak bergantung langsung pada Huma.
type API = huma.API

// Middleware adalah middleware Huma yang dijalankan sebelum handler.
type Middleware = func(ctx huma.Context, next func(huma.Context))

// Operation adalah konfigurasi route Huma yang reusable.
type Operation struct {
	ID          string
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Security    []map[string][]string
	Middlewares []Middleware
}

// Register mendaftarkan handler ke Huma API dengan konfigurasi Operation.
func Register[I, O any](api API, op Operation, handler func(context.Context, *I) (*O, error)) {
	huma.Register(api, huma.Operation{
		OperationID: op.ID,
		Method:      op.Method,
		Path:        op.Path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Security:    op.Security,
		Middlewares: op.Middlewares,
	}, handler)
}

// Error membuat error HTTP dengan status dan message tertentu.
func Error(status int, message string, errs ...error) error {
	return huma.NewError(status, message, errs...)
}

// BadRequest membuat error HTTP 400.
func BadRequest(message string) error { return huma.Error400BadRequest(message) }

// Unauthorized membuat error HTTP 401.
func Unauthorized(message string) error { return huma.Error401Unauthorized(message) }

// Forbidden membuat error HTTP 403.
func Forbidden(message string) error { return huma.Error403Forbidden(message) }

// NotFound membuat error HTTP 404.
func NotFound(message string) error { return huma.Error404NotFound(message) }

// Conflict membuat error HTTP 409.
func Conflict(message string) error { return huma.Error409Conflict(message) }

// UnprocessableEntity membuat error HTTP 422.
func UnprocessableEntity(message string) error { return huma.Error422UnprocessableEntity(message) }

// InternalServerError membuat error HTTP 500.
func InternalServerError(message string) error { return huma.Error500InternalServerError(message) }
