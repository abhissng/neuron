// Package qr provides a provider-agnostic QR generation abstraction for Neuron.
//
// Supported output formats:
//   - PNG (default)
//   - SVG
//
// Default request behavior:
//   - Format: PNG
//   - Error correction: Medium
//   - Scale: 10
//   - Margin (quiet zone): 4
//   - Encoding mode: Auto
//
// The package is framework-independent and returns in-memory results suitable
// for HTTP responses (Content-Type + raw bytes).
//
// Concurrency:
// Generator implementations are expected to be safe for concurrent use.
// A single generator instance can be reused across requests.
package qr
