package qr_test

import (
	"context"
	"fmt"

	"github.com/abhissng/neuron/adapters/qr"
	"github.com/abhissng/neuron/adapters/qr/piglig"
)

func Example_basic() {
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

	fmt.Println(result.ContentType)
	// Output: image/png
}

func Example_advanced() {
	generator, err := piglig.NewGenerator(
		qr.WithDefaultErrorCorrection(qr.ErrorCorrectionHigh),
	)
	if err != nil {
		panic(err)
	}

	result, err := generator.Generate(context.Background(), qr.Request{
		Payload: "myapp://auth/session/abc",
		Format:  qr.FormatSVG,
		Scale:   8,
		Margin:  4,
		Render: &qr.RenderOptions{
			Foreground: "#111111",
			Background: "#FFFFFF",
			CompactSVG: true,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result.ContentType)
	// Output: image/svg+xml
}
