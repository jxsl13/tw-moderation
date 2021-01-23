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
	msgChannel chan string
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
func (p *Publisher) Publish(msg string) {
	if p.isClosed {
		log.Println("Publish skipped, channel closed:", msg)
		return
	}
	p.msgChannel <- msg
	log.Println("Publisher pushed message into channel:", msg)
}

// NewPublisher creates and starts a new Publisher that receives new messages via
// a string channel that can be
// address has the format: tcp://localhost:1883
func NewPublisher(address, clientID, topic string) (*Publisher, error) {

	publisher := &Publisher{
		address:    address,
		clientID:   clientID,
		topic:      topic,
		client:     nil,
		msgChannel: make(chan string, 1024),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(address)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
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
			if token := publisher.client.Publish(topic, 1, false, msg); token.Wait() && token.Error() != nil {
				log.Println("Publisher could not send message to", publisher.address, "on topic", publisher.topic)
			} else {
				log.Println("Published message at broker:", msg)
			}
		}
	}()

	return publisher, nil
}
