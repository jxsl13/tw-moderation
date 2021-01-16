package events

// BaseEvent contains the common fields that all events have.
type BaseEvent struct {
	// Type of the event
	Type string `json:"type,omitempty"`

	// Source server address:port that has created the event
	// Every server has its own topic that it is listening on, which is called
	// This is usually "<econ ip>:<econ port>"
	Source string `json:"source,omitempty"`

	// When was the event created
	Timestamp string `json:"timestamp,omitempty"`
}
