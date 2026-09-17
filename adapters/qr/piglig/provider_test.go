package piglig_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/abhissng/neuron/adapters/qr"
	"github.com/abhissng/neuron/adapters/qr/piglig"
)

func TestProviderGenerate_PNG(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com",
		Format:  qr.FormatPNG,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, qr.FormatPNG, result.Format)
	assert.Equal(t, "image/png", result.ContentType)
	assert.NotEmpty(t, result.Data)
	assert.Greater(t, result.Width, 0)
	assert.Greater(t, result.Height, 0)
	assert.True(t, bytes.HasPrefix(result.Data, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}))

	cfg, err := png.DecodeConfig(bytes.NewReader(result.Data))
	require.NoError(t, err)
	assert.Equal(t, result.Width, cfg.Width)
	assert.Equal(t, result.Height, cfg.Height)
}

func TestDefaultGenerator_UsesRegisteredProvider(t *testing.T) {
	generator, err := qr.DefaultGenerator()
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com/default",
	})
	require.NoError(t, err)
	assert.Equal(t, "image/png", result.ContentType)
}

func TestProviderGenerate_SVGScaledQuietZone(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	const scale = 3
	const margin = 4
	result, err := generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com/scaled-svg",
		Format:  qr.FormatSVG,
		Scale:   scale,
		Margin:  margin,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	viewBox := regexp.MustCompile(`viewBox="0 0 (\d+) (\d+)"`).FindStringSubmatch(string(result.Data))
	require.Len(t, viewBox, 3)
	assert.Equal(t, viewBox[1], viewBox[2])
	assert.Equal(t, result.Width, atoi(t, viewBox[1]))
	assert.Equal(t, result.Height, atoi(t, viewBox[2]))
	assert.Greater(t, result.Width, 0)
}

func TestProviderGenerate_SVG(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), qr.Request{
		Payload: "myapp://auth/session/abc",
		Format:  qr.FormatSVG,
		Render: &qr.RenderOptions{
			CompactSVG:          true,
			IncludeSVGXMLHeader: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, qr.FormatSVG, result.Format)
	assert.Equal(t, "image/svg+xml", result.ContentType)
	assert.NotEmpty(t, result.Data)
	assert.Contains(t, string(result.Data), "<svg")
}

func TestProviderGenerate_ErrorCorrectionLevels(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	levels := []qr.ErrorCorrection{
		qr.ErrorCorrectionLow,
		qr.ErrorCorrectionMedium,
		qr.ErrorCorrectionQuartile,
		qr.ErrorCorrectionHigh,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			result, err := generator.Generate(context.Background(), qr.Request{
				Payload:         "https://example.com/test-ecc",
				ErrorCorrection: level,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, result.Data)
		})
	}
}

func TestProviderGenerate_URIPayloads(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	payloads := []string{
		"https://example.com/payment/123",
		"upi://pay?pa=merchant@upi&pn=Merchant&am=499&cu=INR",
		"myapp://auth/session/abc",
		`{"kind":"auth","token":"abc123"}`,
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			result, err := generator.Generate(context.Background(), qr.Request{Payload: payload})
			require.NoError(t, err)
			assert.NotEmpty(t, result.Data)
		})
	}
}

func TestProviderGenerate_UnsupportedFormat(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), qr.Request{
		Payload: "data",
		Format:  "jpeg",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, qr.ErrUnsupportedFormat)
}

func TestProviderGenerate_InvalidColorAndLogo(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com",
		Render: &qr.RenderOptions{
			Foreground: "not-a-color",
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, qr.ErrInvalidConfig)

	_, err = generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com",
		Render: &qr.RenderOptions{
			Foreground: "#00000g",
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, qr.ErrInvalidConfig)

	logoBytes := makeTinyPNG(t)
	_, err = generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com",
		Render: &qr.RenderOptions{
			Logo:          logoBytes,
			LogoSizeRatio: 1.2,
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, qr.ErrInvalidConfig)
}

func TestProviderGenerate_OutputDimensionTooLarge(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com",
		Scale:   20_000,
		Margin:  4,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, qr.ErrInvalidSize)
}

func TestProviderGenerate_PayloadTooLarge(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	payload := strings.Repeat("x", 20000)
	_, err = generator.Generate(context.Background(), qr.Request{
		Payload:         payload,
		ErrorCorrection: qr.ErrorCorrectionHigh,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, qr.ErrPayloadTooLarge)
}

func TestProviderGenerate_Concurrent(t *testing.T) {
	generator, err := piglig.NewGenerator()
	require.NoError(t, err)

	const workers = 32
	const perWorker = 10

	errCh := make(chan error, workers*perWorker)
	var wg sync.WaitGroup

	for i := range workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range perWorker {
				result, err := generator.Generate(context.Background(), qr.Request{
					Payload: fmt.Sprintf("https://example.com/w/%d/j/%d", workerID, j),
					Format:  qr.FormatPNG,
				})
				if err != nil {
					errCh <- err
					return
				}
				if len(result.Data) == 0 {
					errCh <- assert.AnError
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	require.NoError(t, err)
	return n
}

func makeTinyPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)
	return buf.Bytes()
}
