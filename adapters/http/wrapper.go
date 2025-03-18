package http

import (
	"net/http"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
)

// HttpClientWrapper holds the HTTP request configurations
type HttpClientWrapper struct {
	URL         string
	Method      string
	Headers     map[string]string
	QueryParams map[string]any
	Timeout     time.Duration
	ContentType types.ContentType
	Files       map[string]string // Field name -> File path (for multipart/form-data)
	FormValues  map[string]string // Key-Value for form-urlencoded or multipart
	IsTLS       bool
	CertFile    string
	KeyFile     string
	SkipVerify  bool
	Log         *log.Log
	UseFastHTTP bool // New flag to enable fastHTTP
}

// NewHttpClientWrapper initializes a new HttpClientWrapper with default values
func NewHttpClientWrapper(requestURL string, opts ...RequestOption) *HttpClientWrapper {
	// **Default HttpClientWrapper**
	config := &HttpClientWrapper{
		URL:         requestURL,
		Method:      http.MethodGet,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]any),
		Timeout:     10 * time.Second,
		ContentType: ContentTypeJSON,
		Files:       make(map[string]string),
		FormValues:  make(map[string]string),
		SkipVerify:  false,
		UseFastHTTP: false,
	}

	for _, opt := range opts {
		opt(config)
	}
	if config.Log == nil {
		config.Log = log.NewBasicLogger(helpers.IsProdEnvironment())
		config.Log.Warn("Logger not provided, using default logger")
	}
	return config
}
