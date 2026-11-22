// Package entities berisi model domain yang dipakai lintas layer.
package entities

// User adalah entitas pengguna.
type User struct {
	ID           int64
	Name         string
	Identifier   string
	PasswordHash string
}
