#!/usr/bin/env bash

echo 'ensuring root cert...'
mkcert -install

echo 'generating test cert...'
mkcert example.com localhost 127.0.0.1

echo 'formatting...'
go fmt ./...

echo 'vetting...'
go vet ./...

echo 'linting...'
golint ./...

echo 'testing...'
go test ./...
