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
	"time"
	"fmt"
	"sync/atomic"
	"strconv"

	"github.com/gomqtt/transport"
	"github.com/gomqtt/packet"
)

const interval = 100
const update = 5

var url = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var workers = flag.Int("workers", 1, "number of workers")

var sent = make(chan int)
var received = make(chan int)

func main() {
	flag.Parse()

	fmt.Println("start benchmark of '" + *url +"' with " + strconv.Itoa(*workers) + " workers")

	for i := 0; i < *workers; i++ {
		id := "id-" + strconv.Itoa(i)

		go counter(id)
		go bomber(id)
	}

	reporter()
}

func connection(id string) transport.Conn {
	conn, err := transport.Dial(*url)
	if err != nil {
		panic(err)
	}

	connect := packet.NewConnectPacket()
	connect.ClientID = []byte("gomqtt-benchmark/" + id)

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
			fmt.Println(id + ": connected")

			return conn
		}
	}

	panic(id + ": connection failed")
}

func counter(id string) {
	conn := connection("counter/" + id)

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		packet.Subscription{
			Topic: []byte(id),
			QOS: 0,
		},
	}

	err := conn.Send(subscribe)
	if err != nil {
		panic(err)
	}

	fmt.Println(id + ": subscribed to '" + id + "'")

	counter:= 0

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

func bomber(id string) {
	conn := connection("bomber/" + id)

	publish := packet.NewPublishPacket()
	publish.Topic = []byte(id)
	publish.Payload = []byte("foo")

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
	var sentCounter int32 = 0
	var receivedCounter int32 = 0

	go func(){
		for {
			atomic.AddInt32(&sentCounter, int32(<-sent))
		}
	}()

	go func(){
		for {
			atomic.AddInt32(&receivedCounter, int32(<-received))
		}
	}()

	for {
		<-time.After(update * time.Second)

		sentPerSecond := atomic.LoadInt32(&sentCounter) / update
		receivedPerSecond := atomic.LoadInt32(&receivedCounter) / update

		fmt.Printf("sent: %d msg/s, ", sentPerSecond)
		fmt.Printf("received: %d msg/s, ", receivedPerSecond)
		fmt.Printf("diff: %d\n", sentPerSecond - receivedPerSecond)

		atomic.StoreInt32(&sentCounter, 0)
		atomic.StoreInt32(&receivedCounter, 0)
	}
}
