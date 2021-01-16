package events

// ExecCommand is used to request a command execution
type ExecCommand struct {
	BaseEvent
	User    string
	Command string
}
