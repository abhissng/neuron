# QR Adapter

The `adapters/qr` package provides a provider-agnostic QR generation API for Neuron services.

## What it supports

- Opaque payload strings (URL, deep link, UPI URI, auth URL, JSON text, plain text)
- PNG and SVG output
- Error correction levels: low, medium, quartile, high
- Size controls: scale and margin (quiet zone)
- Optional rendering controls: foreground/background colors, compact SVG, SVG XML header, logo
- In-memory output for HTTP responses (no filesystem dependency)

## Defaults

- Format: `png`
- Error correction: `medium`
- Scale: `10`
- Margin: `4`
- Encoding mode: `auto`

## Quick start

```go
import (
    "context"

    "github.com/abhissng/neuron/adapters/qr"
    "github.com/abhissng/neuron/adapters/qr/piglig"
)

generator, err := piglig.NewGenerator()
if err != nil {
    panic(err)
}

result, err := generator.Generate(context.Background(), qr.Request{
    Payload: "https://example.com",
})
if err != nil {
    panic(err)
}

// result.ContentType => image/png
// result.Data => response body bytes
```

## Advanced configuration

```go
result, err := generator.Generate(ctx, qr.Request{
    Payload:         "upi://pay?pa=merchant@upi&pn=Merchant&am=499&cu=INR",
    Format:          qr.FormatSVG,
    ErrorCorrection: qr.ErrorCorrectionHigh,
    Scale:           8,
    Margin:          4,
    Encoding:        qr.EncodingModeOptimalSegments,
    Render: &qr.RenderOptions{
        Foreground:          "#111111",
        Background:          "#FFFFFF",
        CompactSVG:          true,
        IncludeSVGXMLHeader: true,
    },
})
```

## Error behavior

The package returns sentinel errors for common failure classes:

- `qr.ErrEmptyPayload`
- `qr.ErrInvalidSize`
- `qr.ErrInvalidMargin`
- `qr.ErrUnsupportedFormat`
- `qr.ErrInvalidConfig`
- `qr.ErrPayloadTooLarge`
- `qr.ErrEncodingFailed`
- `qr.ErrRenderFailed`

Use `errors.Is(err, target)` for handling.

## Concurrency

Generators are safe for concurrent use when backed by the piglig provider. A single instance can be reused across many requests.
