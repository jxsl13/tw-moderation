package mqtt

import (
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Publisher wraps the mqtt client that is subscribed to a specific topic
// in a pretty simple to use manner.
// initially you connect to your broker and fetch reveived messages with the method
// Next(). Next() is a blocking call that waits for a channel to contain a message or
// until the Close() method has been called that cancels an internally wrapped context, which
// immediatly terminates
type Publisher struct {
	address    string
	clientID   string
	topic      string
	client     mqtt.Client
	msgChannel chan Message
	once       sync.Once
	isClosed   bool
}

// Close waits a second and then closes the client connection as well as the subsciber
// and all internally used channels
func (p *Publisher) Close() {
	p.once.Do(func() {
		if p.client != nil && p.client.IsConnected() {
			p.client.Disconnect(1000)
		}
		close(p.msgChannel)
		p.isClosed = true
		log.Println("Closed Publisher with address:", p.address, "and topic:", p.topic, "with ID: ", p.clientID)
	})
}

// Publish pushes the message into a channel which is emptied by a concurrent goroutine
// and published to th ebroker at the specified topic.
// use string, []byte, int, int64, float32, float64 or
// Message{
//	  Topic string,
//    Payload inteface{}
// }
// allows you to control the taget topic
// if you are using custom structs, please convert them into JSON before
// passing them to this function
func (p *Publisher) Publish(msg interface{}) {
	if p.isClosed {
		log.Println("Publish skipped, channel closed:", msg)
		return
	}
	if p.topic == "" {
		log.Panicln("Trying to publish a message to an empty topic string.")
	}

	switch m := msg.(type) {
	case Message:
		// if it's a message, you can explicity control the topic
		p.msgChannel <- m
	case string, []byte, int, int64, float32, float64, bool:
		// if it's not a message type, you can simply
		// send it to the default topic
		p.msgChannel <- Message{
			Topic:    p.topic,
			Playload: msg,
		}
	default:
		log.Panicf("Invalid type: %T", msg)
	}
	log.Println("Publisher pushed message into channel:", msg)
}

// PublishTo allows to specify a different topic other than the default one.
func (p *Publisher) PublishTo(topic string, msg interface{}) {
	if p.isClosed {
		log.Println("Publish skipped, channel closed:", msg)
		return
	}
	if topic == "" {
		log.Panicln("Trying to publish a message to an empty topic string.")
	}

	switch m := msg.(type) {
	case string, []byte, int, int64, float32, float64, bool:
		// if it's not a message type, you can simply send it to the default
		// topic
		p.msgChannel <- Message{
			Topic:    topic,
			Playload: m,
		}
	default:
		log.Panicf("Invalid type: %T", msg)
	}
	log.Println("Publisher pushed message into channel:", msg)
}

// NewPublisher creates and starts a new Publisher that receives new messages via
// a string channel that can be
// address has the format: tcp://localhost:1883
func NewPublisher(address, clientID, topic string) (*Publisher, error) {

	publisher := &Publisher{
		address:    address,
		clientID:   clientID + topic,
		topic:      topic,
		client:     nil,
		msgChannel: make(chan Message, 64),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(address)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.OnConnect = func(_ mqtt.Client) {
		log.Println("Publisher connected to", address, "and topic:", topic, "with ID:", clientID)
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Println("Publisher lost connection of", address, "with topic:", topic, "and ID:", clientID, "error:", err)
	}
	opts.OnReconnecting = func(client mqtt.Client, options *mqtt.ClientOptions) {
		log.Println("Publisher reconnected to", address, "with topic:", topic, "and ID:", clientID)
	}

	c := mqtt.NewClient(opts)
	publisher.client = c
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		publisher.Close()
		return nil, token.Error()
	}

	// Create two publishing routines, as one might be blocked for receiving the
	// confirmation token.
	go func() {
		for msg := range publisher.msgChannel {
			topic := msg.Topic
			payload := msg.Playload

			if token := publisher.client.Publish(topic, 1, false, payload); token.Wait() && token.Error() != nil {
				log.Println("Publisher could not send message to", publisher.address, "on topic", publisher.topic)
			} else {
				log.Println("Published message at broker:", msg)
			}
		}
	}()

	return publisher, nil
}
