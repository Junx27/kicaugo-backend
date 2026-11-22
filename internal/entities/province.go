package entities

import "time"

// Province adalah entitas province.
type Province struct {
	ID        int64
	Code      string
	Name      string
	Slug      string
	CreatedAt time.Time
}
