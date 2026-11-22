package request

import "encoding/json"

// StartEventStreamingInput adalah input POST /api/v1/events/{id}/streamings/start.
type StartEventStreamingInput struct {
	ID   int64 `path:"id" minimum:"1" doc:"ID event"`
	Body struct {
		LinkStream string          `json:"link_stream" required:"true" minLength:"1" maxLength:"2048" doc:"URL/link streaming"`
		Juara      json.RawMessage `json:"juara,omitempty" doc:"Informasi juara dalam format JSON"`
	}
}

// StopEventStreamingInput adalah input POST /api/v1/events/{id}/streamings/stop.
type StopEventStreamingInput struct {
	ID int64 `path:"id" minimum:"1" doc:"ID event"`
}
