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

	// parse url
	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	// get host and port
	host, port, err := net.SplitHostPort(urlParts.Host)
	if err != nil {
		host = urlParts.Host
		port = ""
	}

	// check scheme
	switch urlParts.Scheme {
	case "tcp", "mqtt":
		// set default port
		if port == "" {
			port = d.DefaultTCPPort
		}

		// make connection
		conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, err
		}

		return NewNetConn(conn, d.MaxWriteDelay), nil
	case "tls", "ssl", "mqtts":
		// set default port
		if port == "" {
			port = d.DefaultTLSPort
		}

		// make connection
		conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), d.TLSConfig)
		if err != nil {
			return nil, err
		}

		return NewNetConn(conn, d.MaxWriteDelay), nil
	case "ws":
		// set default port
		if port == "" {
			port = d.DefaultWSPort
		}

		// format url
		wsURL := fmt.Sprintf("ws://%s:%s%s", host, port, urlParts.Path)

		// make connection
		conn, _, err := d.webSocketDialer.Dial(wsURL, d.RequestHeader)
		if err != nil {
			return nil, err
		}

		return NewWebSocketConn(conn, d.MaxWriteDelay), nil
	case "wss":
		// set default port
		if port == "" {
			port = d.DefaultWSSPort
		}

		// format url
		wsURL := fmt.Sprintf("wss://%s:%s%s", host, port, urlParts.Path)

		// make connection
		d.webSocketDialer.TLSClientConfig = d.TLSConfig
		conn, _, err := d.webSocketDialer.Dial(wsURL, d.RequestHeader)
		if err != nil {
			return nil, err
		}

		return NewWebSocketConn(conn, d.MaxWriteDelay), nil
	}

	return nil, ErrUnsupportedProtocol
}
