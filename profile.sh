#!/usr/bin/env bash

mkdir -p ./profile

echo "building gomqttb for linux..."
env GOOS=linux GOARCH=amd64 go build -o ./profile/gomqttb ./gomqttb/gomqttb.go

echo "running gomqttb..."
vagrant ssh -c '/vagrant/profile/gomqttb -cpuprofile /vagrant/profile/cpu.prof -memprofile /vagrant/profile/mem.prof'

echo "generating cpu callgraph..."
go tool pprof --pdf ./profile/gomqttb ./profile/cpu.prof > ./profile/cpu.pdf
open ./profile/cpu.pdf

echo "generating mem callgraph..."
go tool pprof --pdf ./profile/gomqttb ./profile/mem.prof > ./profile/mem.pdf
open ./profile/mem.pdf
