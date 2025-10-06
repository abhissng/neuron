package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/gin-gonic/gin"
)

// ServerOption defines a functional option for configuring the server

type ServerOption func(*ServerOptions)

// Shutdown gracefully shuts down the server
func (s *ServerOptions) Shutdown(ctx context.DefaultContext) error {
	// Retrieve the value safely using the comma-ok idiom
	logs := s.log
	if logs == nil {
		logs = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
	}
	logs.Info("Gracefully Shutting down server......")

	// Channel to notify when shutdown is complete
	done := make(chan struct{})
	// Create a goroutine to handle the delayed exit
	go func() {
		defer func() {
			// Close Logger safely
			logs.Info(constant.ConnectionClosed, log.Any("message", "Closing logger"))
			_ = logs.Sync()
			close(done)
		}()
		// Wait for 10 seconds before exiting
		select {
		case <-ctx.Done():
			logs.Info("Context canceled, exiting immediately")
		case <-time.After(constant.ServerDefaultGracefulTime):
			logs.Info(constant.ServerDefaultGracefulTime.String() + " seconds elapsed, exiting now")
		}
	}()
	// Wait for shutdown completion
	<-done

	// Allow graceful exit
	os.Exit(0)
	return nil
}

// WithPort sets the server port
func WithPort(port string) ServerOption {
	return func(o *ServerOptions) {
		o.port = port
	}
}

// WithBaseURL sets the base URL for the server
func WithBaseURL(baseURL string) ServerOption {
	return func(o *ServerOptions) {
		o.baseURL = baseURL
	}
}

// WithBaseURL sets the base URL for the server
func WithServeStatic(config *ServeStaticConfig) ServerOption {
	return func(o *ServerOptions) {
		o.serveStatic = config
	}
}

// WithHTMLBlob sets the HTML Blob or template for the server
func WithHTMLBlob(htmlBlobPath string) ServerOption {
	return func(o *ServerOptions) {
		o.htmlBlobPath = htmlBlobPath
	}
}

// WithGlobalMiddleware adds global middleware
func WithGlobalMiddleware(middleware gin.HandlerFunc) ServerOption {
	return func(o *ServerOptions) {
		o.GlobalMiddlewares = append(o.GlobalMiddlewares, middleware)
	}
}

// WithRouteGroup adds a route group
func WithRouteGroup(group RouteGroupConfig) ServerOption {
	return func(o *ServerOptions) {
		o.RouteGroups = append(o.RouteGroups, group)
	}
}

// WithRoutes adds routes directly under the base URL
func WithRoutes(routes []RouteConfig) ServerOption {
	return func(o *ServerOptions) {
		o.RouteGroups = append(o.RouteGroups, RouteGroupConfig{
			Prefix: "", // Empty prefix means routes are added directly under the base URL
			Routes: routes,
		})
	}
}

// WithGracefulTimeOut adds a graceful time out
func WithGracefulTimeOut(timeOut time.Duration) ServerOption {
	return func(o *ServerOptions) {
		o.gracefulTimeOut = timeOut
	}
}

// WithRoutingConfigurator allows you to supply a custom routing function.
// This is useful if your routes need to inject additional middleware in between.
func WithRoutingConfigurator(fn func(*gin.Engine)) ServerOption {
	return func(o *ServerOptions) {
		o.RoutingConfigurator = fn
	}
}

// applyServeStatic applies server static content to the router
func applyServeStatic(router *gin.Engine, config *ServeStaticConfig) {
	if !helpers.IsEmpty(config.relativePath) && !helpers.IsEmpty(config.rootPath) {
		router.Static(config.relativePath, config.rootPath)
	}
}

// applyHTMLBlob applies HTML Blob to the router
func applyHTMLBlob(router *gin.Engine, htmlBlobPath string) {
	if !helpers.IsEmpty(htmlBlobPath) {
		router.LoadHTMLGlob(htmlBlobPath)
	}
}

// applyGlobalMiddlewares applies global middlewares to the router
func applyGlobalMiddlewares(router *gin.Engine, middlewares []gin.HandlerFunc) {
	for _, mw := range middlewares {
		router.Use(mw)
	}
}

// configureRouteGroups configures route groups and their routes
func configureRouteGroups(baseGroup *gin.RouterGroup, routeGroups []RouteGroupConfig) {
	for _, groupConfig := range routeGroups {
		group := baseGroup.Group(groupConfig.Prefix)
		for _, mw := range groupConfig.Middlewares {
			group.Use(mw)
		}

		// Add routes to the group
		for _, route := range groupConfig.Routes {
			switch route.Method {
			case http.MethodGet:
				group.GET(route.Path, route.Handler)
			case http.MethodPost:
				group.POST(route.Path, route.Handler)
			case http.MethodPut:
				group.PUT(route.Path, route.Handler)
			case http.MethodDelete:
				group.DELETE(route.Path, route.Handler)
			case "UPDATE": // Custom `UPDATE` method
				group.Handle("UPDATE", route.Path, route.Handler)
			default:
				helpers.Println(constant.ERROR, fmt.Sprintf("Unsupported method: %s\n", route.Method))
			}
		}
	}
}

// WithLogger sets the logger for the server
func WithLogger(log *log.Log) ServerOption {
	return func(o *ServerOptions) {
		o.log = log
	}
}
