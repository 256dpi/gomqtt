// Package implements session objects to be used with MQTT clients.
package session

// Direction denotes a packets direction.
type Direction int

const (
	// Incoming packets are being received.
	Incoming Direction = iota

	// Outgoing packets are being be sent.
	Outgoing
)
