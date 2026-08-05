package uploadFile

import (
	"errors"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

/*
========================================
 Validation Options - Functional Pattern
========================================
*/

// Config represents config.
type Config struct {
	rule         *FileRule
	virusScanner VirusScanner
}

// Option represents option.
type Option func(*Config)

// WithRule applies an option value.
func WithRule(rule *FileRule) Option {
	return func(c *Config) {
		c.rule = rule
	}
}

// WithProfile applies an option value.
func WithProfile(profile UploadProfile) Option {
	return func(c *Config) {
		rule, ok := GetUploadProfile(profile)
		if !ok {
			helpers.Println(constant.ERROR, "WithProfile: profile not found")
			return
		}
		c.rule = rule
	}
}

// WithVirusScanner applies an option value.
func WithVirusScanner(scanner VirusScanner) Option {
	return func(c *Config) {
		c.virusScanner = scanner
	}
}

// WithClamAV applies an option value.
func WithClamAV(address string) Option {
	return func(c *Config) {
		c.virusScanner = NewClamAVScanner(address)
	}
}

// WithCustomRule applies an option value.
func WithCustomRule(maxSize int64, mimes []string, exts []string) Option {
	return func(c *Config) {
		c.rule = NewCustomRule(maxSize, mimes, exts)
	}
}

// NewUploadFileValidator creates a new instance.
func NewUploadFileValidator(opts ...Option) (*Config, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.rule == nil {
		helpers.Println(constant.ERROR, "NewUploadFileValidator: rule is required")
		return nil, errors.New("rule is required")
	}
	return cfg, nil
}
