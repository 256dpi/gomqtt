all: fmt vet lint

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golint ./...

install:
	go install ...

cert:
	mkcert -install
	mkcert -cert-file example.crt -key-file example.key example.com localhost 127.0.0.1
