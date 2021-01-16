package events

// PlayerJoinEvent is created when a player joins some server.
type PlayerJoinEvent struct {
	BaseEvent
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Clan    string `json:"clan,omitempty"`
	Country int    `json:"country,omitempty"`
	IP      string `json:"ip,omitempty"`
	Port    int    `json:"port,omitempty"`
	Version int    `json:"version,omitempty"`
}
