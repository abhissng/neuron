package nats

import "time"

const (
	BreakerName             = "NATSRequest"
	DefaultReconnectWait    = 5 * time.Second
	DefautMaxReconnects     = -1 // Infinite reconnection attempts
	ConnectionFailedMessage = "connection to NATS is not yet established or failed"
)
