package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector is a struct for collecting Prometheus metrics.
type MetricsCollector struct {
	registry             *prometheus.Registry
	requestCount         *prometheus.CounterVec
	requestDuration      *prometheus.HistogramVec
	responseSize         *prometheus.HistogramVec
	serviceName          string
	httpRequestsInFlight prometheus.Gauge
	customMetrics        map[string]prometheus.Collector
}

// NewMetricsCollector creates a new Prometheus metrics collector with options.
func NewMetricsCollector(options ...MetricsCollectorOptions) *MetricsCollector {
	registry := prometheus.NewRegistry()
	collector := &MetricsCollector{
		registry:      registry,
		customMetrics: make(map[string]prometheus.Collector),
	}

	// Apply options
	for _, option := range options {
		option(collector)
	}

	// Register default metrics
	collector.registerDefaultMetrics()

	return collector
}

func (mc *MetricsCollector) registerDefaultMetrics() {
	mc.requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: mc.serviceName + "_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status_code"},
	)

	mc.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    mc.serviceName + "_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status_code"},
	)

	mc.responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    mc.serviceName + "_http_response_size_bytes",
			Help:    "Size of HTTP responses",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"service", "method", "path", "status_code"},
	)

	mc.httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: mc.serviceName + "_http_requests_in_flight",
			Help: "Current number of HTTP requests in flight",
		},
	)

	mc.registry.MustRegister(
		mc.requestCount,
		mc.requestDuration,
		mc.responseSize,
		mc.httpRequestsInFlight,
	)
}

// AddCustomMetric adds a custom metric to the collector
func (mc *MetricsCollector) AddCustomMetric(name string, metric prometheus.Collector) {
	mc.customMetrics[name] = metric
	mc.registry.MustRegister(metric)
}

// GetCounter creates a new counter metric
func (mc *MetricsCollector) GetCounter(name, help string, labels []string) prometheus.Counter {
	counter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: mc.serviceName + "_" + name,
			Help: help,
		},
	)
	mc.AddCustomMetric(name, counter)
	return counter
}

// GetGauge creates a new gauge metric
func (mc *MetricsCollector) GetGauge(name, help string, labels []string) prometheus.Gauge {
	gauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: mc.serviceName + "_" + name,
			Help: help,
		},
	)
	mc.AddCustomMetric(name, gauge)
	return gauge
}

// GetHistogram creates a new histogram metric
func (mc *MetricsCollector) GetHistogram(name, help string, buckets []float64, labels []string) prometheus.Histogram {
	histogram := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    mc.serviceName + "_" + name,
			Help:    help,
			Buckets: buckets,
		},
	)
	mc.AddCustomMetric(name, histogram)
	return histogram
}

/*
Todo Usage delete later


// Initialize Gin
	router := gin.Default()

	// Initialize Prometheus collector
	metrics := prometheuswrapper.NewMetricsCollector("my_service")

	// Add middleware
	router.Use(metrics.GinMiddleware())

	// Register metrics endpoint
	metrics.RegisterMetricsEndpoint(router)

	// Add custom metric
	errorCounter := metrics.GetCounter("errors_total", "Total number of errors", nil)

	// Example route
	router.GET("/api/data", func(c *gin.Context) {
		// Your business logic
		errorCounter.Inc()
		c.JSON(200, gin.H{"status": "ok"})
	})


	// digfferent logic
	// Register a new counter metric
	requestCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests",
		},
	)
	prometheus.MustRegister(requestCounter)

	// Register a new histogram metric for request durations
	requestDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
	)
	prometheus.MustRegister(requestDuration)

	// Middleware to instrument Gin requests
	gin.DefaultWriter = iowriter.Discard // Disable Gin's default logger
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		requestCounter.Inc()
		requestDuration.Observe(duration.Seconds())
	})

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())


*/
