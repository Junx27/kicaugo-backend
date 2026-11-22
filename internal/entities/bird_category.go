// Package entities berisi model domain yang dipakai lintas layer.
package entities

import "time"

// BirdCategory adalah entitas kategori burung (lookup umum).
type BirdCategory struct {
	ID        int64
	Name      string
	Slug      string
	Icon      string
	CreatedAt time.Time
}
