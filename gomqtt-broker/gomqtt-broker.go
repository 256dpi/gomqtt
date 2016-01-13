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
	"os"
	"log"
	"runtime/pprof"
	"os/signal"
	"syscall"
	"fmt"

	"github.com/gomqtt/transport"
	"github.com/gomqtt/broker"
)

var url = flag.String("url", "tcp://0.0.0.0:1883", "broker url")

var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")
var memProfile = flag.String("memprofile", "", "write memory profile to this file")

func main() {
	fmt.Println("starting...")

	flag.Parse()

	if *cpuProfile != "" {
		fmt.Println("start cpuprofile!")
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// start

	fmt.Println("starting broker on url '" + *url + "' ...")

	server, err := transport.Launch(*url)
	if err != nil {
		panic(err)
	}

	fmt.Println("launched server!")

	memoryBackend := broker.NewMemoryBackend()
	broker := broker.New(memoryBackend, nil, nil)

	go func(){
		for {
			conn, err := server.Accept()
			if err != nil {
				panic(err)
			}

			broker.Handle(conn)
		}
	}()

	fmt.Println("started broker!")

	// finish

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	<-finish

	if *memProfile != "" {
		fmt.Println("write memprofile!")
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}

	fmt.Println("exiting...")
}
