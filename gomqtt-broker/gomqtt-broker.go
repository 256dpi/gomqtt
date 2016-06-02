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
	"syscall"

	"github.com/gomqtt/broker"
	"github.com/gomqtt/transport"
)

var url = flag.String("url", "tcp://0.0.0.0:1884", "broker url")

func main() {
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// start

	fmt.Printf("Starting broker on url %s... ", *url)

	server, err := transport.Launch(*url)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")

	broker := broker.New()

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				panic(err)
			}

			broker.Handle(conn)
		}
	}()

	// finish

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	<-finish

	fmt.Println("Exiting...")
}
