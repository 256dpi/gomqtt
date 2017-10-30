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

var sharedLauncher *Launcher

func init() {
	sharedLauncher = NewLauncher()
}

// Launch is a shorthand function.
func Launch(urlString string) (Server, error) {
	return sharedLauncher.Launch(urlString)
}

// Launch will launch a server based on information extracted from an URL.
func (l *Launcher) Launch(urlString string) (Server, error) {
	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	switch urlParts.Scheme {
	case "tcp", "mqtt":
		return NewNetServer(urlParts.Host)
	case "tls", "mqtts":
		return NewSecureNetServer(urlParts.Host, l.TLSConfig)
	case "ws":
		return NewWebSocketServer(urlParts.Host)
	case "wss":
		return NewSecureWebSocketServer(urlParts.Host, l.TLSConfig)
	}

	return nil, ErrUnsupportedProtocol
}
