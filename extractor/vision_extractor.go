package extractor

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/utils"
	"github.com/vd09-projects/vision-traits/traits"
)

type traitsPathExtractor interface {
	ExtractFromPaths(ctx context.Context, paths []string) (traits.ExtractedTraits, error)
}

type VisionExtractor struct {
	behaviourTr traitsPathExtractor
	personaTr   traitsPathExtractor
}

func NewVisionExtractor(eCfg *ExtractorConfig) (Extractor, error) {
	if eCfg == nil {
		return nil, fmt.Errorf("Extractor config is nil")
	}

	bTr, err := traits.New(
		traits.WithConfigPath(eCfg.BehaviourCfgPath),
	)
	if err != nil {
		return nil, err
	}

	pTr, err := traits.New(
		traits.WithConfigPath(eCfg.PersonaCfgPath),
	)
	if err != nil {
		return nil, err
	}

	return &VisionExtractor{
		behaviourTr: bTr,
		personaTr:   pTr,
	}, nil
}

func (e *VisionExtractor) ExtractBehaviour(ctx context.Context, imagePaths []string) (*domain.BehaviourTraits, error) {
	// check file exists
	for _, imagePath := range imagePaths {
		if _, err := os.Stat(imagePath); err != nil {
			return nil, fmt.Errorf("image not found: %w", err)
		}
	}

	et, err := e.behaviourTr.ExtractFromPaths(ctx, imagePaths)
	return mapToBehaviourTraits(&et), err
}

func (e *VisionExtractor) ExtractPhotoPersona(ctx context.Context, imagePaths []string) (*traits.ExtractedTraits, error) {
	for _, imagePath := range imagePaths {
		if _, err := os.Stat(imagePath); err != nil {
			return nil, fmt.Errorf("image not found: %w", err)
		}
	}

	et, err := e.personaTr.ExtractFromPaths(ctx, imagePaths)
	return &et, err
}

func mapToBehaviourTraits(in *traits.ExtractedTraits) *domain.BehaviourTraits {
	if in == nil {
		return nil
	}

	out := &domain.BehaviourTraits{
		GlobalConfidence: in.GlobalConfidence,
	}

	// ---- raw_text (ui_free_text) ----
	if tc, ok := in.Traits["ui_free_text"]; ok {
		out.RawText = &domain.RawTextBlock{
			Confidence: tc.Confidence,
		}

		// Prefer signals_by_key.lines if present, else fallback to signals
		if lines, ok := tc.SignalsByKey["lines"]; ok && len(lines) > 0 {
			out.RawText.Lines = trimKeepOrder(lines)
		} else {
			out.RawText.Lines = trimKeepOrder(tc.Signals)
		}
	}

	// ---- qa_sections (ui_sections) ----
	if tc, ok := in.Traits["ui_sections"]; ok {
		qa := map[string][]string{}
		for k, v := range tc.SignalsByKey {
			// direct mapping: header -> list of lines
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			qa[key] = trimKeepOrder(v)
		}

		out.QASections = &domain.QASectionsBlock{
			Confidence: tc.Confidence,
			QA:         qa,
		}
	}

	// ---- profile_tags (ui_tags) ----
	if tc, ok := in.Traits["ui_tags"]; ok {
		out.ProfileTags = &domain.ProfileTagsBlock{
			Confidence: tc.Confidence,
		}
		out.ProfileTags.Tags = map[string][]string{}

		for k, v := range tc.SignalsByKey {
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			if key == "raw" {
				out.ProfileTags.Raw = trimKeepOrder(v)
				continue
			}
			out.ProfileTags.Tags[key] = trimKeepOrder(v)
		}
	}

	return out
}

func trimKeepOrder(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func MapPhotosToPersonaBundle(photos []*traits.ExtractedTraits) *domain.PhotoPersonaBundle {
	out := &domain.PhotoPersonaBundle{
		Images: map[string]domain.PhotoPersonaProfile{},
	}

	kept := 0
	for _, photo := range photos {
		if isInvalidPhoto(photo) {
			continue
		}
		kept++
		key := "image_" + strconv.Itoa(kept)
		out.Images[key] = mapSinglePhoto(photo)
	}

	return out
}

func mapSinglePhoto(photo *traits.ExtractedTraits) domain.PhotoPersonaProfile {
	res := domain.PhotoPersonaProfile{
		Traits: map[string][]string{},
	}

	if photo == nil || len(photo.Traits) == 0 {
		return res
	}

	// deterministic trait iteration
	traitKeys := make([]string, 0, len(photo.Traits))
	for k := range photo.Traits {
		traitKeys = append(traitKeys, k)
	}
	sort.Strings(traitKeys)

	for _, traitKey := range traitKeys {
		tc := photo.Traits[traitKey]

		// IMPORTANT: only tc.Signals (not signals_by_key)
		signals := normalizeSignals(tc.Signals)

		// drop "unknown"
		filtered := make([]string, 0, len(signals))
		for _, s := range signals {
			if s == "unknown" {
				continue
			}
			filtered = append(filtered, s)
		}
		if len(filtered) == 0 {
			continue
		}

		// grouped traits
		res.Traits[traitKey] = addUniqueSorted(res.Traits[traitKey], filtered)

		// flat tags
		res.Tags = addUniqueSorted(res.Tags, filtered)

		// optional: keep a human statement from summary
		if sum := strings.TrimSpace(tc.Summary); sum != "" {
			res.Statements = addUniqueSorted(res.Statements, []string{sum})
		}
	}

	// Clean empties (nice JSON)
	if len(res.Tags) == 0 {
		res.Tags = nil
	}
	if len(res.Traits) == 0 {
		res.Traits = nil
	}
	if len(res.Statements) == 0 {
		res.Statements = nil
	}

	return res
}

func isAllUnknownOrEmpty(signals []string) bool {
	if len(signals) == 0 {
		return true
	}
	for _, s := range signals {
		if s != "unknown" {
			return false
		}
	}
	return true
}

// isInvalidPhoto checks if ALL traits are unknown/empty (or traits missing).
func isInvalidPhoto(et *traits.ExtractedTraits) bool {
	if et == nil || len(et.Traits) == 0 {
		return true
	}
	for _, tc := range et.Traits {
		ns := normalizeSignals(tc.Signals)
		// If any trait contains a non-unknown signal => photo is valid
		if !isAllUnknownOrEmpty(ns) {
			return false
		}
	}
	return true
}

func normalizeSignals(in []string) []string {
	return utils.NormalizeAndDedupe(
		in,
		func(s string) (string, bool) {
			s = strings.ToLower(strings.TrimSpace(s))
			return s, s != ""
		},
		utils.Less,
		utils.Identity,
	)
}

func addUniqueSorted(dst, src []string) []string {
	return utils.MergeUniqueAndSort(
		dst,
		src,
		utils.Identity,
		utils.Less,
	)
}
