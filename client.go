package client

import (
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/gomqtt/message"
	"github.com/gomqtt/stream"
	"sync"
)

type (
	ConnectCallback func(bool)
	MessageCallback func(string, []byte)
	ErrorCallback   func(error)
)

type Client struct {
	opts   *Options
	conn   net.Conn
	stream stream.Stream
	quit   chan bool

	connectCallback ConnectCallback
	messageCallback MessageCallback
	errorCallback   ErrorCallback

	lastContact     time.Time
	pingrespPending bool
	start           sync.WaitGroup
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{
		quit: make(chan bool),
	}
}

// OnConnect sets the callback for successful connections.
func (this *Client) OnConnect(callback ConnectCallback) {
	this.connectCallback = callback
}

// OnMessage sets the callback for incoming messages.
func (this *Client) OnMessage(callback MessageCallback) {
	this.messageCallback = callback
}

// OnError sets the callback for failed connection attempts and parsing errors.
func (this *Client) OnError(callback ErrorCallback) {
	this.errorCallback = callback
}

// QuickConnect opens a connection to the broker specified by an URL.
func (this *Client) QuickConnect(urlString string, clientID string) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	opts := &Options{
		URL:          url,
		ClientID:     clientID,
		CleanSession: true,
	}

	return this.Connect(opts)
}

// Connect opens the connection to the broker.
func (this *Client) Connect(opts *Options) error {
	host, port, err := net.SplitHostPort(opts.URL.Host)
	if err != nil {
		return err
	}

	// connect based on scheme
	switch opts.URL.Scheme {
	case "mqtt", "tcp":
		if port == "" {
			port = "1883"
		}

		conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return err
		}

		this.stream = stream.NewNetStream(conn)
	case "mqtts":
		if port == "" {
			port = "8883"
		}

		conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), nil)
		if err != nil {
			return err
		}

		this.stream = stream.NewNetStream(conn)
	case "ws":
		panic("ws protocol not implemented!")
	case "wss":
		panic("wss protocol not implemented!")
	}

	// save opts
	this.opts = opts

	// start watcher and processor
	this.start.Add(3)
	go this.process()
	go this.keepAlive()
	go this.watch()

	// prepare connect message
	m := message.NewConnectMessage()
	m.Version = message.Version311
	m.ClientId = []byte(opts.ClientID)
	m.KeepAlive = uint16(opts.KeepAlive.Seconds())
	m.CleanSession = opts.CleanSession

	// check for credentials
	if opts.URL.User != nil {
		m.Username = []byte(opts.URL.User.Username())
		p, _ := opts.URL.User.Password()
		m.Password = []byte(p)
	}

	// set will
	m.WillTopic = []byte(opts.WillTopic)
	m.WillPayload = opts.WillPayload
	m.WillQoS = opts.WillQos
	m.WillRetain = opts.WillRetained

	// send connect message
	this.send(m)

	this.start.Wait()
	return nil
}

func (this *Client) Publish(topic string, payload []byte, qos byte, retain bool) {
	m := message.NewPublishMessage()
	m.Topic = []byte(topic)
	m.Payload = payload
	m.QoS = qos
	m.Retain = retain
	m.Dup = false
	m.PacketId = 1

	this.send(m)
}

func (this *Client) Subscribe(topic string, qos byte) {
	this.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

func (this *Client) SubscribeMultiple(filters map[string]byte) {
	m := message.NewSubscribeMessage()
	m.Subscriptions = make([]message.Subscription, 0, len(filters))
	m.PacketId = 1

	for topic, qos := range filters {
		m.Subscriptions = append(m.Subscriptions, message.Subscription{
			Topic: []byte(topic),
			QoS:   qos,
		})
	}

	this.send(m)
}

func (this *Client) Unsubscribe(topic string) {
	this.UnsubscribeMultiple([]string{topic})
}

func (this *Client) UnsubscribeMultiple(topics []string) {
	m := message.NewUnsubscribeMessage()
	m.Topics = make([][]byte, 0, len(topics))
	m.PacketId = 1

	for _, t := range topics {
		m.Topics = append(m.Topics, []byte(t))
	}

	this.send(m)
}

func (this *Client) Disconnect() {
	m := message.NewDisconnectMessage()

	this.send(m)

	close(this.quit)
}

// process incoming messages
func (this *Client) process() {
	this.start.Done()

	for {
		select {
		case <-this.quit:
			return
		case msg, ok := <-this.stream.Incoming():
			if !ok {
				return
			}

			this.lastContact = time.Now()

			switch msg.Type() {
			case message.CONNACK:
				m, ok := msg.(*message.ConnackMessage)

				if ok {
					if m.ReturnCode == message.ConnectionAccepted {
						if this.connectCallback != nil {
							this.connectCallback(m.SessionPresent)
						}
					} else {
						this.error(errors.New(m.ReturnCode.Error()))
					}
				} else {
					this.error(errors.New("failed to convert CONNACK message"))
				}
			case message.PINGRESP:
				_, ok := msg.(*message.PingrespMessage)

				if ok {
					this.pingrespPending = false
				} else {
					this.error(errors.New("failed to convert PINGRESP message"))
				}
			case message.PUBLISH:
				m, ok := msg.(*message.PublishMessage)

				if ok {
					this.messageCallback(string(m.Topic), m.Payload)
				} else {
					this.error(errors.New("failed to convert PUBLISH message"))
				}
			}
		}
	}
}

func (this *Client) send(msg message.Message) {
	this.lastContact = time.Now()
	this.stream.Send(msg)
}

func (this *Client) error(err error) {
	if this.errorCallback != nil {
		this.errorCallback(err)
	}
}

// manages the sending of ping packets to keep the connection alive
func (this *Client) keepAlive() {
	this.start.Done()

	for {
		select {
		case <-this.quit:
			return
		default:
			last := uint(time.Since(this.lastContact).Seconds())

			if last > uint(this.opts.KeepAlive.Seconds()) {
				if !this.pingrespPending {
					this.pingrespPending = true

					m := message.NewPingrespMessage()
					this.send(m)
				} else {
					// TODO: close connection?
					// return
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}

// watch for EOFs and errors on the stream
func (this *Client) watch() {
	this.start.Done()

	for {
		select {
		case <-this.quit:
			return
		case err, ok := <-this.stream.Error():
			if !ok {
				return
			}

			this.error(err)
		}
	}
}
