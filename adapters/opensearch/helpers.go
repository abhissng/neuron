package opensearch

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

const (
	// DefaultBatchSize is the default number of log entries to buffer before sending.
	DefaultBatchSize = 100
	// DefaultFlushInterval is the default interval for flushing logs, regardless of batch size.
	DefaultFlushInterval = 5 * time.Second
)

// TLSOptions holds TLS configuration for the OpenSearch client.
type TLSOptions struct {
	// CACert is the CA certificate for server verification
	CACert []byte
	// ClientCert is the client certificate for mutual TLS
	ClientCert []byte
	// ClientKey is the client private key for mutual TLS
	ClientKey []byte
	// InsecureSkipVerify controls whether to skip server certificate verification
	InsecureSkipVerify bool
}

// Options holds configuration for the OpenSearch writer.
type Options struct {
	BatchSize    int
	FlushTimeout time.Duration
	TLS          *TLSOptions
	Disable      bool
}

// Option defines a function type to modify options.
type Option func(*Options)

// WithBatchSize sets the batch size for the writer.
func WithBatchSize(size int) Option {
	return func(o *Options) {
		o.BatchSize = size
	}
}

// WithFlushTimeout sets the flush timeout for the writer.
func WithFlushTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.FlushTimeout = timeout
	}
}

// WithTLSConfig configures TLS for the OpenSearch client.
func WithTLSConfig(caCert, clientCert, clientKey []byte, insecureSkipVerify bool) Option {
	return func(o *Options) {
		o.TLS = &TLSOptions{
			CACert:             caCert,
			ClientCert:         clientCert,
			ClientKey:          clientKey,
			InsecureSkipVerify: insecureSkipVerify,
		}
	}
}

// WithInsecureTLS enables insecure TLS (for development only).
func WithInsecureTLS() Option {
	return func(o *Options) {
		o.TLS = &TLSOptions{
			InsecureSkipVerify: true,
		}
	}
}

func WithDisableOpenSearch() Option {
	return func(o *Options) {
		o.Disable = true
	}
}

// NewClient creates a new OpenSearch client with the given options.
func NewClient(addresses []string, username, password string, opts ...Option) (*opensearchapi.Client, error) {
	// Apply default options
	options := &Options{
		BatchSize:    DefaultBatchSize,
		FlushTimeout: DefaultFlushInterval,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	if options.Disable {
		return nil, errors.New(constant.OpenSearchDisabledError.String())
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: options.TLS != nil && options.TLS.InsecureSkipVerify,
	}

	// Load CA certificate if provided
	if options.TLS != nil && len(options.TLS.CACert) > 0 {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(options.TLS.CACert) {
			return nil, errors.New("failed to add CA certificate to pool")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided
	if options.TLS != nil && len(options.TLS.ClientCert) > 0 && len(options.TLS.ClientKey) > 0 {
		cert, err := tls.X509KeyPair(options.TLS.ClientCert, options.TLS.ClientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	config := opensearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	// Create OpenSearch client with TLS config
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: config,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewOpenSearchWriter creates a new OpenSearchWriter instance with the given options.
func NewOpenSearchWriter(client *opensearchapi.Client, indexName string, opts ...Option) (*OpenSearchWriter, error) {
	// Apply default options
	options := &Options{
		BatchSize:    DefaultBatchSize,
		FlushTimeout: DefaultFlushInterval,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	if options.Disable {
		return nil, errors.New(constant.OpenSearchDisabledError.String())
	}

	return &OpenSearchWriter{
		client:       client,
		indexName:    indexName,
		logChannel:   make(chan []byte, options.BatchSize),
		doneChannel:  make(chan struct{}),
		batchSize:    options.BatchSize,
		flushTimeout: options.FlushTimeout,
	}, nil
}
