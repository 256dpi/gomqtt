package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "broker url")
var pairs = flag.Int("pairs", 1, "number of pairs")
var inflight = flag.Int("inflight", 0, "available tokens")
var duration = flag.Int("duration", 0, "duration in seconds")
var length = flag.Int("length", 1, "payload length")
var retained = flag.Bool("retained", false, "retain flag")

var sent int32
var received int32
var delta int32
var total int32

var done = make(chan struct{})
var wg sync.WaitGroup

var payload = bytes.Repeat([]byte{'X'}, *length)

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
	go reporter()

	// await finish
	<-finish
	close(done)

	// wait for exit
	wg.Wait()
}

func connect(id string) transport.Conn {
	// dial broker
	conn, err := transport.Dial(*broker)
	if err != nil {
		panic(err)
	}

	// set may delay
	conn.SetMaxWriteDelay(time.Second)

	// parse url
	mqttURL, err := url.Parse(*broker)
	if err != nil {
		panic(err)
	}

	// prepare connect packet
	connect := packet.NewConnect()
	connect.ClientID = "gomqtt-benchmark/" + id

	// add authentication if available
	if mqttURL.User != nil {
		connect.Username = mqttURL.User.Username()
		pw, _ := mqttURL.User.Password()
		connect.Password = pw
	}

	// send connect
	err = conn.Send(connect, false)
	if err != nil {
		panic(err)
	}

	// receive connack
	pkt, err := conn.Receive()
	if err != nil {
		panic(err)
	}

	// coerce connack
	connack, ok := pkt.(*packet.Connack)
	if !ok {
		panic("expected connack")
	}

	// check return code
	if connack.ReturnCode != packet.ConnectionAccepted {
		panic("expected connection expected")
	}

	return conn
}

func consumer(id string, tokens chan<- struct{}) {
	// ensure exit signal
	defer wg.Done()

	// connect
	conn := connect("consumer/" + id)
	defer conn.Close()

	// ensure close
	go func() {
		<-done
		conn.Close()
	}()

	// prepare subscribe
	subscribe := packet.NewSubscribe()
	subscribe.ID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: id, QOS: 0},
	}

	// send subscribe
	err := conn.Send(subscribe, false)
	if err != nil {
		panic(err)
	}

	// receive suback
	_, err = conn.Receive()
	if err != nil {
		panic(err)
	}

	for {
		// receive publish
		_, err := conn.Receive()
		if err != nil {
			check(err)
			return
		}

		// add token if available
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
	}
}

func publisher(id string, tokens <-chan struct{}) {
	// ensure exit signal
	defer wg.Done()

	// connect
	conn := connect("publisher/" + id)
	defer conn.Close()

	// ensure close
	go func() {
		<-done
		conn.Close()
	}()

	// prepare publish
	publish := packet.NewPublish()
	publish.Message.Topic = id
	publish.Message.Payload = payload
	publish.Message.Retain = *retained

	for {
		// get token if available
		if tokens != nil {
			select {
			case <-tokens:
			case <-done:
				return
			}
		}

		// send publish
		err := conn.Send(publish, true)
		if err != nil {
			check(err)
			return
		}

		// update statistics
		atomic.AddInt32(&sent, 1)
		atomic.AddInt32(&delta, 1)
	}
}

func reporter() {
	// prepare counter
	var iterations int32

	for {
		// wait a second
		time.Sleep(1 * time.Second)

		// get counters
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

func check(err error) {
	select {
	case <-done:
		println(err.Error())
	default:
		panic(err)
	}
}
