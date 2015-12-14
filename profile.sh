#!/usr/bin/env bash

mkdir -p ./profile

echo "building gomqttb for linux..."
env GOOS=linux GOARCH=amd64 go build -o ./profile/gomqttb ./gomqttb/gomqttb.go

echo "running gomqttb..."
vagrant ssh -c '/vagrant/profile/gomqttb -cpuprofile /vagrant/profile/cpu.prof'

echo "generating callgraph..."
go tool pprof --pdf ./profile/gomqttb ./profile/cpu.prof > ./profile/callgraph.pdf
open ./profile/callgraph.pdf
