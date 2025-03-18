package graceful

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// Shutdowner is an interface that defines a Shutdown method.
type Shutdowner interface {
	Shutdown(ctx context.DefaultContext) error
}

// ShutdownFunc is a function type that matches the Shutdown method signature.
type ShutdownFunc func(ctx context.DefaultContext) error

// Shutdown implements the Shutdowner interface for ShutdownFunc.
func (f ShutdownFunc) Shutdown(ctx context.DefaultContext) error {
	return f(ctx)
}

// GracefulShutdown handles graceful shutdown for any type that implements the Shutdowner interface.
// timeout specifies the duration to wait before forcefully shutting down.
func GracefulShutdown(service Shutdowner, timeout time.Duration) {
	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-stop

	// Graceful shutdown
	ctx, cancel := context.DefaultContextWithTimeout(timeout)
	defer cancel()
	if err := service.Shutdown(ctx); err != nil {
		helpers.Println(constant.ERROR, "Error during shutdown: "+err.Error())
	} else {
		helpers.Println(constant.INFO, "Service stopped")
	}
}
