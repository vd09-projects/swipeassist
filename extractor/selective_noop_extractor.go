package extractor

import (
	"context"
	"fmt"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/utils"
	"github.com/vd09-projects/vision-traits/traits"
)

// SelectiveNoopExtractor optionally no-ops behaviour and/or persona extraction while
// delegating the other calls to an underlying extractor (typically VisionExtractor).
type SelectiveNoopExtractor struct {
	delay          time.Duration
	behaviourNoop  bool
	personaNoop    bool
	innerExtractor Extractor
}

// NewSelectiveNoopExtractor wraps an existing extractor and allows toggling no-op behaviour.
// Negative delays are clamped to zero.
func NewSelectiveNoopExtractor(inner Extractor, delay time.Duration, behaviourNoop, personaNoop bool) (*SelectiveNoopExtractor, error) {
	if inner == nil {
		return nil, fmt.Errorf("inner extractor is nil")
	}
	if delay < 0 {
		delay = 0
	}
	return &SelectiveNoopExtractor{
		delay:          delay,
		behaviourNoop:  behaviourNoop,
		personaNoop:    personaNoop,
		innerExtractor: inner,
	}, nil
}

// NewVisionExtractorWithNoops constructs a VisionExtractor and wraps it with selective no-ops.
func NewVisionExtractorWithNoops(cfg *ExtractorConfig, delay time.Duration, behaviourNoop, personaNoop bool) (Extractor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Extractor config is nil")
	}
	inner, err := NewVisionExtractor(cfg)
	if err != nil {
		return nil, err
	}
	return NewSelectiveNoopExtractor(inner, delay, behaviourNoop, personaNoop)
}

func (s *SelectiveNoopExtractor) ExtractBehaviour(ctx context.Context, imagePaths []string) (*domain.BehaviourTraits, error) {
	if s.behaviourNoop {
		if err := utils.RandomSleepCtx(ctx, s.delay, 2*s.delay); err != nil {
			return nil, err
		}
		return &domain.BehaviourTraits{}, nil
	}
	return s.innerExtractor.ExtractBehaviour(ctx, imagePaths)
}

func (s *SelectiveNoopExtractor) ExtractPhotoPersona(ctx context.Context, imagePaths []string) (*traits.ExtractedTraits, error) {
	if s.personaNoop {
		if err := utils.RandomSleepCtx(ctx, s.delay, 2*s.delay); err != nil {
			return nil, err
		}
		return &traits.ExtractedTraits{}, nil
	}
	return s.innerExtractor.ExtractPhotoPersona(ctx, imagePaths)
}
