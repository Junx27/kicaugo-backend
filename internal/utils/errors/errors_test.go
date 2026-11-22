package errors

import (
	"errors"
	"testing"
)

func TestStandardError(t *testing.T) {
	err := &StandardError{Status: 404, Message: "not found"}
	if err.Error() != "not found" {
		t.Errorf("Error() = %q, ingin not found", err.Error())
	}
	if err.GetStatus() != 404 {
		t.Errorf("GetStatus() = %d, ingin 404", err.GetStatus())
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "invalid credentials", err: ErrInvalidCredentials, want: "invalid credentials"},
		{name: "identifier taken", err: ErrIdentifierTaken, want: "identifier already registered"},
		{name: "invalid payload", err: ErrInvalidPayload, want: "invalid payload"},
		{name: "not found", err: ErrNotFound, want: "user not found"},
		{name: "invalid token", err: ErrInvalidToken, want: "invalid token"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.err) {
				t.Errorf("errors.Is(%v, %v) = false", tc.err, tc.err)
			}
			if tc.err.Error() != tc.want {
				t.Errorf("Error() = %q, ingin %q", tc.err.Error(), tc.want)
			}
		})
	}
}
