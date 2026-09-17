package qr

import (
	"fmt"
	"math"
	"strings"
)

// maxRenderDimension caps output width/height to avoid oversized allocations.
const maxRenderDimension = 16_384

func validateRequest(req Request) error {
	if strings.TrimSpace(req.Payload) == "" {
		return ErrEmptyPayload
	}

	switch req.Format {
	case FormatPNG, FormatSVG:
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedFormat, req.Format)
	}

	if req.Scale <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidSize, req.Scale)
	}

	if req.Margin < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidMargin, req.Margin)
	}

	switch req.ErrorCorrection {
	case ErrorCorrectionLow, ErrorCorrectionMedium, ErrorCorrectionQuartile, ErrorCorrectionHigh:
	default:
		return fmt.Errorf("%w: %q", ErrInvalidConfig, req.ErrorCorrection)
	}

	switch req.Encoding {
	case EncodingModeAuto, EncodingModeOptimalSegments:
	default:
		return fmt.Errorf("%w: %q", ErrInvalidConfig, req.Encoding)
	}

	if req.Render != nil && len(req.Render.Logo) > 0 {
		if req.Render.LogoSizeRatio <= 0 || req.Render.LogoSizeRatio >= 1 {
			return fmt.Errorf("%w: logo size ratio must be within (0,1), got %v", ErrInvalidConfig, req.Render.LogoSizeRatio)
		}
	}

	return nil
}

// ValidateRenderedDimensions rejects requests whose scaled output would exceed
// safe PNG/SVG bounds. moduleCount is the encoded QR module width (code size).
func ValidateRenderedDimensions(moduleCount int, req Request) error {
	if req.Margin > math.MaxInt32/2 {
		return fmt.Errorf("%w: %d", ErrInvalidMargin, req.Margin)
	}

	modules := int64(moduleCount) + int64(req.Margin)*2
	if modules <= 0 {
		return fmt.Errorf("%w: invalid module dimensions", ErrInvalidSize)
	}
	if modules > math.MaxInt32/int64(req.Scale) {
		return fmt.Errorf("%w: scale or margin too large", ErrInvalidSize)
	}

	dimension := modules * int64(req.Scale)
	if dimension > int64(maxRenderDimension) {
		return fmt.Errorf("%w: output dimension %d exceeds maximum %d", ErrInvalidSize, dimension, maxRenderDimension)
	}

	return nil
}
