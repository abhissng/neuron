package qr

type generatorConfig struct {
	provider               Provider
	defaultFormat          Format
	defaultErrorCorrection ErrorCorrection
	defaultScale           int
	defaultMargin          int
	defaultEncodingMode    EncodingMode
}

// GeneratorOption configures generator defaults and provider wiring.
type GeneratorOption func(*generatorConfig)

func defaultGeneratorConfig() generatorConfig {
	return generatorConfig{
		defaultFormat:          FormatPNG,
		defaultErrorCorrection: ErrorCorrectionMedium,
		defaultScale:           10,
		defaultMargin:          4,
		defaultEncodingMode:    EncodingModeAuto,
	}
}

// WithProvider sets the concrete QR provider implementation.
func WithProvider(provider Provider) GeneratorOption {
	return func(cfg *generatorConfig) {
		cfg.provider = provider
	}
}

// WithDefaultFormat overrides default output format.
func WithDefaultFormat(format Format) GeneratorOption {
	return func(cfg *generatorConfig) {
		cfg.defaultFormat = format
	}
}

// WithDefaultErrorCorrection overrides default error correction.
func WithDefaultErrorCorrection(level ErrorCorrection) GeneratorOption {
	return func(cfg *generatorConfig) {
		cfg.defaultErrorCorrection = level
	}
}

// WithDefaultScale overrides default rendered module scale.
func WithDefaultScale(scale int) GeneratorOption {
	return func(cfg *generatorConfig) {
		cfg.defaultScale = scale
	}
}

// WithDefaultMargin overrides default quiet zone (in modules).
func WithDefaultMargin(margin int) GeneratorOption {
	return func(cfg *generatorConfig) {
		cfg.defaultMargin = margin
	}
}

// WithDefaultEncodingMode overrides default payload encoding mode.
func WithDefaultEncodingMode(mode EncodingMode) GeneratorOption {
	return func(cfg *generatorConfig) {
		cfg.defaultEncodingMode = mode
	}
}
