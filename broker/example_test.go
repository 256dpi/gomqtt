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

package broker

import (
	"fmt"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
)

func Example() {
	server, err := transport.Launch("tcp://localhost:8080")
	if err != nil {
		panic(err)
	}

	engine := NewEngine()
	engine.Accept(server)

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) {
		if err != nil {
			panic(err)
		}

		fmt.Println(msg.String())
		close(wait)
	}

	cf, err := c.Connect(client.NewConfig("tcp://localhost:8080"))
	if err != nil {
		panic(err)
	}

	cf.Wait()

	sf, err := c.Subscribe("test", 0)
	if err != nil {
		panic(err)
	}

	sf.Wait()

	pf, err := c.Publish("test", []byte("test"), 0, false)
	if err != nil {
		panic(err)
	}

	pf.Wait()

	<-wait

	err = c.Disconnect()
	if err != nil {
		panic(err)
	}

	err = server.Close()
	if err != nil {
		panic(err)
	}

	engine.Close()

	// Output:
	// <Message Topic="test" QOS=0 Retain=false Payload=[116 101 115 116]>
}
