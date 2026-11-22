package response

import (
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

func TestNewSuccess(t *testing.T) {
	got := NewSuccess(201, "created", map[string]string{"id": "1"})
	if got.Status != 201 {
		t.Errorf("Status = %d, ingin 201", got.Status)
	}
	if got.Message != "created" {
		t.Errorf("Message = %q, ingin created", got.Message)
	}
	if got.Data["id"] != "1" {
		t.Errorf("Data = %+v, ingin id=1", got.Data)
	}
}

func TestNewError(t *testing.T) {
	detail := &huma.ErrorDetail{Message: "invalid"}
	got := NewError(400, "bad request", detail)
	if got.Status != 400 {
		t.Errorf("Status = %d, ingin 400", got.Status)
	}
	if got.Message != "bad request" {
		t.Errorf("Message = %q, ingin bad request", got.Message)
	}
	if got.Data != nil {
		t.Errorf("Data = %+v, ingin nil", got.Data)
	}
	if len(got.Errors) != 1 || got.Errors[0] != detail {
		t.Errorf("Errors = %+v, ingin [%+v]", got.Errors, detail)
	}
}
