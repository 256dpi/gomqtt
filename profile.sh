#!/usr/bin/env bash

mkdir -p ./profile

echo "building gomqtt-broker for linux..."
env GOOS=linux GOARCH=amd64 go build -o ./profile/gomqtt-broker ./gomqtt-broker/gomqtt-broker.go

echo "running gomqtt-broker..."
vagrant ssh -c '/vagrant/profile/gomqtt-broker -cpuprofile /vagrant/profile/cpu.prof -memprofile /vagrant/profile/mem.prof'

echo "generating cpu callgraph..."
go tool pprof --pdf ./profile/gomqtt-broker ./profile/cpu.prof > ./profile/cpu.pdf
open ./profile/cpu.pdf

echo "generating mem callgraph..."
go tool pprof --pdf ./profile/gomqtt-broker ./profile/mem.prof > ./profile/mem.pdf
open ./profile/mem.pdf
