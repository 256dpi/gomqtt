#!/usr/bin/env bash

mkdir -p ./profile

echo "building gomqtt-broker for linux..."
env GOOS=linux GOARCH=amd64 go build -o ./profile/gomqtt-broker ./gomqtt-broker/gomqtt-broker.go

echo "running gomqtt-broker..."
vagrant ssh -c '/vagrant/profile/gomqtt-broker -cpuprofile /vagrant/profile/cpu.prof -memprofile /vagrant/profile/mem.prof'

