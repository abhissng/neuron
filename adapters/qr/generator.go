package qr

import (
	"context"
	"sync"
)

// Generator is the public abstraction used by services to produce QR artifacts.
// Implementations should be safe for concurrent use.
type Generator interface {
	Generate(ctx context.Context, req Request) (*Result, error)
}

// Provider is the pluggable backend contract.
type Provider interface {
	Generate(ctx context.Context, req Request) (*Result, error)
}

type generator struct {
	cfg generatorConfig
}

var (
	defaultProviderMu      sync.RWMutex
	defaultProviderFactory func() Provider
)

// RegisterDefaultProvider registers a process-wide default provider factory.
// Intended to be called by provider adapter packages during init.
func RegisterDefaultProvider(factory func() Provider) {
	if factory == nil {
		return
	}
	defaultProviderMu.Lock()
	defer defaultProviderMu.Unlock()
	if defaultProviderFactory == nil {
		defaultProviderFactory = factory
	}
}

// DefaultGenerator builds a generator using registered defaults.
func DefaultGenerator(options ...GeneratorOption) (Generator, error) {
	return NewGenerator(options...)
}

// NewGenerator creates a generator with sensible defaults and an injected provider.
func NewGenerator(options ...GeneratorOption) (Generator, error) {
	cfg := defaultGeneratorConfig()
	for _, opt := range options {
		if opt != nil {
			opt(&cfg)
		}
	}

	if cfg.provider == nil {
		defaultProviderMu.RLock()
		factory := defaultProviderFactory
		defaultProviderMu.RUnlock()
		if factory != nil {
			cfg.provider = factory()
		}
	}

	if cfg.provider == nil {
		return nil, ErrProviderRequired
	}

	return &generator{cfg: cfg}, nil
}

func (g *generator) Generate(ctx context.Context, req Request) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	normalized := g.applyDefaults(req)
	if err := validateRequest(normalized); err != nil {
		return nil, err
	}

	result, err := g.cfg.provider.Generate(ctx, normalized)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (g *generator) applyDefaults(req Request) Request {
	if req.Format == "" {
		req.Format = g.cfg.defaultFormat
	}
	if req.ErrorCorrection == "" {
		req.ErrorCorrection = g.cfg.defaultErrorCorrection
	}
	if req.Scale == 0 {
		req.Scale = g.cfg.defaultScale
	}
	if req.Margin == 0 {
		req.Margin = g.cfg.defaultMargin
	}
	if req.Encoding == "" {
		req.Encoding = g.cfg.defaultEncodingMode
	}

	return req
}
