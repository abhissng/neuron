package uploadFile

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
		if rule, ok := GetUploadProfile(profile); ok {
			c.rule = rule
		}
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

func NewUploadFileValidator(opts ...Option) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
