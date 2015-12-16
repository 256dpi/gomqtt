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
	"net"
	"time"
	"fmt"
	"sync/atomic"
	"strconv"

	"github.com/gomqtt/stream"
	"github.com/gomqtt/packet"
)

const interval = 100

var host = flag.String("host", "0.0.0.0", "broker host")
var port = flag.String("port", "1884", "broker port")
var workers = flag.Int("workers", 1, "number of workers")

var sent = make(chan int)
var received = make(chan int)

func main() {
	flag.Parse()

	for i := 0; i < *workers; i++ {
		id := "id-" + strconv.Itoa(i)
		go counter(id)
		go bomber(id)
	}

	go reporter()

	select{}
}

func connection(name string) stream.Stream {
	conn, err := net.Dial("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		panic(err)
	}

	s := stream.NewNetStream(conn)

	cp := packet.NewConnectPacket()
	cp.ClientID = []byte("gomqtt-benchmark/" + name)
	s.Send(cp)

	ap := <-s.Incoming()
	if ap == nil {
		return nil
	} else {
		if _ap, ok := ap.(*packet.ConnackPacket); ok {
			if _ap.ReturnCode != packet.ConnectionAccepted {
				fmt.Println(name + ": connection denied")
				return nil
			}
		}
	}

	return s
}

func counter(id string) {
	s := connection("counter/" + id)
	if s == nil {
		return
	}

	sp := packet.NewSubscribePacket()
	sp.PacketID = 1
	sp.Subscriptions = []packet.Subscription{
		packet.Subscription{
			Topic: []byte(id),
			QOS: 0,
		},
	}
	s.Send(sp)

	i := 0

	for {
		_, ok := <- s.Incoming()

		if ok {
			i++

			if i >= interval {
				received <- i
				i = 0
			}
		}
	}
}

func bomber(id string) {
	s := connection("bomber/" + id)
	if s == nil {
		return
	}

	pp := packet.NewPublishPacket()
	pp.Topic = []byte(id)
	pp.Payload = []byte("foo")

	i := 0

	for {
		s.Send(pp)
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
		fmt.Printf("sent: %d msg/s, received: %d msg/s\n", atomic.LoadInt32(&s) / 5, atomic.LoadInt32(&r) / 5)
		atomic.StoreInt32(&s, 0)
		atomic.StoreInt32(&r, 0)
	}
}
