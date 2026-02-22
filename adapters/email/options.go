package email

import (
	"crypto/tls"

	"github.com/abhissng/neuron/adapters/log"
)

// ClientOptions holds the configuration for the email client
type ClientOptions struct {
	Type      string
	Host      string
	Port      int
	Username  string
	Password  string //#nosec
	TLSConfig *tls.Config
	log       *log.Log
}

type Option func(*ClientOptions)

// WithHost sets the host for the email client
func WithHost(host string) Option {
	return func(o *ClientOptions) {
		o.Host = host
	}
}

// WithPort sets the port for the email client
func WithPort(port int) Option {
	return func(o *ClientOptions) {
		o.Port = port
	}
}

// WithCredentials sets the username and password for the email client
// @param username: The username or email for the email client
// @param password: The password for the email client
func WithCredentials(username, password string) Option {
	return func(o *ClientOptions) {
		o.Username = username
		o.Password = password
	}
}

// WithTLSConfig sets the TLS configuration for the email client
// @param config: The TLS configuration for the email client
func WithTLSConfig(config *tls.Config) Option {
	return func(o *ClientOptions) {
		o.TLSConfig = config
	}
}

// WithClientType sets the client type for the email client
// @param clientType: The client type for the email client
func WithClientType(clientType string) Option {
	return func(o *ClientOptions) {
		o.Type = clientType
	}
}

// WithLog sets the log for the email client
func WithLog(log *log.Log) Option {
	return func(o *ClientOptions) {
		o.log = log
	}
}
