package http

import (
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
)

// RequestOption defines a functional option for configuring the HTTP client
type RequestOption func(*HttpClientWrapper)

// WithMethod sets the HTTP method
func WithMethod(method string) RequestOption {
	return func(c *HttpClientWrapper) {
		c.Method = method
	}
}

// **✅ WithURL - Validates & Sets URL**
func WithURL(requestURL string) RequestOption {
	return func(c *HttpClientWrapper) {
		err := helpers.ValidateURL(requestURL)
		if err != nil {
			return
		}
		c.URL = requestURL
	}
}

// **✅ WithQueryParams - Adds Query Parameters**
func WithQueryParams(params map[string]any) RequestOption {
	return func(c *HttpClientWrapper) {
		for key, value := range params {
			c.QueryParams[key] = value
		}
	}
}

// WithHeader sets a custom header
func WithHeader(key, value string) RequestOption {
	return func(c *HttpClientWrapper) {
		c.Headers[key] = value
	}
}

// WithTimeout sets a timeout
func WithTimeout(duration time.Duration) RequestOption {
	return func(c *HttpClientWrapper) {
		c.Timeout = duration
	}
}

// WithContentType sets the content type
func WithContentType(contentType types.ContentType) RequestOption {
	return func(c *HttpClientWrapper) {
		c.ContentType = contentType
	}
}

// WithFile adds a file to a multipart/form-data request
func WithFile(fieldName, filePath string) RequestOption {
	return func(c *HttpClientWrapper) {
		c.Files[fieldName] = filePath
	}
}

// WithFormValue adds form data (for x-www-form-urlencoded or multipart)
func WithFormValue(key, value string) RequestOption {
	return func(c *HttpClientWrapper) {
		c.FormValues[key] = value
	}
}

// WithTLSConfig sets up TLS options
func WithTLSConfig(certFile, keyFile string, skipVerify bool) RequestOption {
	return func(c *HttpClientWrapper) {
		c.IsTLS = true
		c.CertFile = certFile
		c.KeyFile = keyFile
		c.SkipVerify = skipVerify
	}
}

// WithLogger sets the log
func WithLogger(log *log.Log) RequestOption {
	return func(c *HttpClientWrapper) {
		c.Log = log
	}
}

// WithFastHTTP sets the flag to use fastHTTP
func WithFastHTTP() RequestOption {
	return func(c *HttpClientWrapper) {
		c.UseFastHTTP = true
	}
}
