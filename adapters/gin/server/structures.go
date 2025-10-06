package server

import (
	"os"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-gonic/gin"
)

// ServerOptions encapsulates the configuration for the Gin server
type ServerOptions struct {
	port              string
	baseURL           string
	htmlBlobPath      string
	gracefulTimeOut   time.Duration
	serveStatic       *ServeStaticConfig
	GlobalMiddlewares []gin.HandlerFunc
	RouteGroups       []RouteGroupConfig
	// A custom routing configurator allows complete control over route registration.
	RoutingConfigurator func(*gin.Engine)
	log                 *log.Log
}

// DefaultServerOptions returns the default server options
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		port:              os.Getenv(constant.DefaultAppPort), // Default port
		baseURL:           "/",                                // Default base URL
		htmlBlobPath:      "",
		serveStatic:       &ServeStaticConfig{},
		gracefulTimeOut:   time.Duration(10) * time.Second, // Default TimeOut
		GlobalMiddlewares: []gin.HandlerFunc{},
		RouteGroups:       []RouteGroupConfig{},
	}
}

// ServeStaticConfig encapsulates the configuration for the static content which will be served
type ServeStaticConfig struct {
	relativePath string // example "/static"
	rootPath     string // example "./static"

}

// NewServeStaticConfig creates a new ServeStaticConfig instance
func NewServeStaticConfig(relativePath, rootPath string) *ServeStaticConfig {
	return &ServeStaticConfig{
		relativePath: relativePath,
		rootPath:     rootPath,
	}
}

// RouteGroupConfig defines configuration for a specific route group
type RouteGroupConfig struct {
	Prefix      string
	Middlewares []gin.HandlerFunc
	Routes      []RouteConfig
}

// NewRouteGroupConfig creates a new RouteGroupConfig instance
func NewRouteGroupConfig(
	prefix string,
	middlewares []gin.HandlerFunc,
	routes []RouteConfig) RouteGroupConfig {
	return RouteGroupConfig{
		Prefix:      prefix,
		Middlewares: middlewares,
		Routes:      routes,
	}
}

// RouteConfig defines an individual route
type RouteConfig struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// NewRouteConfig creates a new RouteConfig instance
func NewRouteConfig(method, path string, handler gin.HandlerFunc) RouteConfig {
	return RouteConfig{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}
