package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jxsl13/tw-moderation/common/mqtt"
)

// Connect to the broker and publish a message periodically

var (
	topic         = "topic1"
	serverAddress = "tcp://localhost:1883"
	clientID      = "publisher"
)

func init() {
	if brokerAddress := os.Getenv("BROKER_ADDRESS"); brokerAddress != "" {
		serverAddress = brokerAddress
	}
	if id := os.Getenv("BROKER_CLIENT_ID"); id != "" {
		clientID = id
	}
	if t := os.Getenv("BROKER_TOPIC"); t != "" {
		topic = t
	}

	log.Println("Initialized with address: ", serverAddress, " clientID: ", clientID)
}

func main() {
	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	sig := make(chan os.Signal, 1)
	publisher, subscriber, err := mqtt.NewPublisherSubscriber(serverAddress, clientID, "PUBSUB", "PUBSUB")
	if err != nil {
		log.Fatalln("Could not create Publisher:", err)
	}
	defer publisher.Close()
	defer subscriber.Close()

	go func() {
		cnt := 0
		for {
			select {
			case <-time.After(time.Second):
				cnt++
				publisher.Publish(strconv.Itoa(cnt))
				log.Println("Published message:", cnt)
			case <-sig:
				return
			}
		}
	}()

	go func() {
		for msg := range subscriber.Next() {
			log.Println("Received message: ", msg)
		}
	}()

	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Connection is up, press Ctrl-C to shutdown")
	<-sig
	fmt.Println("signal caught - exiting")
	fmt.Println("shutdown complete")
}
