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

func connection(name string) transport.Conn {
	conn, err := transport.Dial(*url)
	if err != nil {
		panic(err)
	}

	cp := packet.NewConnectPacket()
	cp.ClientID = []byte("gomqtt-benchmark/" + name)

	err = conn.Send(cp)
	if err != nil {
		panic(err)
	}

	pkt, err := conn.Receive()
	if err != nil {
		panic(err)
	}

	if ap, ok := pkt.(*packet.ConnackPacket); ok {
		if ap.ReturnCode == packet.ConnectionAccepted {
			return conn
		}
	}

	panic(name + ": connection failed")
}

func counter(id string) {
	s := connection("counter/" + id)

	sp := packet.NewSubscribePacket()
	sp.PacketID = 1
	sp.Subscriptions = []packet.Subscription{
		packet.Subscription{
			Topic: []byte(id),
			QOS: 0,
		},
	}

	err := s.Send(sp)
	if err != nil {
		panic(err)
	}

	i := 0

	for {
		_, err := s.Receive()
		if err != nil {
			panic(err)
		}

		i++

		if i >= interval {
			received <- i
			i = 0
		}
	}
}

func bomber(id string) {
	s := connection("bomber/" + id)

	pp := packet.NewPublishPacket()
	pp.Topic = []byte(id)
	pp.Payload = []byte("foo")

	i := 0

	for {
		err := s.Send(pp)
		if err != nil {
			panic(err)
		}

		i++

		if i >= interval {
			sent <- i
			i = 0
		}
	}
}

func reporter() {
	var s int32 = 0
	var r int32 = 0

	go func(){
		for {
			atomic.AddInt32(&s, int32(<-sent))
		}
	}()

	go func(){
		for {
			atomic.AddInt32(&r, int32(<-received))
		}
	}()

	for {
		<-time.After(5 * time.Second)

		fmt.Printf("sent: %d msg/s, ", atomic.LoadInt32(&s) / 5)
		fmt.Printf("received: %d msg/s, ", atomic.LoadInt32(&r) / 5)
		fmt.Printf("diff: %d\n", (atomic.LoadInt32(&s) / 5) - (atomic.LoadInt32(&r) / 5))

		atomic.StoreInt32(&s, 0)
		atomic.StoreInt32(&r, 0)
	}
}
