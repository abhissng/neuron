package piglig

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	go_qr "github.com/piglig/go-qr"

	"github.com/abhissng/neuron/adapters/qr"
)

const (
	contentTypePNG = "image/png"
	contentTypeSVG = "image/svg+xml"
)

// Provider implements qr.Provider using github.com/piglig/go-qr.
// It is stateless and safe for concurrent use.
type Provider struct{}

func init() {
	qr.RegisterDefaultProvider(func() qr.Provider { return New() })
}

// New creates a new stateless piglig-backed provider.
func New() *Provider {
	return &Provider{}
}

// NewGenerator returns a qr.Generator wired with the piglig provider.
func NewGenerator(options ...qr.GeneratorOption) (qr.Generator, error) {
	allOptions := make([]qr.GeneratorOption, 0, len(options)+1)
	allOptions = append(allOptions, qr.WithProvider(New()))
	allOptions = append(allOptions, options...)
	return qr.NewGenerator(allOptions...)
}

// Generate produces a PNG or SVG QR artifact using the generic request model.
func (p *Provider) Generate(ctx context.Context, req qr.Request) (*qr.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ecc, err := toPigligECC(req.ErrorCorrection)
	if err != nil {
		return nil, err
	}

	code, err := encode(req, ecc)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := qr.ValidateRenderedDimensions(code.Size(), req); err != nil {
		return nil, err
	}

	cfg, err := toPigligConfig(req)
	if err != nil {
		return nil, err
	}

	dimension := (code.Size() + (2 * req.Margin)) * req.Scale
	result := &qr.Result{
		Format: req.Format,
		Width:  dimension,
		Height: dimension,
	}

	switch req.Format {
	case qr.FormatPNG:
		result.ContentType = contentTypePNG
		result.Data, err = code.ToPNGBytes(cfg)
		if err != nil {
			return nil, wrapProviderError(err, qr.ErrRenderFailed)
		}
	case qr.FormatSVG:
		result.ContentType = contentTypeSVG
		result.Data, err = code.ToSVGBytes(cfg)
		if err != nil {
			return nil, wrapProviderError(err, qr.ErrRenderFailed)
		}
	default:
		return nil, fmt.Errorf("%w: %q", qr.ErrUnsupportedFormat, req.Format)
	}

	return result, nil
}

func encode(req qr.Request, ecc go_qr.Ecc) (*go_qr.QrCode, error) {
	switch req.Encoding {
	case qr.EncodingModeAuto:
		code, err := go_qr.EncodeText(req.Payload, ecc)
		if err != nil {
			return nil, wrapProviderError(err, qr.ErrEncodingFailed)
		}
		return code, nil
	case qr.EncodingModeOptimalSegments:
		segments, err := go_qr.MakeSegmentsOptimally(req.Payload, ecc, go_qr.MinVersion, go_qr.MaxVersion)
		if err != nil {
			return nil, wrapProviderError(err, qr.ErrEncodingFailed)
		}

		code, err := go_qr.EncodeSegments(segments, ecc, go_qr.MinVersion, go_qr.MaxVersion, -1, true)
		if err != nil {
			return nil, wrapProviderError(err, qr.ErrEncodingFailed)
		}
		return code, nil
	default:
		return nil, fmt.Errorf("%w: %q", qr.ErrInvalidConfig, req.Encoding)
	}
}

func toPigligConfig(req qr.Request) (*go_qr.QrCodeImgConfig, error) {
	options := make([]go_qr.Option, 0, 4)
	if req.Render != nil {
		if req.Render.Foreground != "" {
			dark, err := parseHexColor(req.Render.Foreground)
			if err != nil {
				return nil, err
			}
			options = append(options, go_qr.WithDark(dark))
		}

		if req.Render.Background != "" {
			light, err := parseHexColor(req.Render.Background)
			if err != nil {
				return nil, err
			}
			options = append(options, go_qr.WithLight(light))
		}

		if req.Render.IncludeSVGXMLHeader {
			options = append(options, go_qr.WithSVGXMLHeader())
		}

		if req.Render.CompactSVG {
			options = append(options, go_qr.WithOptimalSVG())
		}

		if len(req.Render.Logo) > 0 {
			img, _, err := image.Decode(bytes.NewReader(req.Render.Logo))
			if err != nil {
				return nil, fmt.Errorf("%w: failed to decode logo image: %v", qr.ErrInvalidConfig, err)
			}
			options = append(options, go_qr.WithLogo(img, req.Render.LogoSizeRatio))
		}
	}

	margin := req.Margin
	// go-qr v1.1.0 SVG layout treats border as pixels; scale quiet zone by Scale
	// so output matches PNG dimensions. Logo placement still expects module borders,
	// so keep module units when a logo is configured.
	if req.Format == qr.FormatSVG && (req.Render == nil || len(req.Render.Logo) == 0) {
		margin = req.Margin * req.Scale
	}

	return go_qr.NewQrCodeImgConfig(req.Scale, margin, options...), nil
}

func parseHexColor(value string) (color.RGBA, error) {
	clean := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(clean) != 6 {
		return color.RGBA{}, fmt.Errorf("%w: invalid color %q", qr.ErrInvalidConfig, value)
	}

	decoded, err := hex.DecodeString(clean)
	if err != nil || len(decoded) != 3 {
		return color.RGBA{}, fmt.Errorf("%w: invalid color %q", qr.ErrInvalidConfig, value)
	}

	return color.RGBA{R: decoded[0], G: decoded[1], B: decoded[2], A: 0xFF}, nil
}

func toPigligECC(level qr.ErrorCorrection) (go_qr.Ecc, error) {
	switch level {
	case qr.ErrorCorrectionLow:
		return go_qr.Low, nil
	case qr.ErrorCorrectionMedium:
		return go_qr.Medium, nil
	case qr.ErrorCorrectionQuartile:
		return go_qr.Quartile, nil
	case qr.ErrorCorrectionHigh:
		return go_qr.High, nil
	default:
		return 0, fmt.Errorf("%w: %q", qr.ErrInvalidConfig, level)
	}
}

func wrapProviderError(err error, wrapper error) error {
	switch {
	case errors.Is(err, go_qr.ErrDataTooLong):
		return fmt.Errorf("%w: %v", qr.ErrPayloadTooLarge, err)
	case errors.Is(err, go_qr.ErrInvalidConfig), errors.Is(err, go_qr.ErrInvalidArgument), errors.Is(err, go_qr.ErrInvalidVersion), errors.Is(err, go_qr.ErrUnencodableChar):
		return fmt.Errorf("%w: %v", qr.ErrInvalidConfig, err)
	default:
		return fmt.Errorf("%w: %v", wrapper, err)
	}
}
