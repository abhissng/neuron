package prometheus

import "github.com/prometheus/client_golang/prometheus"

// MetricsCollectorOptions defines the options for configuring MetricsCollector.
type MetricsCollectorOptions func(*MetricsCollector)

// WithServiceName sets the service name for the metrics collector.
func WithServiceName(serviceName string) MetricsCollectorOptions {
	return func(collector *MetricsCollector) {
		collector.serviceName = serviceName
	}
}

// WithRegistry sets the Prometheus registry for the metrics collector.
func WithRegistry(registry *prometheus.Registry) MetricsCollectorOptions {
	return func(collector *MetricsCollector) {
		collector.registry = registry
	}
}

// WithCustomMetrics sets custom metrics for the metrics collector.
func WithCustomMetrics(customMetrics map[string]prometheus.Collector) MetricsCollectorOptions {
	return func(collector *MetricsCollector) {
		collector.customMetrics = customMetrics
	}
}

// ServiceName returns the service name.
func (collector *MetricsCollector) ServiceName() string {
	return collector.serviceName
}

// SetServiceName sets the service name.
func (collector *MetricsCollector) SetServiceName(serviceName string) {
	collector.serviceName = serviceName
}

// Registry returns the Prometheus registry.
func (collector *MetricsCollector) Registry() *prometheus.Registry {
	return collector.registry
}

// SetRegistry sets the Prometheus registry.
func (collector *MetricsCollector) SetRegistry(registry *prometheus.Registry) {
	collector.registry = registry
}

// CustomMetrics returns the custom metrics.
func (collector *MetricsCollector) CustomMetrics() map[string]prometheus.Collector {
	return collector.customMetrics
}

// SetCustomMetrics sets the custom metrics.
func (collector *MetricsCollector) SetCustomMetrics(customMetrics map[string]prometheus.Collector) {
	collector.customMetrics = customMetrics
}

// HttpRequestsInFlight returns the gauge metric for the number of HTTP requests in flight.
func (collector *MetricsCollector) HttpRequestsInFlight() prometheus.Gauge {
	return collector.httpRequestsInFlight
}

// SetHttpRequestsInFlight sets the gauge metric for the number of HTTP requests in flight.
func (collector *MetricsCollector) SetHttpRequestsInFlight(httpRequestsInFlight prometheus.Gauge) {
	collector.httpRequestsInFlight = httpRequestsInFlight
}

// RequestCount returns the counter metric for the number of HTTP requests.
func (collector *MetricsCollector) RequestCount() *prometheus.CounterVec {
	return collector.requestCount
}

// SetRequestCount sets the counter metric for the number of HTTP requests.
func (collector *MetricsCollector) SetRequestCount(requestCount *prometheus.CounterVec) {
	collector.requestCount = requestCount
}

// RequestDuration returns the histogram metric for the duration of HTTP requests.
func (collector *MetricsCollector) RequestDuration() *prometheus.HistogramVec {
	return collector.requestDuration
}

// SetRequestDuration sets the histogram metric for the duration of HTTP requests.
func (collector *MetricsCollector) SetRequestDuration(requestDuration *prometheus.HistogramVec) {
	collector.requestDuration = requestDuration
}

// ResponseSize returns the histogram metric for the size of HTTP responses.
func (collector *MetricsCollector) ResponseSize() *prometheus.HistogramVec {
	return collector.responseSize
}

// SetResponseSize sets the histogram metric for the size of HTTP responses.
func (collector *MetricsCollector) SetResponseSize(responseSize *prometheus.HistogramVec) {
	collector.responseSize = responseSize
}
