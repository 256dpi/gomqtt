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

	"github.com/gomqtt/server"
	"github.com/gomqtt/broker"
	"net"
)

var host = flag.String("host", "0.0.0.0", "broker host")
var port = flag.String("port", "1884", "broker port")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func main() {
	fmt.Println("starting...")

	flag.Parse()

	if *cpuprofile != "" {
		fmt.Println("start cpuprofile!")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// start

	m := broker.NewMemoryBackend()

	b := broker.NewBroker()
	b.QueueBackend = m

	s := server.NewServer(b.Handle)
	s.LaunchTCPConfiguration(net.JoinHostPort(*host, *port))

	fmt.Println("started!")

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	<-finish

	if *memprofile != "" {
		fmt.Println("write memprofile!")
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}

	fmt.Println("exiting...")
}
