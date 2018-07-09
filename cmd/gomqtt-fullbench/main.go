package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/juju/ratelimit"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "broker url")
var pairs = flag.Int("pairs", 1, "number of pairs")
var duration = flag.Int("duration", 30, "duration in seconds")
var publishRate = flag.Int("publish-rate", 0, "messages per second")
var processRate = flag.Int("process-rate", 0, "messages per second")
var payloadSize = flag.Int("payload", 1, "message payload size")
var qos = flag.Int("qos", 0, "message qos")
var retained = flag.Bool("retained", false, "message retain flag")
var inflight = flag.Int("inflight", 10, "number of inflight messages")

var sent int32
var received int32
var delta int32
var total int32

var wg sync.WaitGroup

var payload []byte

func main() {
	flag.Parse()

	payload = make([]byte, *payloadSize)

	fmt.Printf("Start benchmark of %s using %d pairs for %d seconds...\n", *broker, *pairs, *duration)

	go func() {
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

		<-finish
		fmt.Println("Closing...")
		os.Exit(0)
	}()

	if int(*duration) > 0 {
		time.AfterFunc(time.Duration(*duration)*time.Second, func() {
			fmt.Println("Finishing...")
			os.Exit(0)
		})
	}

	wg.Add(*pairs * 2)

	for i := 0; i < *pairs; i++ {
		id := strconv.Itoa(i)

		go consumer(id)
		go publisher(id)
	}

	go reporter()

	wg.Wait()
}

func connect(id string) *client.Client {
	cl := client.New()

	cfg := client.NewConfig(*broker)
	cfg.ClientID = "gomqtt-benchmark/" + id

	cf, err := cl.Connect(cfg)
	if err != nil {
		panic(err)
	}

	err = cf.Wait(5 * time.Second)
	if err != nil {
		panic(err)
	}

	return cl
}

func consumer(id string) {
	name := "consumer/" + id
	cl := connect(name)

	var bucket *ratelimit.Bucket
	if *processRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*processRate), int64(*processRate))
	}

	cl.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		if bucket != nil {
			bucket.Wait(1)
		}

		atomic.AddInt32(&received, 1)
		atomic.AddInt32(&delta, -1)
		atomic.AddInt32(&total, 1)

		return nil
	}

	sf, err := cl.Subscribe(id, byte(*qos))
	if err != nil {
		panic(err)
	}

	err = sf.Wait(5 * time.Second)
	if err != nil {
		panic(err)
	}
}

func publisher(id string) {
	name := "publisher/" + id
	cl := connect(name)

	var bucket *ratelimit.Bucket
	if *publishRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*publishRate), int64(*publishRate))
	}

	futures := make(chan client.GenericFuture, *inflight)

	go func() {
		for {
			if bucket != nil {
				bucket.Wait(1)
			}

			pf, err := cl.Publish(id, payload, byte(*qos), *retained)
			if err != nil {
				panic(err)
			}

			futures <- pf
		}
	}()

	for pf := range futures {
		err := pf.Wait(5 * time.Second)
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&sent, 1)
		atomic.AddInt32(&delta, 1)
	}
}

func reporter() {
	var iterations int32

	for {
		time.Sleep(1 * time.Second)

		curSent := atomic.LoadInt32(&sent)
		curReceived := atomic.LoadInt32(&received)
		curDelta := atomic.LoadInt32(&delta)
		curTotal := atomic.LoadInt32(&total)

		iterations++

		fmt.Printf("Sent: %d msgs - ", curSent)
		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Buffered: %d msgs) ", curDelta)
		fmt.Printf("(Average Throughput: %d msg/s)\n", curTotal/iterations)

		atomic.StoreInt32(&sent, 0)
		atomic.StoreInt32(&received, 0)
	}
}
