package qr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/abhissng/neuron/adapters/qr"
)

type mockProvider struct {
	lastRequest qr.Request
	result      *qr.Result
	err         error
}

func (m *mockProvider) Generate(_ context.Context, req qr.Request) (*qr.Result, error) {
	m.lastRequest = req
	if m.err != nil {
		return nil, m.err
	}
	if m.result == nil {
		m.result = &qr.Result{Format: qr.FormatPNG, ContentType: "image/png", Data: []byte("ok")}
	}
	return m.result, nil
}

func TestGenerate_AppliesDefaults(t *testing.T) {
	provider := &mockProvider{}
	generator, err := qr.NewGenerator(qr.WithProvider(provider))
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), qr.Request{
		Payload: "https://example.com",
	})
	require.NoError(t, err)

	assert.Equal(t, qr.FormatPNG, provider.lastRequest.Format)
	assert.Equal(t, qr.ErrorCorrectionMedium, provider.lastRequest.ErrorCorrection)
	assert.Equal(t, qr.EncodingModeAuto, provider.lastRequest.Encoding)
	assert.Equal(t, 10, provider.lastRequest.Scale)
	assert.Equal(t, 4, provider.lastRequest.Margin)
}

func TestGenerate_ValidationFailures(t *testing.T) {
	provider := &mockProvider{}
	generator, err := qr.NewGenerator(qr.WithProvider(provider))
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       qr.Request
		targetErr error
	}{
		{
			name:      "empty payload",
			req:       qr.Request{Payload: "   "},
			targetErr: qr.ErrEmptyPayload,
		},
		{
			name: "invalid format",
			req: qr.Request{
				Payload: "data",
				Format:  "gif",
			},
			targetErr: qr.ErrUnsupportedFormat,
		},
		{
			name: "invalid scale",
			req: qr.Request{
				Payload: "data",
				Scale:   -1,
			},
			targetErr: qr.ErrInvalidSize,
		},
		{
			name: "invalid margin",
			req: qr.Request{
				Payload: "data",
				Margin:  -2,
			},
			targetErr: qr.ErrInvalidMargin,
		},
		{
			name: "invalid ecc",
			req: qr.Request{
				Payload:         "data",
				ErrorCorrection: "extreme",
			},
			targetErr: qr.ErrInvalidConfig,
		},
		{
			name: "invalid encoding",
			req: qr.Request{
				Payload:  "data",
				Encoding: "manual",
			},
			targetErr: qr.ErrInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := generator.Generate(context.Background(), tc.req)
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.targetErr)
		})
	}
}

func TestGenerate_ContextCancellation(t *testing.T) {
	provider := &mockProvider{}
	generator, err := qr.NewGenerator(qr.WithProvider(provider))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = generator.Generate(ctx, qr.Request{Payload: "https://example.com"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}
