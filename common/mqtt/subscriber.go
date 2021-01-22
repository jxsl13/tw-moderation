package mqtt

import (
	"log"
	"os"
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
	topic      string
	client     mqtt.Client
	msgChannel chan string
}

func (s *Subscriber) getForwardHandler() func(mqtt.Client, mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		s.msgChannel <- string(msg.Payload())
		if Debug {
			log.Println("Subscriber pushed message into channel")
		}
	}
}

// Close waits a second and then closes the client connection as well as the subsciber
// and all internally used channels
func (s *Subscriber) Close() {
	if token := s.client.Unsubscribe(s.topic); token.Wait() && token.Error() != nil {
		log.Println("Unsubscribing from topic: ", s.topic, " failed: ", token.Error())
	}
	s.client.Disconnect(1000)
	close(s.msgChannel)
	log.Println("Closed subscriber with address: ", s.address, " and topic: ", s.topic, " with ID: ", s.clientID)
}

// Next blocks until the next message from the broker is received
// the bool indicates that the subscriber was closed
// you can use this in a for loop until ok is false, preferrably in an own goroutine
func (s *Subscriber) Next() <-chan string {
	return s.msgChannel
}

// NewSubscriber creates and starts a new subscriber that receives new messages via
// a string channel that can be
// address has the format: tcp://localhost:1883
func NewSubscriber(address, clientID, topic string) (*Subscriber, error) {
	if debug {
		mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
		mqtt.CRITICAL = log.New(os.Stdout, "[CRITICAL] ", 0)
		mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
		mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	}
	subscriber := &Subscriber{
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
		log.Println("Subscriber connected to ", address, " and topic: ", topic, " with ID: ", clientID)
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Println("Subscriber lost connection of ", address, " with topic: ", topic, " and ID: ", clientID, " error: ", err)
	}
	opts.OnReconnecting = func(client mqtt.Client, options *mqtt.ClientOptions) {
		log.Println("Subscriber reconnected to ", address, " with topic: ", topic, " and ID: ", clientID)
	}

	c := mqtt.NewClient(opts)
	subscriber.client = c
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		subscriber.Close()
		return nil, token.Error()
	}

	if token := c.Subscribe(topic, 1, subscriber.getForwardHandler()); token.Wait() && token.Error() != nil {
		subscriber.Close()
		return nil, token.Error()
	}

	return subscriber, nil
}
