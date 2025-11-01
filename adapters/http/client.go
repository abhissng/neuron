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

// Client interface for abstraction
type HTTPClient interface {
	Do(config *HttpClientManager, body []byte, contentType types.ContentType) ([]byte, error)
}

// Standard HTTP client implementation
type stdHTTPClient struct{}

// Do implements HTTPClient
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
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return io.ReadAll(resp.Body)
}

// FastHTTP client implementation (if available)
type fastHTTPClient struct{}

// Do implements HTTPClient
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

// **Main Function: Do HTTP Request**
func DoRequest[T any](payload any, config *HttpClientManager) result.Result[T] {
	// If you see this message check on all places where logging can be added for proper checks
	config.Log.Info(constant.TransactionMessage, log.Any("url", config.URL))
	defer config.Clear()

	err := helpers.ValidateURL(config.URL)
	if err != nil {
		config.Log.Error(constant.TransactionMessage, log.Any("helpers.ValidateURL", err))
		return result.NewFailure[T](blame.URLValidationFailed(config.URL, err))
	}

	config.URL, err = helpers.ConstructURLWithParams(config.URL, config.QueryParams)
	if err != nil {
		config.Log.Error(constant.TransactionMessage, log.Any("helpers.ConstructURLWithParams", err))
		return result.NewFailure[T](blame.URLConstructionFailed(config.URL, config.QueryParams, err))
	}

	// Create request body (returns []byte now)
	bodyBytes, contentType, err := config.createRequestBody(payload)
	if err != nil {
		config.Log.Error(constant.TransactionMessage, log.Any("config.createRequestBody", err))
		return result.NewFailure[T](blame.CreateRequestBodyFailed(err))
	}

	// Select client implementation
	var client HTTPClient
	if config.UseFastHTTP {
		client = &fastHTTPClient{}
	} else {
		client = &stdHTTPClient{}
	}

	// Log request details before execution
	config.Log.Info(constant.TransactionMessage,
		log.String("method", config.Method),
		log.String("url", config.URL),
		log.Any("headers", config.Headers),
		log.Any("query_params", config.QueryParams),
		log.Duration("timeout", config.Timeout),
		log.Any("body", string(bodyBytes)),
	)

	// Execute request
	responseBody, err := client.Do(config, bodyBytes, contentType)
	if err != nil {
		config.Log.Error(constant.TransactionMessage, log.Any("client.Do", err))
		return result.NewFailure[T](blame.CreateHTTPClientFailed(err))
	}

	// Decode response
	decodedResp, err := decodeResponse[T](responseBody, contentType)
	if err != nil {
		config.Log.Error(constant.TransactionMessage, log.Any("decodeResponse", err))
		return result.NewFailure[T](blame.DecodeResponseFailed(err))
	}

	return result.NewSuccess(&decodedResp)
}

// **Step 2: Create HTTP Client**
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

// **Step 3: Handle Response and Decode**
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
