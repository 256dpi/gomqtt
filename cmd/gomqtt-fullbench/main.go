package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "broker url")
var pairs = flag.Int("pairs", 1, "number of pairs")
var inflight = flag.Int("inflight", 0, "number of inflight messages")
var futures = flag.Int("futures", 100, "number of active futures")
var duration = flag.Int("duration", 0, "duration in seconds")
var length = flag.Int("length", 1, "message payload length")
var retained = flag.Bool("retained", false, "message retain flag")
var qos = flag.Int("qos", 0, "message qos")

var sent int32
var received int32
var delta int32
var total int32

var done = make(chan struct{})
var wg sync.WaitGroup

var payload = make([]byte, *length)

func main() {
	// parse flags
	flag.Parse()

	// run pprof interface
	go func() {
		panic(http.ListenAndServe("localhost:6061", nil))
	}()

	// print info
	fmt.Printf("Start benchmark of %s using %d pairs for %d seconds...\n", *broker, *pairs, *duration)

	// prepare finish signal
	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	// exit after duration if available
	if *duration > 0 {
		time.AfterFunc(time.Duration(*duration)*time.Second, func() {
			select {
			case finish <- syscall.SIGTERM:
			default:
			}
		})
	}

	// launch pairs
	for i := 0; i < *pairs; i++ {
		// compute id
		id := strconv.Itoa(i)

		// prepare tokens
		var tokens chan struct{}
		if *inflight > 0 {
			tokens = make(chan struct{}, *inflight)
			for i := 0; i < *inflight; i++ {
				tokens <- struct{}{}
			}
		}

		// launch consumer and publisher
		wg.Add(2)
		go consumer(id, tokens)
		go publisher(id, tokens)
	}

	// launch reporter
	wg.Add(1)
	go reporter()

	// await finish
	<-finish
	close(done)

	// wait for exit
	wg.Wait()
}

func connect(id string) *client.Client {
	// create client
	cl := client.New()

	// prepare config
	cfg := client.NewConfigWithClientID(*broker, "gomqtt-benchmark/"+id)

	// connect client
	cf, err := cl.Connect(cfg)
	if err != nil {
		panic(err)
	}

	// await connect
	err = cf.Wait(5 * time.Second)
	if err != nil {
		panic(err)
	}

	return cl
}

func consumer(id string, tokens chan<- struct{}) {
	// ensure exit signal
	defer wg.Done()

	// connect
	consumer := connect("consumer/" + id)
	defer consumer.Close()

	// set callback
	consumer.Callback = func(msg *packet.Message, err error) error {
		// check error
		if err != nil {
			panic(err)
		}

		// add token
		if tokens != nil {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}

		// update statistics
		atomic.AddInt32(&received, 1)
		atomic.AddInt32(&delta, -1)
		atomic.AddInt32(&total, 1)

		return nil
	}

	// subscribe topic
	sf, err := consumer.Subscribe(id, packet.QOS(*qos))
	if err != nil {
		panic(err)
	}

	// await subscription
	err = sf.Wait(5 * time.Second)
	if err != nil {
		panic(err)
	}

	// await finish
	<-done
}

func publisher(id string, tokens <-chan struct{}) {
	// ensure exit signal
	defer wg.Done()

	// connect
	publisher := connect("publisher/" + id)
	defer publisher.Close()

	// prepare future queue
	list := make(chan client.GenericFuture, *futures)

	// run publisher
	go func() {
		for {
			// get token if available
			if tokens != nil {
				select {
				case <-tokens:
				case <-done:
					close(list)
					return
				}
			}

			// publish message
			future, err := publisher.Publish(id, payload, packet.QOS(*qos), *retained)
			if err != nil {
				panic(err)
			}

			// queue future
			select {
			case list <- future:
			case <-done:
				close(list)
				return
			}
		}
	}()

	// handle futures
	for future := range list {
		// await result
		err := future.Wait(5 * time.Second)
		if err != nil {
			panic(err)
		}

		// update statistics
		atomic.AddInt32(&sent, 1)
		atomic.AddInt32(&delta, 1)
	}
}

func reporter() {
	// ensure exit signal
	defer wg.Done()

	// count iterations
	var iterations int32

	for {
		// wait a second
		select {
		case <-time.After(time.Second):
		case <-done:
			return
		}

		// load values
		curSent := atomic.LoadInt32(&sent)
		curReceived := atomic.LoadInt32(&received)
		curDelta := atomic.LoadInt32(&delta)
		curTotal := atomic.LoadInt32(&total)

		// increment
		iterations++

		// print statistics
		fmt.Printf("Sent: %d msgs - ", curSent)
		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Buffered: %d msgs) ", curDelta)
		fmt.Printf("(Average Throughput: %d msg/s)\n", curTotal/iterations)

		// reset counters
		atomic.StoreInt32(&sent, 0)
		atomic.StoreInt32(&received, 0)
	}
}
