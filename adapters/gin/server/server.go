package server

import (
	"errors"

	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/gin-gonic/gin"
)

// StartServer initializes and starts the server based on the provided options

func StartServer(opts ...ServerOption) error {

	// Merge options into a single ServerOptions instance
	options := DefaultServerOptions()
	for _, opt := range opts {
		opt(options)
	}

	if helpers.IsProdEnvironment() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	//gin.Logger()
	router.Use(gin.Recovery())

	// apply serve static
	applyServeStatic(router, options.serveStatic)

	// apply HTML blob
	applyHTMLBlob(router, options.htmlBlobPath)

	// Apply global middlewares
	applyGlobalMiddlewares(router, options.GlobalMiddlewares)

	// Configure routing:
	// If a custom routing configurator is provided, use it.
	// Otherwise, use the default base URL group and route groups.
	if options.RoutingConfigurator != nil {
		options.RoutingConfigurator(router)
	} else {
		// Set up the base URL group
		baseGroup := router.Group(options.baseURL)

		// Configure route groups
		configureRouteGroups(baseGroup, options.RouteGroups)
	}

	port, err := helpers.GetAvailablePort(constant.TCP, options.port)
	if err != nil {
		helpers.Println(constant.ERROR, blame.ErrorServerStartFailed.String()+"\n"+err.Error())
		return err
	}

	if port != options.port {
		helpers.Println(constant.WARN, "Server will be running on ["+port+"] port, as configured port: "+options.port+" is not available")
	}

	// Start the server
	if err := router.Run(":" + port); err != nil {
		helpers.Println(constant.ERROR, blame.ErrorServerStartFailed.String()+"\n"+err.Error())
		return errors.New(blame.ErrorServerStartFailed.String())
	}
	return nil
}
