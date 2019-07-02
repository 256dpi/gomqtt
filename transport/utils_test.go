package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

var serverTLSConfig *tls.Config

var testDialer *Dialer
var testLauncher *Launcher

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	crt, err := tls.LoadX509KeyPair(filepath.Join(wd, "../example.com+2.pem"), filepath.Join(wd, "../example.com+2-key.pem"))
	if err != nil {
		panic(err)
	}

	serverTLSConfig = &tls.Config{
		Certificates: []tls.Certificate{crt},
		NextProtos:   []string{"h2", "http/1.1"},
	}

	testDialer = NewDialer()

	testLauncher = NewLauncher()
	testLauncher.TLSConfig = serverTLSConfig
}

// returns a client-ish and server-ish pair of connections
func connectionPair(protocol string, handler func(Conn)) (Conn, chan struct{}) {
	done := make(chan struct{})

	server, err := testLauncher.Launch(protocol + "://localhost:0")
	if err != nil {
		panic(err)
	}

	go func() {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}

		handler(conn)

		server.Close()
		close(done)
	}()

	conn, err := testDialer.Dial(getURL(server, protocol))
	if err != nil {
		panic(err)
	}

	return conn, done
}

func getPort(s Server) string {
	_, port, _ := net.SplitHostPort(s.Addr().String())
	return port
}

func getURL(s Server, protocol string) string {
	return fmt.Sprintf("%s://%s", protocol, s.Addr().String())
}

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(1 * time.Minute):
		panic("nothing received")
	case <-ch:
	}
}
