package main

import (
	"github.com/gomqtt/server"
	"github.com/gomqtt/broker"
)

func main() {
	m := broker.NewMemoryBackend()

	b := broker.NewBroker()
	b.QueueBackend = m
//	b.WillBackend = m
//	b.RetainedBackend = m

	s := server.NewServer(b.Handle)
	s.LaunchTCPConfiguration("localhost:1885")

	select{}
}
