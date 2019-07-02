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

var wg sync.WaitGroup

var payload []byte

func main() {
	flag.Parse()

	go func() {
		panic(http.ListenAndServe("localhost:6061", nil))
	}()

	payload = make([]byte, *length)

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

		var tokens chan struct{}
		if *inflight > 0 {
			tokens = make(chan struct{}, *inflight)
			for i := 0; i < *inflight; i++ {
				tokens <- struct{}{}
			}
		}

		go consumer(id, tokens)
		go publisher(id, tokens)
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

func consumer(id string, tokens chan<- struct{}) {
	name := "consumer/" + id
	cl := connect(name)

	cl.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		if tokens != nil {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}

		atomic.AddInt32(&received, 1)
		atomic.AddInt32(&delta, -1)
		atomic.AddInt32(&total, 1)

		return nil
	}

	sf, err := cl.Subscribe(id, packet.QOS(*qos))
	if err != nil {
		panic(err)
	}

	err = sf.Wait(5 * time.Second)
	if err != nil {
		panic(err)
	}
}

func publisher(id string, tokens <-chan struct{}) {
	name := "publisher/" + id
	cl := connect(name)

	list := make(chan client.GenericFuture, *futures)

	go func() {
		for {
			if tokens != nil {
				<-tokens
			}

			pf, err := cl.Publish(id, payload, packet.QOS(*qos), *retained)
			if err != nil {
				panic(err)
			}

			list <- pf
		}
	}()

	for pf := range list {
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
