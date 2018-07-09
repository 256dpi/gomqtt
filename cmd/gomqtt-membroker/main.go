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

	"github.com/256dpi/gomqtt/broker"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

var url = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var sqz = flag.Int("sqz", 100, "session queue size")

func main() {
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Printf("Starting broker on URL %s... ", *url)

	server, err := transport.Launch(*url)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")

	backend := broker.NewMemoryBackend()
	backend.SessionQueueSize = *sqz

	engine := broker.NewEngine(backend)
	engine.Accept(server)

	var published int32
	var forwarded int32

	engine.Logger = func(event broker.LogEvent, client *broker.Client, pkt packet.GenericPacket, msg *packet.Message, err error) {
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

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	engine.OnError = func(err error) {
		fmt.Println(err.Error())
		finish <- nil
	}

	<-finish

	backend.Close(5 * time.Second)

	server.Close()

	engine.Close()

	fmt.Println("Bye!")
}
