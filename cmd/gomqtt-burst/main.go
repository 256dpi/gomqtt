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

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

var urlString = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var topic = flag.String("topic", "speedtest", "the used topic")
var qos = flag.Uint("qos", 0, "the qos level")
var amount = flag.Int("n", 100, "the amount of messages")

func main() {
	flag.Parse()

	cl := client.New()

	cl.Callback = func(msg *packet.Message, err error) {
		if err != nil {
			panic(err)
		}
	}

	cf, err := cl.Connect(client.NewConfig(*urlString))
	if err != nil {
		panic(err)
	}

	err = cf.Wait()
	if err != nil {
		panic(err)
	}

	var futures []*client.PublishFuture

	for i := 0; i < *amount; i++ {
		pf, err := cl.Publish(*topic, []byte("burst"), uint8(*qos), false)
		if err != nil {
			panic(err)
		}

		futures = append(futures, pf)
	}

	for _, pf := range futures {
		err = pf.Wait()
		if err != nil {
			panic(err)
		}
	}

	err = cl.Disconnect()
	if err != nil {
		panic(err)
	}
}
