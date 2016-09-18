package client

import "github.com/gomqtt/packet"

// ClearSession will connect/disconnect once with a clean session request to force
// the broker to reset the clients session. This is useful in situations where
// its not clear in what state the last session was left.
func ClearSession(opts *Options) error {
	if opts == nil {
		panic("No options specified")
	}

	client := New()

	// connect to broker
	future, err := client.Connect(&Options{
		BrokerURL:    opts.BrokerURL,
		ClientID:     opts.ClientID,
		CleanSession: true,
		KeepAlive:    opts.KeepAlive,
	})
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
func ClearRetainedMessage(opts *Options, topic string) error {
	client := New()

	// connect to broker
	future, err := client.Connect(&Options{
		BrokerURL:    opts.BrokerURL,
		CleanSession: true,
		KeepAlive:    opts.KeepAlive,
	})
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
