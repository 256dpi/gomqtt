package client

import (
	"net"

	"github.com/gomqtt/stream"
	"github.com/gomqtt/message"
)

type Client struct {
	conn net.Conn
	stream *stream.Stream

	connectCallback func(bool)
}

func NewClient() (*Client) {
	return &Client{}
}

func (this *Client)OnConnect(callback func(bool)) {
	this.connectCallback = callback
}

func (this *Client)Connect(url string) (error) {
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		return err
	}

	this.conn = conn
	this.stream = stream.NewStream(conn, conn)

	m := message.NewConnectMessage()
	m.Version = message.Version311
	m.ClientId = []byte("random")
	m.KeepAlive = 30
	m.Username = []byte("")
	m.Password = []byte("")
	m.CleanSession = true
	m.WillTopic = []byte("")
	m.WillPayload = []byte("")
	m.WillQoS = 0
	m.WillRetain = false

	go this.receive()

	this.stream.Out <- m
	return nil
}

func (this *Client)receive(){
	for {
		msg, ok := <- this.stream.In

		if !ok {
			break
		}

		switch msg.Type() {
		case message.CONNACK:
			m, ok := msg.(*message.ConnackMessage)

			if(ok && this.connectCallback != nil) {
				if m.ReturnCode == message.ConnectionAccepted {
					this.connectCallback(m.SessionPresent)
				}
			}
		}
	}
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

func (this *Client)Subscribe(topic []byte, qos byte) {
	m := message.NewSubscribeMessage()
	m.Subscriptions = []message.Subscription{
		message.Subscription{
			Topic: topic,
			QoS: qos,
		},
	}
	m.PacketId = 1

	this.stream.Out <- m
}

func (this *Client)Unsubscribe(topic []byte) {
	m := message.NewUnsubscribeMessage()
	m.Topics = make([][]byte, 1)
	m.Topics[0] = topic
	m.PacketId = 1

	this.stream.Out <- m
}

func (this *Client)Disconnect() {
	m := message.NewDisconnectMessage()

	this.stream.Out <- m
}
