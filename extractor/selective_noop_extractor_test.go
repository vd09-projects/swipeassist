package extractor

import (
	"context"
	"testing"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/vision-traits/traits"
)

type fakeExtractor struct {
	behaviourCalls int
	personaCalls   int
	lastBehaviour  []string
	lastPersona    []string

	behaviourResp *domain.BehaviourTraits
	personaResp   *traits.ExtractedTraits
	behaviourErr  error
	personaErr    error
}

func (f *fakeExtractor) ExtractBehaviour(ctx context.Context, imagePaths []string) (*domain.BehaviourTraits, error) {
	f.behaviourCalls++
	f.lastBehaviour = append([]string(nil), imagePaths...)
	return f.behaviourResp, f.behaviourErr
}

func (f *fakeExtractor) ExtractPhotoPersona(ctx context.Context, imagePaths []string) (*traits.ExtractedTraits, error) {
	f.personaCalls++
	f.lastPersona = append([]string(nil), imagePaths...)
	return f.personaResp, f.personaErr
}

func TestSelectiveNoopExtractor_NoopsBehaviourOnly(t *testing.T) {
	inner := &fakeExtractor{
		behaviourResp: &domain.BehaviourTraits{GlobalConfidence: 10},
		personaResp:   &traits.ExtractedTraits{GlobalConfidence: 20},
	}

	ext, err := NewSelectiveNoopExtractor(inner, 0, true, false)
	if err != nil {
		t.Fatalf("unexpected error constructing selective noop: %v", err)
	}

	bh, err := ext.ExtractBehaviour(context.Background(), []string{"a.png"})
	if err != nil {
		t.Fatalf("ExtractBehaviour returned error: %v", err)
	}
	if inner.behaviourCalls != 0 {
		t.Fatalf("behaviour extractor should not be called when noop=true")
	}
	if bh == nil || bh.GlobalConfidence != 0 {
		t.Fatalf("expected empty behaviour traits, got %#v", bh)
	}

	ph, err := ext.ExtractPhotoPersona(context.Background(), []string{"b.png"})
	if err != nil {
		t.Fatalf("ExtractPhotoPersona returned error: %v", err)
	}
	if inner.personaCalls != 1 {
		t.Fatalf("persona extractor should be called once, got %d", inner.personaCalls)
	}
	if ph == nil || ph.GlobalConfidence != 20 {
		t.Fatalf("unexpected persona response: %#v", ph)
	}
}

func TestSelectiveNoopExtractor_PassThroughWhenDisabled(t *testing.T) {
	inner := &fakeExtractor{
		behaviourResp: &domain.BehaviourTraits{GlobalConfidence: 30},
		personaResp:   &traits.ExtractedTraits{GlobalConfidence: 40},
	}

	ext, err := NewSelectiveNoopExtractor(inner, 0, false, false)
	if err != nil {
		t.Fatalf("unexpected error constructing selective noop: %v", err)
	}

	_, _ = ext.ExtractBehaviour(context.Background(), []string{"c.png"})
	_, _ = ext.ExtractPhotoPersona(context.Background(), []string{"d.png"})

	if inner.behaviourCalls != 1 || inner.personaCalls != 1 {
		t.Fatalf("expected pass-through calls, got behaviour=%d persona=%d", inner.behaviourCalls, inner.personaCalls)
	}
	if inner.lastBehaviour == nil || inner.lastBehaviour[0] != "c.png" {
		t.Fatalf("behaviour paths not forwarded, got %#v", inner.lastBehaviour)
	}
	if inner.lastPersona == nil || inner.lastPersona[0] != "d.png" {
		t.Fatalf("persona paths not forwarded, got %#v", inner.lastPersona)
	}
}
