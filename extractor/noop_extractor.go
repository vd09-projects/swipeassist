package extractor

import (
	"context"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/utils"
	"github.com/vd09-projects/vision-traits/traits"
)

// NoopExtractor returns empty results while optionally sleeping to simulate latency.
type NoopExtractor struct {
	delay time.Duration
}

// NewNoopExtractor builds a no-op extractor that waits for the given delay
// before returning empty results. A zero delay skips sleeping; negative delays
// are treated as zero.
func NewNoopExtractor(delay time.Duration) *NoopExtractor {
	if delay < 0 {
		delay = 0
	}
	return &NoopExtractor{delay: delay}
}

func (n *NoopExtractor) ExtractBehaviour(ctx context.Context, _ []string) (*domain.BehaviourTraits, error) {
	if err := utils.RandomSleepCtx(ctx, n.delay, 2*n.delay); err != nil {
		return nil, err
	}
	return &domain.BehaviourTraits{}, nil
}

func (n *NoopExtractor) ExtractPhotoPersona(ctx context.Context, _ []string) (*traits.ExtractedTraits, error) {
	if err := utils.RandomSleepCtx(ctx, n.delay, 2*n.delay); err != nil {
		return nil, err
	}
	return &traits.ExtractedTraits{}, nil
}
