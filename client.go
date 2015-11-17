package client

import (
	"net"

	"github.com/gomqtt/stream"
	"github.com/gomqtt/message"
)

type Client struct {
	conn *net.Conn
	stream *stream.Stream
}

func NewClient() (*Client) {
	return &Client{}
}

func (this *Client)Connect(url string) {
	m := message.NewConnectMessage()
	m.Version = message.Version311
	m.ClientId = "random"
	m.KeepAlive = 30
	m.Username = ""
	m.Password = ""
	m.CleanSession = true
	m.WillTopic = ""
	m.WillPayload = ""
	m.WillQoS = 0
	m.WillRetain = false

	this.stream.Out <- m
}

func (this *Client)Publish(topic []byte, payload []byte, qos int, retain bool) {
	m := message.NewPublishMessage()
	m.Topic = topic
	m.Payload = payload
	m.QoS = qos
	m.Retain = retain
	m.Dup = false
	m.PacketId = 1

	this.stream.Out <- m
}

func (this *Client)Subscribe(topic []byte, qos int) {
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
	m.Topics = []byte{topic}
	m.PacketId = 1

	this.stream.Out <- m
}

func (this *Client)Disconnect() {
	m := message.NewDisconnectMessage()

	this.stream.Out <- m
}
