package mqtt

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// NewPublisherSubscriber creates a Publisher as well as a Subscriber from a single connection.
func NewPublisherSubscriber(address, clientID, publishTopic, subscribeTopic string) (*Publisher, *Subscriber, error) {
	publisher := &Publisher{
		address:    address,
		clientID:   clientID,
		topic:      publishTopic,
		client:     nil,
		msgChannel: make(chan string, 1024),
	}

	subscriber := &Subscriber{
		address:    address,
		clientID:   clientID,
		topic:      subscribeTopic,
		client:     nil,
		msgChannel: make(chan string, 1024),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(address)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.OnConnect = func(_ mqtt.Client) {
		log.Println("PubSub connected to", address, "with ID:", clientID)
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Println("PubSub lost connection of", address, "with ID:", clientID, "error:", err)
	}
	opts.OnReconnecting = func(client mqtt.Client, options *mqtt.ClientOptions) {
		log.Println("PubSub reconnected to", address, "with ID:", clientID)
	}

	c := mqtt.NewClient(opts)
	publisher.client = c
	subscriber.client = c
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		publisher.Close()
		subscriber.Close()
		return nil, nil, token.Error()
	}

	// create publishing routine
	go func() {
		for msg := range publisher.msgChannel {
			if t := publisher.client.Publish(publishTopic, 1, false, msg); t.Wait() && t.Error() != nil {
				log.Println("Publisher could not send message to", publisher.address, "on topic", publisher.topic)
			} else {
				log.Println("Published message:", msg)
			}
		}
	}()

	// Subscribe to second topic
	subscriber.client = c
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		publisher.Close()
		subscriber.Close()
		return nil, nil, token.Error()
	}

	fwdHandler := subscriber.getForwardHandler()
	if token := c.Subscribe(subscribeTopic, 1, fwdHandler); token.Wait() && token.Error() != nil {
		publisher.Close()
		subscriber.Close()
		return nil, nil, token.Error()
	}

	return publisher, subscriber, nil
}

// NewTestPublisherSubscriber is a test pubsub that creates two independent clients
func NewTestPublisherSubscriber(address, clientID, publishTopic, subscribeTopic string) (*Publisher, *Subscriber, error) {

	pub, err := NewPublisher(address, clientID+"pub", publishTopic)
	if err != nil {
		return nil, nil, err
	}

	sub, err := NewSubscriber(address, clientID+"sub", subscribeTopic)
	if err != nil {
		return nil, nil, err
	}

	return pub, sub, nil
}
