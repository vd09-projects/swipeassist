package extractor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNoopExtractorDelayAndResults(t *testing.T) {
	t.Parallel()

	delay := 20 * time.Millisecond
	ext := NewNoopExtractor(delay)

	start := time.Now()
	bt, err := ext.ExtractBehaviour(context.Background(), "", nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("ExtractBehaviour returned error: %v", err)
	}
	if bt == nil {
		t.Fatalf("expected BehaviourTraits, got nil")
	}
	if elapsed < delay {
		t.Fatalf("expected delay of at least %v, got %v", delay, elapsed)
	}

	pt, err := ext.ExtractPhotoPersona(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("ExtractPhotoPersona returned error: %v", err)
	}
	if pt == nil {
		t.Fatalf("expected ExtractedTraits, got nil")
	}
}

func TestNoopExtractorRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	ext := NewNoopExtractor(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	if _, err := ext.ExtractBehaviour(ctx, "", nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	if _, err := ext.ExtractPhotoPersona(ctx, "", nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
