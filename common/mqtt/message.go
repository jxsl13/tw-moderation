package mqtt

// Message is a simple struct that allows to publish to different topics
// with a single publisher client connection
type Message struct {
	Topic    string
	Playload interface{}
}
