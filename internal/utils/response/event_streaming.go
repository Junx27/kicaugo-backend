package response

import "encoding/json"

// EventStreamingData adalah representasi streaming event di response.
type EventStreamingData struct {
	ID         int64           `json:"id"`
	EventID    int64           `json:"event_id"`
	StartTime  string          `json:"start_time"`
	EndTime    *string         `json:"end_time"`
	TotalViews int64           `json:"total_views"`
	Juara      json.RawMessage `json:"juara"`
	LinkStream string          `json:"link_stream"`
}

// StartEventStreamingOutput adalah output POST /api/v1/events/{id}/streamings/start.
type StartEventStreamingOutput struct {
	Status int `status:"201"`
	Body   Success[EventStreamingData]
}
