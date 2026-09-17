package qr

import "errors"

var (
	ErrProviderRequired  = errors.New("qr: provider is required")
	ErrEmptyPayload      = errors.New("qr: payload is empty")
	ErrInvalidSize       = errors.New("qr: invalid size")
	ErrInvalidMargin     = errors.New("qr: invalid margin")
	ErrUnsupportedFormat = errors.New("qr: unsupported output format")
	ErrInvalidConfig     = errors.New("qr: invalid configuration")
	ErrPayloadTooLarge   = errors.New("qr: payload too large")
	ErrEncodingFailed    = errors.New("qr: encoding failed")
	ErrRenderFailed      = errors.New("qr: render failed")
)
