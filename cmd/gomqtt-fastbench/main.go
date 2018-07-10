package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"

	"github.com/juju/ratelimit"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "broker url")
var pairs = flag.Int("pairs", 1, "number of pairs")
var duration = flag.Int("duration", 0, "duration in seconds")
var sendRate = flag.Int("send-rate", 0, "messages per second")
var receiveRate = flag.Int("receive-rate", 0, "messages per second")
var payloadSize = flag.Int("payload", 1, "payload size in bytes")

var sent int32
var received int32
var delta int32
var total int32

var wg sync.WaitGroup

var payload []byte

func main() {
	flag.Parse()

	payload = make([]byte, *payloadSize)

	fmt.Printf("Start benchmark of %s using %d pairs for %d seconds...\n", *broker, *pairs, *duration)

	go func() {
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

		<-finish
		fmt.Println("Closing...")
		os.Exit(0)
	}()

	if int(*duration) > 0 {
		time.AfterFunc(time.Duration(*duration)*time.Second, func() {
			fmt.Println("Finishing...")
			os.Exit(0)
		})
	}

	wg.Add(*pairs * 2)

	for i := 0; i < *pairs; i++ {
		id := strconv.Itoa(i)

		go consumer(id)
		go publisher(id)
	}

	go reporter()

	wg.Wait()
}

func connection(id string) transport.Conn {
	conn, err := transport.Dial(*broker)
	if err != nil {
		panic(err)
	}

	mqttURL, err := url.Parse(*broker)
	if err != nil {
		panic(err)
	}

	connect := packet.NewConnectPacket()
	connect.ClientID = "gomqtt-benchmark/" + id

	if mqttURL.User != nil {
		connect.Username = mqttURL.User.Username()
		pw, _ := mqttURL.User.Password()
		connect.Password = pw
	}

	err = conn.Send(connect)
	if err != nil {
		panic(err)
	}

	pkt, err := conn.Receive()
	if err != nil {
		panic(err)
	}

	if connack, ok := pkt.(*packet.ConnackPacket); ok {
		if connack.ReturnCode == packet.ConnectionAccepted {
			fmt.Printf("Connected: %s\n", id)

			return conn
		}
	}

	panic("connection failed")
}

func consumer(id string) {
	name := "consumer/" + id
	conn := connection(name)

	subscribe := packet.NewSubscribePacket()
	subscribe.ID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: id, QOS: 0},
	}

	err := conn.Send(subscribe)
	if err != nil {
		panic(err)
	}

	var bucket *ratelimit.Bucket
	if *receiveRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*receiveRate), int64(*receiveRate))
	}

	for {
		if bucket != nil {
			bucket.Wait(1)
		}

		_, err := conn.Receive()
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&received, 1)
		atomic.AddInt32(&delta, -1)
		atomic.AddInt32(&total, 1)
	}
}

func publisher(id string) {
	name := "publisher/" + id
	conn := connection(name)

	publish := packet.NewPublishPacket()
	publish.Message.Topic = id
	publish.Message.Payload = payload

	var bucket *ratelimit.Bucket
	if *sendRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*sendRate), int64(*sendRate))
	}

	for {
		if bucket != nil {
			bucket.Wait(1)
		}

		err := conn.BufferedSend(publish)
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&sent, 1)
		atomic.AddInt32(&delta, 1)
	}
}

func reporter() {
	var iterations int32

	for {
		time.Sleep(1 * time.Second)

		curSent := atomic.LoadInt32(&sent)
		curReceived := atomic.LoadInt32(&received)
		curDelta := atomic.LoadInt32(&delta)
		curTotal := atomic.LoadInt32(&total)

		iterations++

		fmt.Printf("Sent: %d msgs - ", curSent)
		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Buffered: %d msgs) ", curDelta)
		fmt.Printf("(Average Throughput: %d msg/s)\n", curTotal/iterations)

		atomic.StoreInt32(&sent, 0)
		atomic.StoreInt32(&received, 0)
	}
}
