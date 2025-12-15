package nats

import "time"

const (
	BreakerName             = "NATSRequest"
	DefaultReconnectWait    = 5 * time.Second
	DefaultMaxReconnects    = -1 // Infinite reconnection attempts
	ConnectionFailedMessage = "connection to NATS is not yet established or failed"
)
