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

// Config holds validator inputs: rule defines allowed uploads and virusScanner
// performs optional malware checks (nil skips virus scanning).
type Config struct {
	rule         *FileRule
	virusScanner VirusScanner
}

// Option represents option.
type Option func(*Config)

// WithRule sets the FileRule used for MIME, extension, and size validation.
func WithRule(rule *FileRule) Option {
	return func(c *Config) {
		c.rule = rule
	}
}

// WithProfile sets the rule from a registered UploadProfile; unknown profiles are ignored.
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

// WithVirusScanner sets the scanner invoked after rule checks; nil disables scanning.
func WithVirusScanner(scanner VirusScanner) Option {
	return func(c *Config) {
		c.virusScanner = scanner
	}
}

// WithClamAV configures a ClamAVScanner at address as the virus scanner.
func WithClamAV(address string) Option {
	return func(c *Config) {
		c.virusScanner = NewClamAVScanner(address)
	}
}

// WithCustomRule sets the rule from explicit MIME, extension, and size limits.
func WithCustomRule(maxSize int64, mimes []string, exts []string) Option {
	return func(c *Config) {
		c.rule = NewCustomRule(maxSize, mimes, exts)
	}
}

// NewUploadFileValidator builds a Config from options. At least one option must
// supply a non-nil rule (WithRule, WithProfile, or WithCustomRule); otherwise it returns an error.
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
