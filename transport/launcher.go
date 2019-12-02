package transport

import (
	"crypto/tls"
	"net/url"
)

// The Launcher helps with launching a server and accepting connections.
type Launcher struct {
	TLSConfig *tls.Config
}

// NewLauncher returns a new Launcher.
func NewLauncher() *Launcher {
	return &Launcher{}
}

var sharedLauncher = NewLauncher()

// Launch is a shorthand function.
func Launch(address string) (Server, error) {
	return sharedLauncher.Launch(address)
}

// Launch will launch a server based on information extracted from the address.
func (l *Launcher) Launch(address string) (Server, error) {
	// parse address
	addr, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}

	// check scheme
	switch addr.Scheme {
	case "tcp", "mqtt":
		return CreateNetServer(addr.Host)
	case "tls", "ssl", "mqtts":
		return CreateSecureNetServer(addr.Host, l.TLSConfig)
	case "ws":
		return CreateWebSocketServer(addr.Host)
	case "wss":
		return CreateSecureWebSocketServer(addr.Host, l.TLSConfig)
	}

	return nil, ErrUnsupportedProtocol
}
