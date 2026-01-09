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

type Config struct {
	rule         *FileRule
	virusScanner VirusScanner
}

type Option func(*Config)

func WithRule(rule *FileRule) Option {
	return func(c *Config) {
		c.rule = rule
	}
}

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

func WithVirusScanner(scanner VirusScanner) Option {
	return func(c *Config) {
		c.virusScanner = scanner
	}
}

func WithClamAV(address string) Option {
	return func(c *Config) {
		c.virusScanner = NewClamAVScanner(address)
	}
}

func WithCustomRule(maxSize int64, mimes []string, exts []string) Option {
	return func(c *Config) {
		c.rule = NewCustomRule(maxSize, mimes, exts)
	}
}

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
