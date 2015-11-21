package client

import (
	"net"

	"github.com/gomqtt/stream"
	"github.com/gomqtt/message"
	"errors"
	"net/url"
	"crypto/tls"
)

type Client struct {
	conn net.Conn
	stream *stream.Stream

	connectCallback func(bool)
	errorCallback func(error)
}

// NewClient returns a new client.
func NewClient() (*Client) {
	return &Client{}
}

// OnConnect sets the callback for successful connections.
func (this *Client)OnConnect(callback func(bool)) {
	this.connectCallback = callback
}

// OnError sets the callback for failed connection attempts and parsing errors.
func (this *Client)OnError(callback func(error)) {
	this.errorCallback = callback
}

// QuickConnect opens a connection to the broker specified by an URL.
func (this *Client)QuickConnect(urlString string, clientID string) (error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	opts := &Options{
		URL: url,
		ClientID: clientID,
		CleanSession: true,
	}

	return this.Connect(opts)
}

// Connect opens the connection to the broker.
func (this *Client)Connect(opts *Options) (error) {
	host, port, err := net.SplitHostPort(opts.URL.Host)
	if err != nil {
		return err
	}

	// prepare conn
	var conn net.Conn

	// connect based on scheme
	switch opts.URL.Scheme {
	case "mqtt", "tcp":
		if port == "" {
			port = "1883"
		}

		conn, err = net.Dial("tcp", net.JoinHostPort(host, port))
	case "mqtts":
		if port == "" {
			port = "8883"
		}

		conn, err = tls.Dial("tcp", net.JoinHostPort(host, port), nil)
	case "ws":
		panic("ws protocol not implemented!")
	case "wss":
		panic("wss protocol not implemented!")
	}

	// check for errors
	if err != nil {
		return err
	}

	this.conn = conn
	this.stream = stream.NewStream(conn, conn)

	u := ""
	p := ""

	if opts.URL.User != nil {
		u = opts.URL.User.Username()
		p, _ = opts.URL.User.Password()
	}

	m := message.NewConnectMessage()
	m.Version = message.Version311
	m.ClientId = []byte(opts.ClientID)
	m.KeepAlive = uint16(opts.KeepAlive.Seconds())
	m.Username = []byte(u)
	m.Password = []byte(p)
	m.CleanSession = opts.CleanSession
	m.WillTopic = []byte(opts.WillTopic)
	m.WillPayload = opts.WillPayload
	m.WillQoS = opts.WillQos
	m.WillRetain = opts.WillRetained

	go this.receive()
	go this.watch()

	this.stream.Out <- m
	return nil
}

func (this *Client)Publish(topic []byte, payload []byte, qos int, retain bool) {
	m := message.NewPublishMessage()
	m.Topic = topic
	m.Payload = payload
	m.QoS = 0x00
	m.Retain = retain
	m.Dup = false
	m.PacketId = 1

	this.stream.Out <- m
}

func (this *Client)Subscribe(topic string, qos byte) {
	this.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

func (this *Client)SubscribeMultiple(filters map[string]byte) {
	m := message.NewSubscribeMessage()
	m.Subscriptions = make([]message.Subscription, 0, len(filters))
	m.PacketId = 1

	for topic, qos := range filters {
		m.Subscriptions = append(m.Subscriptions, message.Subscription{
			Topic: []byte(topic),
			QoS: qos,
		})
	}

	this.stream.Out <- m
}

func (this *Client)Unsubscribe(topic []byte) {
	this.UnsubscribeMultiple([][]byte{topic})
}

func (this *Client)UnsubscribeMultiple(topics [][]byte) {
	m := message.NewUnsubscribeMessage()
	m.Topics = topics
	m.PacketId = 1

	this.stream.Out <- m
}

func (this *Client)Disconnect() {
	m := message.NewDisconnectMessage()

	this.stream.Out <- m
}

// process incoming messages
func (this *Client)receive(){
	for {
		msg, ok := <- this.stream.In

		if !ok {
			break
		}

		switch msg.Type() {
		case message.CONNACK:
			m, ok := msg.(*message.ConnackMessage)

			if ok {
				if m.ReturnCode == message.ConnectionAccepted {
					if this.connectCallback != nil {
						this.connectCallback(m.SessionPresent)
					}
				} else {
					if this.errorCallback != nil {
						this.errorCallback(errors.New(m.ReturnCode.Error()))
					}
				}
			} else {
				if this.errorCallback != nil {
					this.errorCallback(errors.New("failed to convert CONNACK message"))
				}
			}
		}
	}
}

// watch for EOFs and errors on the stream
func (this *Client)watch(){
	for {
		select {
		case err := <- this.stream.Error:
			if this.errorCallback != nil {
				this.errorCallback(err)
			}
		case <- this.stream.EOF:
			println("connection closed")
		}
	}
}
