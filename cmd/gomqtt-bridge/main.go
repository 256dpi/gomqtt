package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

func main() {
	// func TestBridge(t *testing.T) {
	o, err := parseFlag()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Parse argument failed:", err)
		return
	}
	if o.help || len(o.conf) == 0 {
		flag.Usage()
		return
	}
	confBytes, err := ioutil.ReadFile(o.conf)
	// confBytes, err := ioutil.ReadFile("conf.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Read configuration file failed:", err)
		return
	}
	conf := new(bridge)
	err = json.Unmarshal(confBytes, conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Load configuration failed:", err)
		return
	}
	localClientConfig := conf.Local.toClientConfig()
	remoteClientConfig := conf.Remote.toClientConfig()
	fmt.Fprintln(os.Stdout, "Local cleint config:", *localClientConfig)
	fmt.Fprintln(os.Stdout, "Remote cleint config:", *remoteClientConfig)

	localClient := client.NewService(10)
	remoteClient := client.NewService(10)
	localClient.Logger = func(s string) {
		fmt.Fprintln(os.Stdout, "[ LOCAL]", s)
	}
	remoteClient.Logger = func(s string) {
		fmt.Fprintln(os.Stdout, "[REMOTE]", s)
	}

	localClientConfig.ProcessPublish = func(p *packet.Publish) error {
		remoteClient.Send(p)
		return nil
	}
	localClientConfig.ProcessPuback = func(p *packet.Puback) error {
		remoteClient.Send(p)
		return nil
	}
	localClientConfig.ProcessPubrec = func(p *packet.Pubrec) error {
		remoteClient.Send(p)
		return nil
	}
	localClientConfig.ProcessPubrel = func(p *packet.Pubrel) error {
		remoteClient.Send(p)
		return nil
	}
	localClientConfig.ProcessPubcomp = func(p *packet.Pubcomp) error {
		remoteClient.Send(p)
		return nil
	}

	remoteClientConfig.ProcessPublish = func(p *packet.Publish) error {
		localClient.Send(p)
		return nil
	}
	remoteClientConfig.ProcessPuback = func(p *packet.Puback) error {
		localClient.Send(p)
		return nil
	}
	remoteClientConfig.ProcessPubrec = func(p *packet.Pubrec) error {
		localClient.Send(p)
		return nil
	}
	remoteClientConfig.ProcessPubrel = func(p *packet.Pubrel) error {
		localClient.Send(p)
		return nil
	}
	remoteClientConfig.ProcessPubcomp = func(p *packet.Pubcomp) error {
		localClient.Send(p)
		return nil
	}

	localClient.Start(localClientConfig)
	remoteClient.Start(remoteClientConfig)
	defer localClient.Stop(true)
	defer remoteClient.Stop(true)

	var lf, rf client.SubscribeFuture
	if len(conf.Local.Subscriptions) != 0 {
		lf = localClient.SubscribeMultiple(toSubscriptions(conf.Local.Subscriptions))
	}
	if len(conf.Remote.Subscriptions) != 0 {
		rf = remoteClient.SubscribeMultiple(toSubscriptions(conf.Remote.Subscriptions))
	}
	if lf != nil {
		lf.Wait(time.Minute)
	}
	if rf != nil {
		rf.Wait(time.Minute)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}
