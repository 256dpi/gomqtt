package broker

import (
	"testing"
	"time"
	"sync/atomic"
	"strings"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

// The MandatoryAcceptanceTest will fully test a Broker with its Backend and
// Session implementation to support all mandatory features. The passed builder
// callback should always return a fresh instances of the Broker.
func MandatoryAcceptanceTest(t *testing.T, builder func() *Broker) {
	t.Log("Running Broker Publish Subscribe Test (QOS 0)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 0, 0)

	t.Log("Running Broker Publish Subscribe Test (QOS 1)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 1, 1)

	t.Log("Running Broker Publish Subscribe Test (QOS 2)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 2, 2)

	t.Log("Running Broker Publish Subscribe Test (Wildcard One)")
	brokerPublishSubscribeTest(t, builder(), "foo/bar", "foo/+", 0, 0)

	t.Log("Running Broker Publish Subscribe Test (Wildcard Some)")
	brokerPublishSubscribeTest(t, builder(), "foo/bar", "#", 0, 0)

	t.Log("Running Broker Publish Subscribe Test (QOS Downgrade 1->0)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 0, 1)

	t.Log("Running Broker Publish Subscribe Test (QOS Downgrade 2->0)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 0, 2)

	t.Log("Running Broker Publish Subscribe Test (QOS Downgrade 2->1)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 1, 2)

	t.Log("Running Broker Unsubscribe Test (QOS 0)")
	brokerUnsubscribeTest(t, builder(), 0)

	t.Log("Running Broker Unsubscribe Test (QOS 1)")
	brokerUnsubscribeTest(t, builder(), 1)

	t.Log("Running Broker Unsubscribe Test (QOS 2)")
	brokerUnsubscribeTest(t, builder(), 2)

	t.Log("Running Broker Subscription Upgrade Test (QOS 0->1)")
	brokerSubscriptionUpgradeTest(t, builder(), 0, 1)

	t.Log("Running Broker Subscription Upgrade Test (QOS 1->2)")
	brokerSubscriptionUpgradeTest(t, builder(), 1, 2)

	t.Log("Running Broker Retained Message Test (QOS 0)")
	brokerRetainedMessageTest(t, builder(), "test", "test", 0, 0)

	t.Log("Running Broker Retained Message Test (QOS 1)")
	brokerRetainedMessageTest(t, builder(), "test", "test", 1, 1)

	t.Log("Running Broker Retained Message Test (QOS 2)")
	brokerRetainedMessageTest(t, builder(), "test", "test", 2, 2)

	t.Log("Running Broker Retained Message Test (Wildcard One)")
	brokerRetainedMessageTest(t, builder(), "foo/bar", "foo/+", 0, 0)

	t.Log("Running Broker Retained Message Test (Wildcard Some)")
	brokerRetainedMessageTest(t, builder(), "foo/bar", "#", 0, 0)

	t.Log("Running Broker Clear Retained Message Test")
	brokerClearRetainedMessageTest(t, builder())

	t.Log("Running Broker Will Test (QOS 0)")
	brokerWillTest(t, builder(), 0, 0)

	t.Log("Running Broker Will Test (QOS 1)")
	brokerWillTest(t, builder(), 1, 1)

	t.Log("Running Broker Will Test (QOS 2)")
	brokerWillTest(t, builder(), 2, 2)

	// TODO: Test Clean Disconnect without forwarding the will.

	t.Log("Running Broker Retained Will Test)")
	brokerRetainedWillTest(t, builder())

	t.Log("Running Broker Keep Alive Test")
	brokerKeepAliveTest(t, builder())

	t.Log("Running Broker Keep Alive Timeout Test")
	brokerKeepAliveTimeoutTest(t, builder())
}

// TODO: Disconnect another client with the same id.
// TODO: Failed Authentication does not disconnect other client with same id.
// TODO: Delivers old Wills in case of a crash.

func runBroker(t *testing.T, broker *Broker, num int) (*tools.Port, chan struct{}) {
	port := tools.NewPort()

	server, err := transport.Launch(port.URL())
	assert.NoError(t, err)

	done := make(chan struct{})

	go func() {
		for i := 0; i < num; i++ {
			conn, err := server.Accept()
			assert.NoError(t, err)

			broker.Handle(conn)
		}

		err := server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	return port, done
}

func errorCallback(t *testing.T) func(*packet.Message, error) {
	return func(msg *packet.Message, err error) {
		if err != nil {
			t.Log(err)
		}

		assert.Fail(t, "callback should not have been called")
	}
}

func brokerPublishSubscribeTest(t *testing.T, broker *Broker, out, in string, sub, pub uint8) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(out, []byte("test"), pub, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerRetainedMessageTest(t *testing.T, broker *Broker, out, in string, sub, pub uint8) {
	port, done := runBroker(t, broker, 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	publishFuture, err := client1.Publish(out, []byte("test"), pub, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerClearRetainedMessageTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 3)

	// client1 retains message

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	publishFuture1, err := client1.Publish("test", []byte("test1"), 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture1.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	// client2 receives retained message and clears it

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test1"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture1, err := client2.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture1.ReturnCodes)

	<-wait

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	// client3 should not receive any message

	client3 := client.New()
	client3.Callback = errorCallback(t)

	connectFuture3, err := client3.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.False(t, connectFuture3.SessionPresent)

	subscribeFuture2, err := client3.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture2.ReturnCodes)

	time.Sleep(50 * time.Millisecond)

	err = client3.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerWillTest(t *testing.T, broker *Broker, sub, pub uint8) {
	port, done := runBroker(t, broker, 2)

	// client1 connets with a will

	client1 := client.New()
	client1.Callback = errorCallback(t)

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     pub,
	}

	connectFuture1, err := client1.Connect(port.URL(), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	// client2 subscribe to the wills topic

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe("test", sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	// client1 dies

	err = client1.Close()
	assert.NoError(t, err)

	// client2 should receive the message

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerRetainedWillTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 2)

	// client1 connects with a retained will and dies

	client1 := client.New()
	client1.Callback = errorCallback(t)

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     0,
		Retain:  true,
	}

	connectFuture1, err := client1.Connect(port.URL(), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	err = client1.Close()
	assert.NoError(t, err)

	// client2 subscribes to the wills topic and receives the retained will

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerUnsubscribeTest(t *testing.T, broker *Broker, qos uint8) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	client.Callback = errorCallback(t)

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client.Unsubscribe("test")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	time.Sleep(50 * time.Millisecond)

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerSubscriptionUpgradeTest(t *testing.T, broker *Broker, from, to uint8) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(to), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture1, err := client.Subscribe("test", from)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())
	assert.Equal(t, []uint8{from}, subscribeFuture1.ReturnCodes)

	subscribeFuture2, err := client.Subscribe("test", to)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())
	assert.Equal(t, []uint8{to}, subscribeFuture2.ReturnCodes)

	publishFuture, err := client.Publish("test", []byte("test"), to, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerKeepAliveTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 1)

	opts := client.NewOptions()
	opts.KeepAlive = "1s"

	client := client.New()
	client.Callback = errorCallback(t)

	var reqCounter int32
	var respCounter int32

	client.Logger = func(message string) {
		if strings.Contains(message, "Pingreq") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "Pingresp") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	connectFuture, err := client.Connect(port.URL(), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	time.Sleep(2500 * time.Millisecond)

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done

	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(2), atomic.LoadInt32(&respCounter))
}

func brokerKeepAliveTimeoutTest(t *testing.T, broker *Broker) {
	connect := packet.NewConnectPacket()
	connect.KeepAlive = 1

	connack := packet.NewConnackPacket()

	client := tools.NewFlow().
		Send(connect).
		Receive(connack).
		End()

	port, done := runBroker(t, broker, 1)

	conn, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	client.Test(t, conn)

	<-done
}
