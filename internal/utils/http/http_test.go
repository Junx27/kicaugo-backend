package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

type testInput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type testOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func TestRegister(t *testing.T) {
	_, api := humatest.New(t)
	middlewareCalled := false

	Register(api, Operation{
		ID:          "test",
		Method:      nethttp.MethodPost,
		Path:        "/test",
		Summary:     "Test",
		Description: "Deskripsi test",
		Tags:        []string{"test"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []Middleware{func(ctx huma.Context, next func(huma.Context)) {
			middlewareCalled = true
			next(ctx)
		}},
	}, func(_ context.Context, input *testInput) (*testOutput, error) {
		out := &testOutput{}
		out.Body.Message = input.Body.Name
		return out, nil
	})

	resp := api.Post("/test", "Content-Type: application/json", strings.NewReader(`{"name":"kicau"}`))
	if resp.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, ingin %d; body = %s", resp.Code, nethttp.StatusOK, resp.Body.String())
	}
	if !middlewareCalled {
		t.Error("middleware tidak dijalankan")
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("body bukan JSON: %v", err)
	}
	if body.Message != "kicau" {
		t.Errorf("message = %q, ingin kicau", body.Message)
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "error", err: Error(418, "teapot"), status: 418},
		{name: "bad request", err: BadRequest("bad"), status: 400},
		{name: "unauthorized", err: Unauthorized("unauthorized"), status: 401},
		{name: "forbidden", err: Forbidden("forbidden"), status: 403},
		{name: "not found", err: NotFound("not found"), status: 404},
		{name: "conflict", err: Conflict("conflict"), status: 409},
		{name: "unprocessable entity", err: UnprocessableEntity("invalid"), status: 422},
		{name: "internal server error", err: InternalServerError("error"), status: 500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			statusErr, ok := tc.err.(huma.StatusError)
			if !ok {
				t.Fatalf("error = %T, ingin huma.StatusError", tc.err)
			}
			if statusErr.GetStatus() != tc.status {
				t.Errorf("status = %d, ingin %d", statusErr.GetStatus(), tc.status)
			}
		})
	}
}
