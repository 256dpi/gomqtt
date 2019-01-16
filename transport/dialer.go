package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// The Dialer handles connecting to a server and creating a connection.
type Dialer struct {
	TLSConfig     *tls.Config
	RequestHeader http.Header
	MaxWriteDelay time.Duration

	DefaultTCPPort string
	DefaultTLSPort string
	DefaultWSPort  string
	DefaultWSSPort string

	webSocketDialer *websocket.Dialer
}

// NewDialer returns a new Dialer.
func NewDialer() *Dialer {
	return &Dialer{
		DefaultTCPPort: "1883",
		DefaultTLSPort: "8883",
		DefaultWSPort:  "80",
		DefaultWSSPort: "443",
		webSocketDialer: &websocket.Dialer{
			Proxy:        http.ProxyFromEnvironment,
			Subprotocols: []string{"mqtt"},
		},
	}
}

var sharedDialer *Dialer

func init() {
	sharedDialer = NewDialer()
}

// Dial is a shorthand function.
func Dial(urlString string) (Conn, error) {
	return sharedDialer.Dial(urlString)
}

// Dial initiates a connection based in information extracted from an URL.
func (d *Dialer) Dial(urlString string) (Conn, error) {
	// ensure write delay default
	if d.MaxWriteDelay == 0 {
		d.MaxWriteDelay = 10 * time.Millisecond
	}

	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(urlParts.Host)
	if err != nil {
		host = urlParts.Host
		port = ""
	}

	switch urlParts.Scheme {
	case "tcp", "mqtt":
		if port == "" {
			port = d.DefaultTCPPort
		}

		conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, err
		}

		return NewNetConn(conn, d.MaxWriteDelay), nil
	case "tls", "mqtts":
		if port == "" {
			port = d.DefaultTLSPort
		}

		conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), d.TLSConfig)
		if err != nil {
			return nil, err
		}

		return NewNetConn(conn, d.MaxWriteDelay), nil
	case "ws":
		if port == "" {
			port = d.DefaultWSPort
		}

		wsURL := fmt.Sprintf("ws://%s:%s%s", host, port, urlParts.Path)

		conn, _, err := d.webSocketDialer.Dial(wsURL, d.RequestHeader)
		if err != nil {
			return nil, err
		}

		return NewWebSocketConn(conn, d.MaxWriteDelay), nil
	case "wss":
		if port == "" {
			port = d.DefaultWSSPort
		}

		wsURL := fmt.Sprintf("wss://%s:%s%s", host, port, urlParts.Path)

		d.webSocketDialer.TLSClientConfig = d.TLSConfig
		conn, _, err := d.webSocketDialer.Dial(wsURL, d.RequestHeader)
		if err != nil {
			return nil, err
		}

		return NewWebSocketConn(conn, d.MaxWriteDelay), nil
	}

	return nil, ErrUnsupportedProtocol
}
