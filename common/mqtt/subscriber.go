package mqtt

// Connect to the broker, subscribe, and write messages received to a file

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// handler is a simple struct that provides a function to be called when a message is received. The message is parsed
// and the count followed by the raw message is written to the file (this makes it easier to sort the file)
type handler struct {
	f *os.File
}

type Subscriber struct {
	address    string
	clientID   string
	topic      string
	ctx        context.Context
	msgChannel chan string
	client     mqtt.Client
	cancel     func()
}

func (s *Subscriber) getForwardHandler() func(mqtt.Client, mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		s.msgChannel <- string(msg.Payload())
		log.Printf("Message received %s", msg.Payload())
	}
}

func (s *Subscriber) Close() {
	s.client.Disconnect(1000)
	close(s.msgChannel)
	s.cancel()
	log.Println("Closed subscriber: ", s.clientID, " : ", s.topic)
}

// Next blocks until the next message from the broker is received
// the bool indicates that the subscriber was closed
// you can use this in a for loop until ok is false, preferrably in an own goroutine
func (s *Subscriber) Next() (msg string, ok bool) {
	select {
	case msg, ok = <-s.msgChannel:
		return
	case <-s.ctx.Done():
		return "", false
	}
}

// NewSubscriber creates and starts a new subscriber that receives new messages via
// a string channel that can be
func NewSubscriber(address, clientID, topic string) (*Subscriber, error) {
	//
	ctx, cancel := context.WithCancel(context.Background())
	subscriber := &Subscriber{
		address:    address,
		clientID:   clientID,
		topic:      topic,
		ctx:        ctx,
		cancel:     cancel,
		msgChannel: make(chan string, 64),
	}

	// Now we establish the connection to the mqtt broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker(address)
	opts.SetClientID(clientID)

	opts.ConnectTimeout = time.Second // Minimal delays on connect
	opts.WriteTimeout = time.Second   // Minimal delays on writes
	opts.KeepAlive = 10               // Keepalive every 10 seconds so we quickly detect network outages
	opts.PingTimeout = time.Second    // local broker so response should be quick

	// Automate connection management (will keep trying to connect and will reconnect if network drops)
	opts.ConnectRetry = true
	opts.AutoReconnect = true

	// If using QOS2 and CleanSession = FALSE then it is possible that we will receive messages on topics that we
	// have not subscribed to here (if they were previously subscribed to they are part of the session and survive
	// disconnect/reconnect). Adding a DefaultPublishHandler lets us detect this.
	opts.DefaultPublishHandler = DefaultPublishHandler

	// Log events
	opts.OnConnectionLost = DefaultOnConnectionLostHandler

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Printf("connection established: clientID: %s topic: %s\n", clientID, topic)

		// Establish the subscription - doing this here means that it will happen every time a connection is established
		// (useful if opts.CleanSession is TRUE or the broker does not reliably store session data)
		t := c.Subscribe(topic, QOS, subscriber.getForwardHandler())
		// the connection handler is called in a goroutine so blocking here would hot cause an issue. However as blocking
		// in other handlers does cause problems its best to just assume we should not block
		go func() {
			<-t.Done() // Can also use '<-t.Done()' in releases > 1.2.0
			if t.Error() != nil {
				fmt.Printf("ERROR SUBSCRIBING: clientID: %s topic: %s : %s\n", clientID, topic, t.Error())
				// close subscriber
				subscriber.cancel()
			} else {
				fmt.Println("clientID: %s subscribed to: %s", clientID, topic)
			}
		}()
	}
	opts.OnReconnecting = DefaultOnReconnectingHandler

	// Connect to the broker
	subscriber.client = mqtt.NewClient(opts)

	// If using QOS2 and CleanSession = FALSE then messages may be transmitted to us before the subscribe completes.
	// Adding routes prior to connecting is a way of ensuring that these messages are processed
	subscriber.client.AddRoute(topic, subscriber.getForwardHandler())

	if token := subscriber.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return subscriber, nil
}
