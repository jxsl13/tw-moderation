package mqtt

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// this package can be used to define reading and writing routines that connect to the Mosquitto broker
// those routines are supposed to run in their own goroutines and either receive or pass events through channels
const (
	QOS = 1
)

var (
	DefaultPublishHandler = func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("UNEXPECTED MESSAGE: %s\n", msg)
	}

	DefaultOnConnectionLostHandler = func(cl mqtt.Client, err error) {
		fmt.Println("connection lost")
	}

	DefaultOnReconnectingHandler = func(mqtt.Client, *mqtt.ClientOptions) {
		fmt.Println("attempting to reconnect")
	}
)
