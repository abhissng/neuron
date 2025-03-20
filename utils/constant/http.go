package constant

import (
	"github.com/abhissng/neuron/utils/types"
)

// These are headers constant for the application
const (
	CorrelationIDHeader = "X-Correlation-ID"
	XSignature          = "X-Signature"
	XPasetoToken        = "X-Paseto-Token" // #nosec G101
	XRefreshToken       = "X-Refresh-Token"
	XSubject            = "X-Subject"
	AuthorizationHeader = "Authorization"
	IPHeader            = "X-IP"
	MessageIdHeader     = "Message-ID"
	ErrorHeader         = "X-Error"
	CSRFTokenHeader     = "X-CSRF-Token" // #nosec G101
	CSRFTokenCookie     = "X-CSRF-Token" // #nosec G101
)

// These are middlewares or plugin constant for the application
const (
	CorsAllowedOriginsKey        = "CorsAllowedOrigins"
	RedisAddrKey                 = "RedisAddr"
	RedisPasswordKey             = "RedisPassword"
	RedisDBKey                   = "RedisDB"
	RateLimitDefaultKey          = "RateLimitDefault"
	RateLimitSpecialKey          = "RateLimitSpecial"
	RateLimitDurationInSecondKey = "RateLimitDurationInSecond"
)

// These are group version constants for the server routes
const (
	Version1Group = "/v1"
	Version2Group = "/v2"
)

// These are protocol constants
const (
	TCP types.Protocol = "tcp"
	UDP types.Protocol = "udp"
)
