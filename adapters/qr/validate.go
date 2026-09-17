package qr

import (
	"fmt"
	"strings"
)

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
