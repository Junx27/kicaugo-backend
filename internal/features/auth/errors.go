package auth

import "errors"

// Kesalahan domain yang dipetakan ke status HTTP oleh handler.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrIdentifierTaken    = errors.New("identifier already registered")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrNotFound           = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
)
