package middleware

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// GinRequestLogger returns a gin.HandlerFunc that logs incoming HTTP requests and their corresponding responses.
// It records method, URL, client IP, request body, request and correlation IDs, user agent, and headers for each request.
// It also logs response status and request latency, and — when enabled by configuration or when not running in production — captures and logs the response body.
func GinRequestLogger(logger *log.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		logResponseBody := viper.GetBool(constant.ResponseBodyPrint)
		if !logResponseBody && !helpers.IsProdEnvironment() {
			logResponseBody = true
		}
		startTime := time.Now()

		// Log Request Details (body read safely and restored for handlers; sensitive keys masked when logger has sanitizer)
		bodyBytes, _ := helpers.ReadBodySafe(c.Request)
		var bodyForLog any
		if len(bodyBytes) > 0 {
			var m map[string]any
			if err := json.Unmarshal(bodyBytes, &m); err == nil {
				bodyForLog = m
			} else {
				bodyForLog = string(bodyBytes)
			}
		}

		logger.Info("Incoming Request",
			log.String("method", c.Request.Method),
			log.String("url", c.Request.RequestURI),
			log.String("client_ip", c.ClientIP()),
			logger.Any("body", bodyForLog),
			log.String("request_id", c.GetString(constant.RequestID)),
			log.String(constant.CorrelationIDHeader, c.GetString(constant.CorrelationID)),
			log.Any("user_agent", c.Request.UserAgent()),
			logger.Any("headers", c.Request.Header),
		)

		// Capture Response Body (if needed)
		var responseBodyBuffer bytes.Buffer
		if logResponseBody {
			writer := &responseWriter{ResponseWriter: c.Writer, body: &responseBodyBuffer}
			c.Writer = writer
		}

		// Process Request
		c.Next()

		// Calculate Latency
		latency := time.Since(startTime)

		// Log Response Details
		fields := []types.Field{
			log.Int("status_code", c.Writer.Status()),
			log.String("latency", latency.String()),
			log.String("request_id", c.GetString(constant.RequestID)),
			log.String(constant.CorrelationIDHeader, c.GetString(constant.CorrelationID)),
		}

		if logResponseBody {
			fields = append(fields, log.String("response_body", responseBodyBuffer.String()))
		}

		logger.Info("Response Details", fields...)
	}
}

// responseWriter is a custom implementation of gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes the response body
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // Capture response body
	return w.ResponseWriter.Write(b)
}