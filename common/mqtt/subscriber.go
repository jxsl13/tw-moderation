package mqtt

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	debug = false
)

// Subscriber wraps the mqtt client that is subscribed to a specific topic
// in a pretty simple to use manner.
// initially you connect to your broker and fetch reveived messages with the method
// Next(). Next() is a blocking call that waits for a channel to contain a message or
// until the Close() method has been called that cancels an internally wrapped context, which
// immediatly terminates
type Subscriber struct {
	address    string
	clientID   string
	topics     []string
	client     mqtt.Client
	msgChannel chan Message
	once       sync.Once
}

func (s *Subscriber) getForwardHandler() func(mqtt.Client, mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		s.msgChannel <- Message{
			Topic:    msg.Topic(),
			Playload: string(msg.Payload()),
		}
		log.Println("Subscriber pushed message into channel")
	}
}

// Close waits a second and then closes the client connection as well as the subsciber
// and all internally used channels
func (s *Subscriber) Close() {
	s.once.Do(func() {
		if s.client != nil && s.client.IsConnected() {
			for _, topic := range s.topics {
				if token := s.client.Unsubscribe(topic); token.WaitTimeout(time.Second) && token.Error() != nil {
					log.Println("Failed to unsubscribe from", s.address, "with topic:", topic)
				}
			}
			s.client.Disconnect(1000)
		}
		close(s.msgChannel)
		log.Println("Closed subscriber with address: ", s.address, " and topics: ", strings.Join(s.topics, ","), " with ID: ", s.clientID)
	})
}

// Next blocks until the next message from the broker is received
// the bool indicates that the subscriber was closed
// you can use this in a for loop until ok is false, preferrably in an own goroutine
func (s *Subscriber) Next() <-chan Message {
	return s.msgChannel
}

// NewSubscriber creates and starts a new subscriber that receives new messages via
// a string channel that can be
// address has the format: tcp://localhost:1883
func NewSubscriber(address, clientID string, topics ...string) (*Subscriber, error) {
	if debug {
		mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
		mqtt.CRITICAL = log.New(os.Stdout, "[CRITICAL] ", 0)
		mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
		mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	}
	subscriber := &Subscriber{
		address:    address,
		clientID:   clientID,
		topics:     topics,
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
		log.Println("Subscriber connected to", address, "and topics:", strings.Join(topics, ","), "with ID:", clientID)
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Println("Subscriber lost connection of", address, "with topics:", strings.Join(topics, ","), "and ID:", clientID, " error: ", err)
	}
	opts.OnReconnecting = func(client mqtt.Client, options *mqtt.ClientOptions) {
		log.Println("Subscriber reconnected to", address, "with topics:", strings.Join(topics, ","), "and ID:", clientID)
	}

	c := mqtt.NewClient(opts)
	subscriber.client = c
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		subscriber.Close()
		return nil, token.Error()
	}

	successfulSubscriptions := make([]string, 0, len(topics))
	var err error = nil
	for _, topic := range topics {
		if token := c.Subscribe(topic, 1, subscriber.getForwardHandler()); token.Wait() && token.Error() != nil {
			err = token.Error()
			break
		} else {
			successfulSubscriptions = append(successfulSubscriptions, topic)
		}
	}

	if err != nil {
		// needed in order to properly close the connection and unsubscribe from all topics
		subscriber.topics = successfulSubscriptions
		subscriber.Close()
		return nil, err
	}

	return subscriber, nil
}
