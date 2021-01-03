package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/mercury"

	"github.com/256dpi/gomqtt/broker"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

var url = flag.String("url", "tcp://0.0.0.0:1883", "broker url")
var queue = flag.Int("queue", 100, "queue size")
var publish = flag.Int("publish", 100, "parallel publishes")
var inflight = flag.Int("inflight", 100, "inflight messages")
var timeout = flag.Int("timeout", 1, "token timeout")

func main() {
	// parse flags
	flag.Parse()

	// run pprof interface
	go func() {
		panic(http.ListenAndServe("localhost:6060", nil))
	}()

	// print info
	fmt.Printf("Starting broker on URL %s...\n", *url)

	// launch server
	server, err := transport.Launch(*url)
	if err != nil {
		panic(err)
	}

	// prepare backend
	backend := broker.NewMemoryBackend()
	backend.SessionQueueSize = *queue
	backend.ClientParallelPublishes = *publish
	backend.ClientParallelSubscribes = *inflight
	backend.ClientInflightMessages = *inflight
	backend.ClientTokenTimeout = time.Duration(*timeout) * time.Second

	// prepare counters
	var published int32
	var forwarded int32
	var clients int32

	// configure logger
	backend.Logger = func(event broker.LogEvent, client *broker.Client, pkt packet.Generic, msg *packet.Message, err error) {
		if event == broker.NewConnection {
			atomic.AddInt32(&clients, 1)
		} else if event == broker.MessagePublished {
			atomic.AddInt32(&published, 1)
		} else if event == broker.MessageForwarded {
			atomic.AddInt32(&forwarded, 1)
		} else if event == broker.LostConnection {
			atomic.AddInt32(&clients, -1)
		}
	}

	// prepare engine
	engine := broker.NewEngine(backend)

	// set error callback
	engine.OnError = func(err error) {
		println(err.Error())
	}

	// accept from server
	engine.Accept(server)

	// run reporter
	go func() {
		// prepare mercury stats
		var stats mercury.Stats

		for {
			// wait a second
			time.Sleep(time.Second)

			// read counters
			pub := atomic.LoadInt32(&published)
			fwd := atomic.LoadInt32(&forwarded)

			// get stats
			s := mercury.GetStats()
			d := s.Sub(stats)
			stats = s

			// get goroutines
			n := runtime.NumGoroutine()

			// print statistics
			fmt.Printf("Publish Rate: %d msg/s, Forward Rate: %d msg/s, Clients: %d, Mercury: %d/%d, Goroutines: %d\n", pub, fwd, clients, d.Executed, d.Skipped, n)

			// reset counters
			atomic.StoreInt32(&published, 0)
			atomic.StoreInt32(&forwarded, 0)
		}
	}()

	// await finish signal
	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish

	// close backend
	backend.Close(5 * time.Second)

	// close server
	_ = server.Close()

	// close engine
	engine.Close()
}
