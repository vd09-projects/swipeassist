package extractor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/vision-traits/traits"
)

type fakeTraitsExtractor struct {
	t             *testing.T
	response      traits.ExtractedTraits
	err           error
	responses     []traits.ExtractedTraits
	errs          []error
	capturedPaths []string
	callCount     int
}

func (f *fakeTraitsExtractor) ExtractFromPaths(ctx context.Context, paths []string) (traits.ExtractedTraits, error) {
	f.callCount++
	f.capturedPaths = append([]string(nil), paths...)
	if ctx == nil {
		f.t.Fatalf("expected context, got nil")
	}

	idx := f.callCount - 1

	resp := f.response
	if len(f.responses) > 0 {
		if idx < len(f.responses) {
			resp = f.responses[idx]
		} else {
			resp = f.responses[len(f.responses)-1]
		}
	}

	err := f.err
	if len(f.errs) > 0 {
		if idx < len(f.errs) {
			err = f.errs[idx]
		} else {
			err = f.errs[len(f.errs)-1]
		}
	}

	return resp, err
}

func TestVisionExtractorExtractBehaviour(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "img.png")
	if err := os.WriteFile(imgPath, []byte("stub"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	mockTraits := &fakeTraitsExtractor{
		t: t,
		response: traits.ExtractedTraits{
			GlobalConfidence: 88,
			Traits: map[string]traits.TraitCategoryResult{
				"ui_free_text": {
					Signals: []string{"should", "fallback", "be", "ignored"},
					SignalsByKey: map[string][]string{
						"lines": {"  first line  ", "", "second line"},
					},
					Confidence: 70,
				},
				"ui_sections": {
					SignalsByKey: map[string][]string{
						" About ": {"  loves art ", "", "travel"},
						"":        {"should skip"},
					},
					Confidence: 55,
				},
				"ui_tags": {
					SignalsByKey: map[string][]string{
						"raw":   {"  adventurous ", "explorer  "},
						"hobby": {" Running "},
						"":      {"ignored"},
					},
					Confidence: 77,
				},
			},
		},
	}

	e := &VisionExtractor{
		behaviourTr: mockTraits,
	}

	got, err := e.ExtractBehaviour(context.Background(), "", []string{imgPath})
	if err != nil {
		t.Fatalf("ExtractBehaviour returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("ExtractBehaviour returned nil traits")
	}
	if mockTraits.callCount != 1 {
		t.Fatalf("expected extractor called once, got %d", mockTraits.callCount)
	}
	if !reflect.DeepEqual(mockTraits.capturedPaths, []string{imgPath}) {
		t.Fatalf("unexpected paths passed to extractor: %v", mockTraits.capturedPaths)
	}

	want := &domain.BehaviourTraits{
		GlobalConfidence: 88,
		RawText: &domain.RawTextBlock{
			Confidence: 70,
			Lines:      []string{"first line", "second line"},
		},
		QASections: &domain.QASectionsBlock{
			Confidence: 55,
			QA: map[string][]string{
				"About": {"loves art", "travel"},
			},
		},
		ProfileTags: &domain.ProfileTagsBlock{
			Confidence: 77,
			Tags: map[string][]string{
				"hobby": {"Running"},
			},
			Raw: []string{"adventurous", "explorer"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractBehaviour mismatch:\nwant %#v\n got %#v", want, got)
	}
}

func TestVisionExtractorRetriesAndSucceeds(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "img.png")
	if err := os.WriteFile(imgPath, []byte("stub"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	extractionErr := errors.New("first attempt failed")
	mockTraits := &fakeTraitsExtractor{
		t:         t,
		responses: []traits.ExtractedTraits{{}, {GlobalConfidence: 42}},
		errs:      []error{extractionErr, nil},
	}

	e := &VisionExtractor{
		behaviourTr:   mockTraits,
		retryAttempts: 2,
		retryDelay:    0,
	}

	got, err := e.ExtractBehaviour(context.Background(), "", []string{imgPath})
	if err != nil {
		t.Fatalf("ExtractBehaviour returned error: %v", err)
	}
	if mockTraits.callCount != 2 {
		t.Fatalf("expected extractor called twice, got %d", mockTraits.callCount)
	}
	if got == nil || got.GlobalConfidence != 42 {
		t.Fatalf("unexpected result after retries: %#v", got)
	}
}

func TestVisionExtractorRetriesAndFails(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "img.png")
	if err := os.WriteFile(imgPath, []byte("stub"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	firstErr := errors.New("first attempt failed")
	secondErr := errors.New("second attempt failed")
	mockTraits := &fakeTraitsExtractor{
		t:    t,
		errs: []error{firstErr, secondErr},
	}

	e := &VisionExtractor{
		behaviourTr:   mockTraits,
		retryAttempts: 2,
		retryDelay:    0,
	}

	got, err := e.ExtractBehaviour(context.Background(), "", []string{imgPath})
	if err == nil {
		t.Fatalf("expected ExtractBehaviour to fail")
	}
	if !errors.Is(err, secondErr) {
		t.Fatalf("expected final error %v, got %v", secondErr, err)
	}
	if mockTraits.callCount != 2 {
		t.Fatalf("expected extractor called twice, got %d", mockTraits.callCount)
	}
	if got == nil {
		t.Fatalf("expected non-nil result even when failing")
	}
}

func TestMapPhotosToPersonaBundle(t *testing.T) {
	t.Parallel()

	invalidPhoto := &traits.ExtractedTraits{
		Traits: map[string]traits.TraitCategoryResult{
			"style": {
				Signals: []string{"unknown"},
			},
		},
	}

	photoOne := &traits.ExtractedTraits{
		Traits: map[string]traits.TraitCategoryResult{
			"style": {
				Signals: []string{" Adventurous ", "unknown", "adventurous"},
				Summary: "  Loves Travel ",
			},
			"vibe": {
				Signals: []string{"calm", " Bold "},
				Summary: "",
			},
			"skip": {
				Signals: []string{"unknown"},
			},
		},
	}

	photoTwo := &traits.ExtractedTraits{
		Traits: map[string]traits.TraitCategoryResult{
			"career": {
				Signals: []string{"Engineer", "unknown", " engineer "},
				Summary: "Builds robots",
			},
			"hobby": {
				Signals: []string{"Photography", " engineer "},
				Summary: "  Enjoys sunrises ",
			},
		},
	}

	bundle := MapPhotosToPersonaBundle([]*traits.ExtractedTraits{invalidPhoto, photoOne, nil, photoTwo})
	if bundle == nil {
		t.Fatalf("MapPhotosToPersonaBundle returned nil")
	}

	if len(bundle.Images) != 2 {
		t.Fatalf("expected 2 valid photos, got %d", len(bundle.Images))
	}

	first, ok := bundle.Images["image_1"]
	if !ok {
		t.Fatalf("image_1 missing from bundle: %#v", bundle.Images)
	}
	second, ok := bundle.Images["image_2"]
	if !ok {
		t.Fatalf("image_2 missing from bundle: %#v", bundle.Images)
	}

	expectedFirstTraits := map[string][]string{
		"style": {"adventurous"},
		"vibe":  {"bold", "calm"},
	}
	if !reflect.DeepEqual(first.Traits, expectedFirstTraits) {
		t.Fatalf("unexpected traits for photo1: %#v", first.Traits)
	}
	if !reflect.DeepEqual(first.Tags, []string{"adventurous", "bold", "calm"}) {
		t.Fatalf("unexpected tags for photo1: %#v", first.Tags)
	}
	if !reflect.DeepEqual(first.Statements, []string{"Loves Travel"}) {
		t.Fatalf("unexpected statements for photo1: %#v", first.Statements)
	}

	expectedSecondTraits := map[string][]string{
		"career": {"engineer"},
		"hobby":  {"engineer", "photography"},
	}
	if !reflect.DeepEqual(second.Traits, expectedSecondTraits) {
		t.Fatalf("unexpected traits for photo2: %#v", second.Traits)
	}
	if !reflect.DeepEqual(second.Tags, []string{"engineer", "photography"}) {
		t.Fatalf("unexpected tags for photo2: %#v", second.Tags)
	}
	if !reflect.DeepEqual(second.Statements, []string{"Builds robots", "Enjoys sunrises"}) {
		t.Fatalf("unexpected statements for photo2: %#v", second.Statements)
	}
}
