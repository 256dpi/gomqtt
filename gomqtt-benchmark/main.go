// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"github.com/juju/ratelimit"
)

var urlString = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var workers = flag.Int("workers", 1, "number of workers")
var duration = flag.Int("duration", 30, "duration in seconds")
var publishRate = flag.Int("publish-rate", 0, "messages per second")
var receiveRate = flag.Int("receive-rate", 0, "messages per second")

var sent int32
var received int32
var delta int32
var total int32

var wg sync.WaitGroup

func main() {
	flag.Parse()

	fmt.Printf("Start benchmark of %s using %d workers for %d seconds.\n", *urlString, *workers, *duration)

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

	wg.Add(*workers * 2)

	for i := 0; i < *workers; i++ {
		id := strconv.Itoa(i)

		go consumer(id)
		go publisher(id)
	}

	go reporter()

	wg.Wait()
}

func connection(id string) transport.Conn {
	conn, err := transport.Dial(*urlString)
	if err != nil {
		panic(err)
	}

	mqttUrl, err := url.Parse(*urlString)
	if err != nil {
		panic(err)
	}

	connect := packet.NewConnectPacket()
	connect.ClientID = "gomqtt-benchmark/" + id

	if mqttUrl.User != nil {
		connect.Username = mqttUrl.User.Username()
		pw, _ := mqttUrl.User.Password()
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
	subscribe.PacketID = 1
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
	publish.Message.Payload = []byte("foo")

	var bucket *ratelimit.Bucket
	if *publishRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*publishRate), int64(*publishRate))
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

		_sent := atomic.LoadInt32(&sent)
		_received := atomic.LoadInt32(&received)
		_delta := atomic.LoadInt32(&delta)
		_total := atomic.LoadInt32(&total)

		iterations++

		fmt.Printf("Sent: %d msgs - ", _sent)
		fmt.Printf("Received: %d msgs ", _received)
		fmt.Printf("(Buffered: %d msgs) ", _delta)
		fmt.Printf("(Average Throughput: %d msg/s)\n", _total/iterations)

		atomic.StoreInt32(&sent, 0)
		atomic.StoreInt32(&received, 0)
	}
}
