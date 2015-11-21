package client

import (
	"net/url"
	"time"
)

type Options struct {
	URL *url.URL
	ClientID string
	CleanSession bool

	WillEnabled bool
	WillTopic string
	WillPayload []byte
	WillQos byte
	WillRetained bool

	KeepAlive time.Duration
}
