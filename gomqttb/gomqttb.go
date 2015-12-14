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
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	fmt.Println("starting...")

	flag.Parse()

	if *cpuprofile != "" {
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
//	b.WillBackend = m
//	b.RetainedBackend = m

	s := server.NewServer(b.Handle)
	s.LaunchTCPConfiguration("0.0.0.0:1884")

	fmt.Println("started!")

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

	<-finish
	fmt.Println("exiting...")
}
