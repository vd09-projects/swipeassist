package extractor

import (
	"context"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/vision-traits/traits"
)

type Extractor interface {
	ExtractBehaviour(ctx context.Context, imagePaths []string) (*domain.BehaviourTraits, error)
	ExtractPhotoPersona(ctx context.Context, imagePaths []string) (*traits.ExtractedTraits, error)
}
