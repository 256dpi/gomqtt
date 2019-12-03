package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

var testTLSConfig *tls.Config
var testDialer *Dialer
var testLauncher *Launcher

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	crt, err := tls.LoadX509KeyPair(filepath.Join(wd, "../example.crt"), filepath.Join(wd, "../example.key"))
	if err != nil {
		panic(err)
	}

	testTLSConfig = &tls.Config{
		Certificates: []tls.Certificate{crt},
		NextProtos:   []string{"h2", "http/1.1"},
	}

	testDialer = NewDialer(DialConfig{})

	testLauncher = NewLauncher(LaunchConfig{
		TLSConfig: testTLSConfig,
	})
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

		conn.SetMaxWriteDelay(10 * time.Millisecond)

		handler(conn)

		server.Close()
		close(done)
	}()

	conn, err := testDialer.Dial(getURL(server, protocol))
	if err != nil {
		panic(err)
	}

	conn.SetMaxWriteDelay(10 * time.Millisecond)

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
