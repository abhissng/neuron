package middleware

import (
	"strconv"
	"time"

	"github.com/abhissng/neuron/adapters/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	MetricsEndpoint = "/metrics"
)

// GinMiddleware returns a Gin middleware for collecting metrics
func GinMiddleware(mc *prometheus.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		mc.HttpRequestsInFlight().Inc()
		defer mc.HttpRequestsInFlight().Dec()

		c.Next()

		statusCode := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()
		responseSize := float64(c.Writer.Size())

		mc.RequestCount().WithLabelValues(
			mc.ServiceName(),
			c.Request.Method,
			path,
			statusCode,
		).Inc()

		mc.RequestDuration().WithLabelValues(
			mc.ServiceName(),
			c.Request.Method,
			path,
			statusCode,
		).Observe(duration)

		mc.ResponseSize().WithLabelValues(
			mc.ServiceName(),
			c.Request.Method,
			path,
			statusCode,
		).Observe(responseSize)
	}
}

// RegisterMetricsEndpoint registers the Prometheus metrics endpoint
func RegisterMetricsEndpoint(router *gin.Engine, mc *prometheus.MetricsCollector) {
	router.GET(MetricsEndpoint, gin.WrapH(promhttp.HandlerFor(
		mc.Registry(),
		promhttp.HandlerOpts{},
	)))
}
