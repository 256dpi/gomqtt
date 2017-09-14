package client

import "github.com/gomqtt/packet"

// ClearSession will connect/disconnect once with a clean session request to force
// the broker to reset the clients session. This is useful in situations where
// its not clear in what state the last session was left.
func ClearSession(config *Config) error {
	if config == nil {
		panic("No config specified")
	}

	client := New()

	// copy config
	config2 := *config
	config2.CleanSession = true

	// connect to broker
	future, err := client.Connect(&config2)
	if err != nil {
		return err
	}

	// wait for connack
	future.Wait()

	// check if connection has been accepted
	if future.ReturnCode != packet.ConnectionAccepted {
		return ErrClientConnectionDenied
	}

	// disconnect
	return client.Disconnect()
}

// ClearRetainedMessage will connect/disconnect and send an empty retained message.
// This is useful in situations where its not clear if a message has already been
// retained.
func ClearRetainedMessage(config *Config, topic string) error {
	if config == nil {
		panic("No config specified")
	}

	client := New()

	// copy config
	config2 := *config
	config2.CleanSession = true

	// connect to broker
	future, err := client.Connect(&config2)
	if err != nil {
		return err
	}

	// wait for connack
	future.Wait()

	// check if connection has been accepted
	if future.ReturnCode != packet.ConnectionAccepted {
		return ErrClientConnectionDenied
	}

	// clear retained message
	_, err = client.Publish(topic, nil, 0, true)
	if err != nil {
		return err
	}

	// disconnect
	return client.Disconnect()
}

func PublishMessage(config *Config, msg *packet.Message) error {
	if config == nil {
		panic("No config specified")
	}

	client := New()

	// copy config
	config2 := *config
	config2.CleanSession = true

	// connect to broker
	future, err := client.Connect(&config2)
	if err != nil {
		return err
	}

	// wait for connack
	future.Wait()

	// check if connection has been accepted
	if future.ReturnCode != packet.ConnectionAccepted {
		return ErrClientConnectionDenied
	}

	// clear retained message
	_, err = client.PublishMessage(msg)
	if err != nil {
		return err
	}

	// disconnect
	return client.Disconnect()
}

func ReceiveMessage(config *Config, topic string, qos byte) (*packet.Message, error) {
	if config == nil {
		panic("No config specified")
	}

	client := New()

	// copy config
	config2 := *config
	config2.CleanSession = true

	// connect to broker
	future, err := client.Connect(&config2)
	if err != nil {
		return nil, err
	}

	// wait for connack
	future.Wait()

	// check if connection has been accepted
	if future.ReturnCode != packet.ConnectionAccepted {
		return nil, ErrClientConnectionDenied
	}

	// create channel
	msgCh := make(chan *packet.Message)
	errCh := make(chan error)

	// set callback
	client.Callback = func(msg *packet.Message, err error) {
		if err != nil {
			errCh <- err
			return
		}

		msgCh <- msg
	}

	// make subscription
	_, err = client.Subscribe(topic, qos)
	if err != nil {
		return nil, err
	}

	// wait
	select {
	case msg := <-msgCh:
		// disconnect
		return msg, client.Disconnect()
	case err := <-errCh:
		return nil, err
	}
}
