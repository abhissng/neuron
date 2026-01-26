package http

import (
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
)

// HttpClientManager holds the HTTP request configurations
type HttpClientManager struct {
	mu          sync.RWMutex
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
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Headers = make(map[string]string)
	c.QueryParams = make(map[string]any)
	c.Files = make(map[string]string)
	c.FormValues = make(map[string]string)
	c.SkipVerify = false
	c.UseFastHTTP = false
}

func (c *HttpClientManager) snapshot() *HttpClientManager {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s := &HttpClientManager{
		URL:         c.URL,
		Method:      c.Method,
		Headers:     make(map[string]string, len(c.Headers)),
		QueryParams: make(map[string]any, len(c.QueryParams)),
		Timeout:     c.Timeout,
		ContentType: c.ContentType,
		Files:       make(map[string]string, len(c.Files)),
		FormValues:  make(map[string]string, len(c.FormValues)),
		IsTLS:       c.IsTLS,
		CertFile:    c.CertFile,
		KeyFile:     c.KeyFile,
		SkipVerify:  c.SkipVerify,
		Log:         c.Log,
		UseFastHTTP: c.UseFastHTTP,
	}
	maps.Copy(s.Headers, c.Headers)
	maps.Copy(s.QueryParams, c.QueryParams)
	maps.Copy(s.Files, c.Files)
	maps.Copy(s.FormValues, c.FormValues)

	return s
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
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Headers = headers
}

func (c *HttpClientManager) ResetHeaders() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Headers = make(map[string]string)
}

func (c *HttpClientManager) AddQueryParams(params map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.QueryParams = params
}

func (c *HttpClientManager) ResetQueryParams() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.QueryParams = make(map[string]any)
}

func (c *HttpClientManager) AddFiles(files map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Files = files
}

func (c *HttpClientManager) ResetFiles() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Files = make(map[string]string)
}

func (c *HttpClientManager) AddFormValues(values map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FormValues = values
}

func (c *HttpClientManager) ResetFormValues() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FormValues = make(map[string]string)
}

func (c *HttpClientManager) AddTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Timeout = timeout
}

func (c *HttpClientManager) ResetTimeout() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Timeout = 10 * time.Second
}

func (c *HttpClientManager) AddContentType(contentType types.ContentType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ContentType = contentType
}

func (c *HttpClientManager) ResetContentType() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ContentType = ContentTypeJSON
}

func (c *HttpClientManager) AddIsTLS(isTLS bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsTLS = isTLS
}

func (c *HttpClientManager) ResetIsTLS() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsTLS = false
}

func (c *HttpClientManager) AddCertFile(certFile string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CertFile = certFile
}

func (c *HttpClientManager) ResetCertFile() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CertFile = ""
}

func (c *HttpClientManager) AddKeyFile(keyFile string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.KeyFile = keyFile
}

func (c *HttpClientManager) ResetKeyFile() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.KeyFile = ""
}

func (c *HttpClientManager) AddSkipVerify(skipVerify bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SkipVerify = skipVerify
}

func (c *HttpClientManager) ResetSkipVerify() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SkipVerify = false
}

func (c *HttpClientManager) AddFastHTTP() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.UseFastHTTP = true
}

func (c *HttpClientManager) ResetFastHTTP() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.UseFastHTTP = false
}
