package ocr

import "context"

type Result struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"` // 0..1 (best-effort)
	Engine     string  `json:"engine"`
}

type Engine interface {
	Name() string
	ExtractText(ctx context.Context, imageBytes []byte) (Result, error)
}