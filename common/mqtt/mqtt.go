package mqtt

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// this package can be used to define reading and writing routines that connect to the Mosquitto broker
// those routines are supposed to run in their own goroutines and either receive or pass events through channels
const (
	QOS = 1
)

// TODO: continue down in the Subscriber
var (
	// Address is the default address of theMosquitto broker
	Address = "tcp://mosquitto:1883"
)

// Subscriber returns a channel that contains new event messages
func Subscriber(ctx context.Context, clientID, topic string) <-chan string {
	subscriber := make(chan string, 1024)

	go func(ctx context.Context, clientID, address, topic string, subscriber chan<- string) {
		defer close(subscriber)

		opts := mqtt.NewClientOptions()
		opts.AddBroker(address)
		opts.SetClientID(clientID)
		opts.SetConnectTimeout(time.Second)
		opts.SetWriteTimeout(time.Second)
		opts.SetKeepAlive(10 * time.Second)
		opts.SetPingTimeout(time.Second)
		opts.SetConnectRetry(true)
		opts.SetAutoReconnect(true)

		opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
			log.Printf("topic: %s UNEXPECTED MESSAGE: %s\n\n", topic, msg)
		})

		// Log events
		opts.OnConnectionLost = func(cl mqtt.Client, err error) {
			log.Printf("connection lost to topic: %s", topic)
		}

		opts.OnConnect = func(c mqtt.Client) {
			fmt.Println("connection established")

			// Establish the subscription - doing this here means that it will happen every time a connection is established
			// (useful if opts.CleanSession is TRUE or the broker does not reliably store session data)
			t := c.Subscribe(topic, QOS, h.handle)
			// the connection handler is called in a goroutine so blocking here would hot cause an issue. However as blocking
			// in other handlers does cause problems its best to just assume we should not block
			go func() {
				_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
				if t.Error() != nil {
					fmt.Printf("ERROR SUBSCRIBING: %s\n", t.Error())
				} else {
					fmt.Println("subscribed to: ", TOPIC)
				}
			}()
		}

	}(ctx, clientID, Address, topic, subscriber)

	return subscriber
}
