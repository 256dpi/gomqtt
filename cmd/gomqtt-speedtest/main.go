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
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/beorn7/perks/quantile"
)

var urlString = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var topic = flag.String("topic", "speedtest", "the used topic")
var qos = flag.Uint("qos", 0, "the qos level")
var wait = flag.Int("wait", 0, "time to wait in milliseconds")

var received = make(chan time.Time)

func main() {
	flag.Parse()

	cl := client.New()

	cl.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		received <- time.Now()
		return nil
	}

	cf, err := cl.Connect(client.NewConfig(*urlString))
	if err != nil {
		panic(err)
	}

	err = cf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	sf, err := cl.Subscribe(*topic, uint8(*qos))
	if err != nil {
		panic(err)
	}

	err = sf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	q := quantile.NewTargeted(map[float64]float64{
		0.50: 0.005,
		0.90: 0.001,
		0.99: 0.0001,
	})

	for {
		t1 := time.Now()

		pf, err := cl.Publish(*topic, []byte("speedtest"), uint8(*qos), false)
		if err != nil {
			panic(err)
		}

		err = pf.Wait(10 * time.Second)
		if err != nil {
			panic(err)
		}

		t2 := <-received

		q.Insert(float64(t2.Sub(t1).Nanoseconds() / 1000 / 1000))

		q50 := time.Duration(q.Query(0.50)) * time.Millisecond
		q90 := time.Duration(q.Query(0.90)) * time.Millisecond
		q99 := time.Duration(q.Query(0.99)) * time.Millisecond

		fmt.Printf("[%d] 0.50: %s, 0.90: %s, 0.99: %s \n", q.Count(), q50, q90, q99)

		time.Sleep(time.Duration(*wait) * time.Millisecond)
	}
}
