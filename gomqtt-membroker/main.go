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
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gomqtt/broker"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
)

var url = flag.String("url", "tcp://0.0.0.0:1883", "broker url")

func main() {
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// start

	fmt.Printf("Starting broker on URL %s... ", *url)

	server, err := transport.Launch(*url)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")

	engine := broker.NewEngine()
	engine.Accept(server)

	var published int32
	var forwarded int32

	engine.Logger = func(event broker.LogEvent, client *broker.Client, pkt packet.Packet, msg *packet.Message, err error) {
		if event == broker.MessagePublished {
			atomic.AddInt32(&published, 1)
		} else if event == broker.MessageForwarded {
			atomic.AddInt32(&forwarded, 1)
		}
	}

	go func() {
		for {
			<-time.After(1 * time.Second)

			pub := atomic.LoadInt32(&published)
			fwd := atomic.LoadInt32(&forwarded)
			fmt.Printf("Publish Rate: %d msg/s, Forward Rate: %d msg/s\n", pub, fwd)

			atomic.StoreInt32(&published, 0)
			atomic.StoreInt32(&forwarded, 0)
		}
	}()

	// finish

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	<-finish

	server.Close()

	engine.Close()
	engine.Wait(1 * time.Second)

	fmt.Println("Bye!")
}
