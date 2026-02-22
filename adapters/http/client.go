package http

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/valyala/fasthttp"
)

// HTTPClient defines the interface for HTTP client implementations.
// It abstracts the HTTP client to support both standard and FastHTTP clients.
type HTTPClient interface {
	Do(config *HttpClientManager, body []byte, contentType types.ContentType) ([]byte, error)
}

// stdHTTPClient implements HTTPClient using the standard net/http package.
// It provides reliable HTTP client functionality with full HTTP/1.1 and HTTP/2 support.
type stdHTTPClient struct{}

// Do executes an HTTP request using the standard net/http client.
// It handles request creation, header setting, and response reading.
func (c *stdHTTPClient) Do(config *HttpClientManager, body []byte, contentType types.ContentType) ([]byte, error) {
	req, err := http.NewRequest(config.Method, config.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType.String())
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	client := config.createHTTPClient()
	//#nosec G704
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return io.ReadAll(resp.Body)
}

// fastHTTPClient implements HTTPClient using the valyala/fasthttp package.
// It provides high-performance HTTP client functionality for speed-critical applications.
type fastHTTPClient struct{}

// Do executes an HTTP request using the FastHTTP client.
// It provides better performance than standard HTTP client for high-throughput scenarios.
func (c *fastHTTPClient) Do(config *HttpClientManager, body []byte, contentType types.ContentType) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(config.URL)
	req.Header.SetMethod(config.Method)
	req.Header.SetContentType(contentType.String())
	req.SetBody(body)

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	client := &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: config.SkipVerify, // #nosec G402
		},
	}

	if config.IsTLS && config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, err
		}
		client.TLSConfig.Certificates = append(client.TLSConfig.Certificates, cert)
	}

	err := client.Do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

// DoRequest executes a complete HTTP request with automatic encoding/decoding.
// It validates URLs, constructs query parameters, handles TLS, and decodes responses.
func DoRequest[T any](payload any, config *HttpClientManager) result.Result[T] {
	// If you see this message check on all places where logging can be added for proper checks
	snap := config.snapshot()
	snap.Log.Info(constant.TransactionMessage, log.Any("url", snap.URL))
	defer config.Clear()

	err := helpers.ValidateURL(snap.URL)
	if err != nil {
		snap.Log.Error(constant.TransactionMessage, log.Any("helpers.ValidateURL", err))
		return result.NewFailure[T](blame.URLValidationFailed(snap.URL, err))
	}

	snap.URL, err = helpers.ConstructURLWithParams(snap.URL, snap.QueryParams)
	if err != nil {
		snap.Log.Error(constant.TransactionMessage, log.Any("helpers.ConstructURLWithParams", err))
		return result.NewFailure[T](blame.URLConstructionFailed(snap.URL, snap.QueryParams, err))
	}

	// Create request body (returns []byte now)
	bodyBytes, contentType, err := snap.createRequestBody(payload)
	if err != nil {
		snap.Log.Error(constant.TransactionMessage, log.Any("config.createRequestBody", err))
		return result.NewFailure[T](blame.CreateRequestBodyFailed(err))
	}

	// Select client implementation
	var client HTTPClient
	if snap.UseFastHTTP {
		client = &fastHTTPClient{}
	} else {
		client = &stdHTTPClient{}
	}

	// Log request details before execution
	snap.Log.Info(constant.TransactionMessage,
		log.String("method", snap.Method),
		log.String("url", snap.URL),
		snap.Log.SanitizeAny("headers", snap.Headers),
		snap.Log.SanitizeAny("query_params", snap.QueryParams),
		log.Duration("timeout", snap.Timeout),
		snap.Log.SanitizeAny("body", string(bodyBytes)),
	)

	// Execute request
	responseBody, err := client.Do(snap, bodyBytes, contentType)
	if err != nil {
		snap.Log.Error(constant.TransactionMessage, log.Any("client.Do", err))
		return result.NewFailure[T](blame.CreateHTTPClientFailed(err))
	}

	// Decode response
	decodedResp, err := decodeResponse[T](responseBody, contentType)
	if err != nil {
		snap.Log.Error(constant.TransactionMessage, log.Any("decodeResponse", err))
		return result.NewFailure[T](blame.DecodeResponseFailed(err))
	}

	return result.NewSuccess(&decodedResp)
}

// createHTTPClient creates a configured HTTP client with TLS support.
// It handles timeout settings, TLS configuration, and client certificates.
func (config *HttpClientManager) createHTTPClient() *http.Client {
	client := &http.Client{Timeout: config.Timeout}

	if config.IsTLS {
		tlsConfig := &tls.Config{InsecureSkipVerify: config.SkipVerify} // #nosec G402
		if config.CertFile != "" && config.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
			if err == nil {
				config.Log.Info(constant.TransactionMessage, log.Any("tls.LoadX509KeyPair", cert))
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}
		client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	return client
}

// decodeResponse decodes HTTP response body based on content type.
// It attempts content-type specific decoding first, then falls back to generic decoders.
func decodeResponse[T any](body []byte, contentType types.ContentType) (T, error) {
	var result T
	reader := bytes.NewReader(body)

	// Decode based on Content-Type
	result, err := decodeByContentType[T](contentType, reader)
	if err == nil {
		return result, nil
	}

	// Use fallback decoders
	return fallbackDecode[T](reader)
}
