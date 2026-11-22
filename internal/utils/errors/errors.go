// Package errors menyimpan error domain yang dipakai lintas fitur.
package errors

import "errors"

// Kesalahan domain yang dipetakan ke status HTTP oleh handler.
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrIdentifierTaken      = errors.New("identifier already registered")
	ErrInvalidPayload       = errors.New("invalid payload")
	ErrNotFound             = errors.New("user not found")
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidCity          = errors.New("invalid city")
	ErrProvinceMismatch     = errors.New("province does not match city")
	ErrInvalidCoordinates   = errors.New("invalid coordinates")
	ErrInvalidCategory      = errors.New("invalid category")
	ErrInvalidRadius        = errors.New("invalid radius")
	ErrForbidden            = errors.New("forbidden")
	ErrOrganizerNotFound    = errors.New("organizer not found")
	ErrOrganizerSlugTaken   = errors.New("organizer slug already registered")
	ErrOrganizerMemberTaken = errors.New("organizer member already registered")
)
