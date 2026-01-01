package extractor

import (
	"context"

	"github.com/vd09-projects/vision-traits/traits"
)

type Extractor interface {
	ExtractText(ctx context.Context, imagePaths []string) (traits.ExtractedTraits, error)
}
