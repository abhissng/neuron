package grpcmanager

import (
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/paseto"
	neuronctx "github.com/abhissng/neuron/context"
)

// ServerConfig holds gRPC server configurations
type ServerConfig struct {
	port             int
	certFile         string
	keyFile          string
	caFile           string
	jwtSecret        string
	enableMetrics    bool
	serviceName      string
	maxRecvMsgSize   int
	maxSendMsgSize   int
	log              *log.Log
	authMode         string
	pasetoManager    *paseto.PasetoManager
	appContext       *neuronctx.AppContext
	serviceRegistrar ServiceRegistrar
	customValidator  CustomValidatorFunc
	skipAuthMethods  map[string]bool
}

// Option is a function that modifies ServerConfig
type Option func(*ServerConfig)

// WithPort sets the gRPC server port
func WithPort(port int) Option {
	return func(c *ServerConfig) {
		c.port = port
	}
}

// WithTLS enables TLS with provided cert/key
func WithTLS(certFile, keyFile, caFile string) Option {
	return func(c *ServerConfig) {
		c.certFile = certFile
		c.keyFile = keyFile
		c.caFile = caFile
	}
}

// WithJWT enables authentication using JWT secret
func WithJWT(secret string) Option {
	return func(c *ServerConfig) {
		c.jwtSecret = secret
	}
}

// WithMetrics enables Prometheus monitoring
func WithMetrics() Option {
	return func(c *ServerConfig) {
		c.enableMetrics = true
	}
}

// WithMaxRecvMsgSize sets max received message size (MB)
func WithMaxRecvMsgSize(size int) Option {
	return func(c *ServerConfig) {
		c.maxRecvMsgSize = size
	}
}

// WithMaxSendMsgSize sets max send message size (MB)
func WithMaxSendMsgSize(size int) Option {
	return func(c *ServerConfig) {
		c.maxSendMsgSize = size
	}
}

func WithLogger(log *log.Log) Option {
	return func(c *ServerConfig) {
		c.log = log
	}
}

// WithAuthMode selects auth mode: "jwt" or "paseto". Empty means no auth.
func WithAuthMode(mode string) Option {
	return func(c *ServerConfig) {
		c.authMode = mode
	}
}

// WithPasetoManager provides a PasetoManager when using PASETO auth mode.
func WithPasetoManager(pm *paseto.PasetoManager) Option {
	return func(c *ServerConfig) {
		c.pasetoManager = pm
	}
}

// WithAppContext sets the neuron AppContext for ServiceContext propagation.
func WithAppContext(appCtx *neuronctx.AppContext) Option {
	return func(c *ServerConfig) {
		c.appContext = appCtx
	}
}

// WithServiceRegistrar sets a callback for registering gRPC services.
// This decouples proto implementations from the server library.
func WithServiceRegistrar(registrar ServiceRegistrar) Option {
	return func(c *ServerConfig) {
		c.serviceRegistrar = registrar
	}
}

// WithCustomValidator sets an external validation function.
// This is called after token parsing but before the handler.
func WithCustomValidator(validator CustomValidatorFunc) Option {
	return func(c *ServerConfig) {
		c.customValidator = validator
	}
}

// WithSkipAuthMethods sets methods that should skip authentication.
// Provide full method names like "/package.Service/Method".
func WithSkipAuthMethods(methods ...string) Option {
	return func(c *ServerConfig) {
		if c.skipAuthMethods == nil {
			c.skipAuthMethods = make(map[string]bool)
		}
		for _, m := range methods {
			c.skipAuthMethods[m] = true
		}
	}
}

// WithServiceName sets the service name for logging and metrics.
func WithServiceName(name string) Option {
	return func(c *ServerConfig) {
		c.serviceName = name
	}
}
