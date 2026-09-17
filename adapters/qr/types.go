package qr

// Format identifies output format for rendered QR codes.
type Format string

const (
	FormatPNG Format = "png"
	FormatSVG Format = "svg"
)

// ErrorCorrection identifies QR error-correction level.
type ErrorCorrection string

const (
	ErrorCorrectionLow      ErrorCorrection = "low"
	ErrorCorrectionMedium   ErrorCorrection = "medium"
	ErrorCorrectionQuartile ErrorCorrection = "quartile"
	ErrorCorrectionHigh     ErrorCorrection = "high"
)

// EncodingMode controls how the payload should be encoded.
type EncodingMode string

const (
	EncodingModeAuto            EncodingMode = "auto"
	EncodingModeOptimalSegments EncodingMode = "optimal_segments"
)

// RenderOptions carries optional rendering customizations that are generic
// enough to remain provider-independent.
type RenderOptions struct {
	Foreground          string
	Background          string
	CompactSVG          bool
	IncludeSVGXMLHeader bool
	Logo                []byte
	LogoSizeRatio       float64
}

// NewRenderOptions returns empty render options. Empty fields leave rendering
// choices to the configured provider.
func NewRenderOptions() *RenderOptions {
	return &RenderOptions{}
}

// DefaultRenderOptions returns conventional QR rendering options. Services can
// start with these values and override only the options they need.
func DefaultRenderOptions() *RenderOptions {
	return &RenderOptions{
		Foreground: "#000000",
		Background: "#FFFFFF",
	}
}

// Request is a provider-agnostic QR generation request.
//
// Defaults are applied by the generator when fields are left zero-valued.
type Request struct {
	Payload         string
	Format          Format
	ErrorCorrection ErrorCorrection
	Scale           int
	Margin          int
	Encoding        EncodingMode
	Render          *RenderOptions
}

// NewRequest returns an empty QR request. The generator applies its configured
// defaults for every optional field when Generate is called.
func NewRequest(payload string) *Request {
	return &Request{
		Payload: payload,
	}
}

// DefaultRequest returns a QR request populated with the package's standard
// generation defaults. Payload remains empty and must be supplied before use.
func DefaultRequest(payload string) *Request {
	return &Request{
		Payload:         payload,
		Format:          FormatPNG,
		ErrorCorrection: ErrorCorrectionMedium,
		Scale:           10,
		Margin:          4,
		Encoding:        EncodingModeAuto,
	}
}

// Result contains a generated QR artifact and metadata useful for HTTP responses.
type Result struct {
	Format      Format
	ContentType string
	Data        []byte
	Width       int
	Height      int
}
