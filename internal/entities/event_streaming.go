package entities

import (
	"encoding/json"
	"time"
)

// EventStreaming adalah sesi streaming event yang sedang berlangsung.
type EventStreaming struct {
	ID         int64
	EventID    int64
	StartTime  time.Time
	EndTime    *time.Time
	TotalViews int64
	Juara      json.RawMessage
	LinkStream string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// StartEventStreamingInput adalah data untuk memulai streaming event.
type StartEventStreamingInput struct {
	EventID    int64
	LinkStream string
	Juara      json.RawMessage
}

// ActiveEventStreaming adalah ringkasan streaming aktif untuk halaman home.
type ActiveEventStreaming struct {
	ID         int64
	EventID    int64
	EventTitle string
	EventSlug  string
	Venue      string
	City       string
	LinkStream string
	StartTime  time.Time
	EndTime    *time.Time
	TotalViews int64
	Juara      json.RawMessage
}
