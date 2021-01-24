package mqtt

// NewPublisherSubscriber currently uss two connections to communicate with the broker.
// based on the clientID a unique ID is created for each connection, given the fact
// that the passe ID is unique.
func NewPublisherSubscriber(address, clientID, publishTopic, subscribeTopic string) (*Publisher, *Subscriber, error) {

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
