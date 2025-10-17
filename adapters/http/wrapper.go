package http

import (
	"net/http"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
)

// HttpClientManager holds the HTTP request configurations
type HttpClientManager struct {
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

func (c *HttpClientManager) Clear() {
	c.Headers = make(map[string]string)
	c.QueryParams = make(map[string]any)
	c.Files = make(map[string]string)
	c.FormValues = make(map[string]string)
	c.SkipVerify = false
	c.UseFastHTTP = false
}

// NewHttpClientManager initializes a new HttpClientManager with default values
func NewHttpClientManager(requestURL string, opts ...RequestOption) *HttpClientManager {
	// **Default HttpClientManager**
	config := &HttpClientManager{
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
		config.Log = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
		config.Log.Warn("Logger not provided, using default logger")
	}
	return config
}

func (c *HttpClientManager) AddHeaders(headers map[string]string) {
	c.Headers = headers
}

func (c *HttpClientManager) AddQueryParams(params map[string]any) {
	c.QueryParams = params
}

func (c *HttpClientManager) AddFiles(files map[string]string) {
	c.Files = files
}

func (c *HttpClientManager) AddFormValues(values map[string]string) {
	c.FormValues = values
}

func (c *HttpClientManager) AddTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

func (c *HttpClientManager) AddContentType(contentType types.ContentType) {
	c.ContentType = contentType
}

func (c *HttpClientManager) AddIsTLS(isTLS bool) {
	c.IsTLS = isTLS
}

func (c *HttpClientManager) AddCertFile(certFile string) {
	c.CertFile = certFile
}

func (c *HttpClientManager) AddKeyFile(keyFile string) {
	c.KeyFile = keyFile
}

func (c *HttpClientManager) AddSkipVerify(skipVerify bool) {
	c.SkipVerify = skipVerify
}

func (c *HttpClientManager) AddFastHTTP() {
	c.UseFastHTTP = true
}
