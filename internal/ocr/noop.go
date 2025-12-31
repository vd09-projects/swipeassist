package ocr

import "context"

type NoopEngine struct{}

func (n NoopEngine) Name() string { return "noop" }

func (n NoopEngine) ExtractText(ctx context.Context, imageBytes []byte) (Result, error) {
	return Result{Text: "", Confidence: 0.0, Engine: n.Name()}, nil
}
