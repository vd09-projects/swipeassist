package domain

type BehaviourTraits struct {
	GlobalConfidence int `json:"global_confidence"`

	RawText *RawTextBlock `json:"raw_text"` // from ui_free_text

	QASections *QASectionsBlock `json:"qa_sections"` // from ui_sections

	ProfileTags *ProfileTagsBlock `json:"profile_tags"` // from ui_tags (better name than just "tags")
}

type RawTextBlock struct {
	Confidence int      `json:"confidence"`
	Lines      []string `json:"lines"` // preserve order
}

type QASectionsBlock struct {
	Confidence int                 `json:"confidence"`
	QA         map[string][]string `json:"qa"` // question -> answers
}

type ProfileTagsBlock struct {
	Confidence int                 `json:"confidence"`
	Tags       map[string][]string `json:"tags"`          // tag_key -> values
	Raw        []string            `json:"raw,omitempty"` // optional passthrough from ui_tags.signals_by_key["raw"]
}

// PhotoPersonaBundle is the per-image output container.
// Keys are "image_1", "image_2", ...
type PhotoPersonaBundle struct {
	Images map[string]PhotoPersonaProfile `json:"images"`
}

// PhotoPersonaProfile is the persona derived from ONE image.
type PhotoPersonaProfile struct {
	Tags       []string            `json:"tags,omitempty"`       // flat union of signals
	Traits     map[string][]string `json:"traits,omitempty"`     // trait_key -> signals
	Statements []string            `json:"statements,omitempty"` // pulled from summaries (optional)
}
