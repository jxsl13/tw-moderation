package main

// Connect to the broker, subscribe, and write messages received to a file

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/jxsl13/tw-moderation/common/mqtt"
)

var (
	topic         = "topic1"
	serverAddress = "tcp://mosquitto:1883"
	clientID      = "detect-vpn"
)

func init() {
	serverAddress = os.Getenv("BROKER_ADDRESS")
	clientID = os.Getenv("CLIENT_ID")
}

func main() {

	subscriber, err := mqtt.NewSubscriber(serverAddress, clientID, topic)
	if err != nil {
		log.Fatalln("Could not establish subscriber connection: ", err)
	}
	defer subscriber.Close()
	go func() {
		for msg, ok := subscriber.Next(); ok; {
			log.Println("MSG: ", msg)
		}
	}()

	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Connection is up, press Ctrl-C to shutdown")
	<-sig
	fmt.Println("signal caught - exiting")
	fmt.Println("shutdown complete")
}
