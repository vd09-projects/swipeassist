package extractor

import (
	"context"
	"strings"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/internal/dbgen"
	"github.com/vd09-projects/swipeassist/internal/persistence"
	vconfig "github.com/vd09-projects/vision-traits/config"
	"github.com/vd09-projects/vision-traits/traits"
)

// PersistingVisionExtractor wraps an Extractor and records LLM requests/responses via an LLMPersister.
type PersistingVisionExtractor struct {
	inner        Extractor
	llm          persistence.LLMPersister
	app          domain.AppName
	behaviourCfg vconfig.Config
	personaCfg   vconfig.Config
}

func NewPersistingExtractor(eCfg *ExtractorConfig, llm persistence.LLMPersister, app domain.AppName) (Extractor, error) {
	behaviourCfg, err := vconfig.Load(eCfg.BehaviourCfgPath)
	if err != nil {
		return nil, err
	}
	eCfg.BehaviourCfg = &behaviourCfg

	personaCfg, err := vconfig.Load(eCfg.PersonaCfgPath)
	if err != nil {
		return nil, err
	}
	eCfg.PersonaCfg = &personaCfg

	ext, err := NewVisionExtractor(eCfg)
	if err != nil {
		return nil, err
	}

	return &PersistingVisionExtractor{
		inner:        ext,
		llm:          llm,
		app:          app,
		behaviourCfg: behaviourCfg,
		personaCfg:   personaCfg,
	}, nil
}

// ExtractBehaviour persists the request/response and returns the extracted behaviour traits.
func (p *PersistingVisionExtractor) ExtractBehaviour(ctx context.Context, profileKey string, mediaPaths []string) (*domain.BehaviourTraits, error) {
	req, err := p.llm.CreateRequest(ctx, persistence.CreateLLMRequestInput{
		Kind:         dbgen.ExtractionKindBehaviour,
		ProfileKey:   profileKey,
		App:          p.app,
		TemplatePath: p.behaviourCfg.Prompt.TemplatePath,
		// PromptText:   p.behaviourPrompt.PromptText,
		// Vars:         p.behaviourPrompt.Vars,
		Model: p.behaviourCfg.Ollama.Model,
		Media: mediaItemsFromPaths(mediaPaths),
	})
	if err != nil {
		return nil, err
	}

	res, err := p.inner.ExtractBehaviour(ctx, profileKey, mediaPaths)
	if err != nil {
		msg := err.Error()
		_ = p.llm.MarkRequestStatus(ctx, req.ID, persistence.LLMRequestStatusFailed, &msg, time.Now())
		return nil, err
	}
	if res != nil {
		if _, err := p.llm.SaveBehaviourResponse(ctx, req.ID, *res, nil); err != nil {
			return res, err
		}
	}
	if err := p.llm.MarkRequestStatus(ctx, req.ID, persistence.LLMRequestStatusSucceeded, nil, time.Now()); err != nil {
		return res, err
	}
	return res, nil
}

// ExtractPersona persists the request/response and returns the persona bundle.
func (p *PersistingVisionExtractor) ExtractPhotoPersona(ctx context.Context, profileKey string, mediaPaths []string) (*traits.ExtractedTraits, error) {
	req, err := p.llm.CreateRequest(ctx, persistence.CreateLLMRequestInput{
		Kind:         dbgen.ExtractionKindPhotoPersona,
		ProfileKey:   profileKey,
		App:          p.app,
		TemplatePath: p.personaCfg.Prompt.TemplatePath,
		// PromptText:   p.personaPrompt.PromptText,
		// Vars:         p.personaPrompt.Vars,
		Model: p.personaCfg.Ollama.Model,
		Media: mediaItemsFromPaths(mediaPaths),
	})
	if err != nil {
		return nil, err
	}

	traitsByPhoto, err := p.inner.ExtractPhotoPersona(ctx, profileKey, mediaPaths)
	if err != nil {
		msg := err.Error()
		_ = p.llm.MarkRequestStatus(ctx, req.ID, persistence.LLMRequestStatusFailed, &msg, time.Now())
		return nil, err
	}
	return traitsByPhoto, nil
}

func mediaItemsFromPaths(paths []string) []persistence.MediaItem {
	items := make([]persistence.MediaItem, 0, len(paths))
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		items = append(items, persistence.MediaItem{URI: p})
	}
	return items
}
