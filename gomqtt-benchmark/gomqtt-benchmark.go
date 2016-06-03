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
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
)

const interval = 1000

var url = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var workers = flag.Int("workers", 1, "number of workers")

var sent = make(chan int)
var received = make(chan int)

func main() {
	flag.Parse()

	fmt.Printf("Start benchmark for %s using %d Workers\n", *url, *workers)

	go func() {
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

		<-finish
		fmt.Println("Closing...")
		os.Exit(0)
	}()

	for i := 0; i < *workers; i++ {
		id := strconv.Itoa(i)

		go consumer(id)
		go publisher(id)
	}

	reporter()
}

func connection(id string) transport.Conn {
	conn, err := transport.Dial(*url)
	if err != nil {
		panic(err)
	}

	connect := packet.NewConnectPacket()
	connect.ClientID = "gomqtt-benchmark/" + id

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
	conn := connection("consumer/" + id)

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: id, QOS: 0},
	}

	err := conn.Send(subscribe)
	if err != nil {
		panic(err)
	}

	counter := 0

	for {
		_, err := conn.Receive()
		if err != nil {
			panic(err)
		}

		counter++

		if counter >= interval {
			received <- counter
			counter = 0
		}
	}
}

func publisher(id string) {
	conn := connection("publisher/" + id)

	publish := packet.NewPublishPacket()
	publish.Message.Topic = id
	publish.Message.Payload = []byte("foo")

	counter := 0

	for {
		err := conn.Send(publish)
		if err != nil {
			panic(err)
		}

		counter++

		if counter >= interval {
			sent <- counter
			counter = 0
		}
	}
}

func reporter() {
	var sentCounter int32
	var receivedCounter int32
	var balance int32

	go func() {
		for {
			n := int32(<-sent)
			atomic.AddInt32(&sentCounter, n)
			atomic.AddInt32(&balance, n)
		}
	}()

	go func() {
		for {
			n := int32(<-received)
			atomic.AddInt32(&receivedCounter, n)
			atomic.AddInt32(&balance, -n)
		}
	}()

	var iterations int32
	var totalReceived int32

	for {
		<-time.After(1 * time.Second)

		sentPerSecond := atomic.LoadInt32(&sentCounter)
		receivedPerSecond := atomic.LoadInt32(&receivedCounter)
		currentBalance := atomic.LoadInt32(&balance)

		iterations++
		totalReceived += receivedPerSecond

		fmt.Printf("Sent: %d msgs - ", sentPerSecond)
		fmt.Printf("Received: %d msgs ", receivedPerSecond)
		fmt.Printf("(Buffered: %d msgs) ", currentBalance)
		fmt.Printf("(Average Throughput: %d msg/s)\n", totalReceived/iterations)

		atomic.StoreInt32(&sentCounter, 0)
		atomic.StoreInt32(&receivedCounter, 0)
	}
}
