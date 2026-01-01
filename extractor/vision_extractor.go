package extractor

import (
	"context"
	"fmt"
	"os"

	"github.com/vd09-projects/vision-traits/traits"
)

type VisionExtractor struct {
	vtr *traits.VisionTraits
}

func New(cfgPath string) (*VisionExtractor, error) {
	tr, err := traits.New(
		traits.WithConfigPath(cfgPath),
	)
	if err != nil {
		return nil, err
	}
	return &VisionExtractor{
		vtr: tr,
	}, nil
}

func (e *VisionExtractor) ExtractText(ctx context.Context, imagePaths []string) (*traits.ExtractedTraits, error) {
	// check file exists
	for _, imagePath := range imagePaths {
		if _, err := os.Stat(imagePath); err != nil {
			return nil, fmt.Errorf("image not found: %w", err)
		}
	}

	et, err := e.vtr.ExtractFromPaths(ctx, imagePaths)
	return &et, err
}
