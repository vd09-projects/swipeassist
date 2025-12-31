package domain

import "time"

type MediaType string
const (
	MediaTypeImage MediaType = "image"
)

type TextKind string
const (
	TextKindPrompt    TextKind = "prompt"
	TextKindBio       TextKind = "bio"
	TextKindInterest  TextKind = "interest"
	TextKindMetadata  TextKind = "metadata"
)

type MediaItem struct {
	MediaID         string   `json:"media_id"`
	Type            MediaType `json:"type"`
	SourceImageRef  string   `json:"source_image_ref"` // sha256:<hex>
	VisionSummary   string   `json:"vision_summary,omitempty"`
	VisionConfidence float64 `json:"vision_confidence,omitempty"`
}

type TextBlock struct {
	Kind       TextKind `json:"kind"`
	Label      *string  `json:"label,omitempty"`
	Content    string   `json:"content"`
	Confidence float64  `json:"confidence"`
}

type Metadata struct {
	Age        *int    `json:"age,omitempty"`
	Location   *string `json:"location,omitempty"`
	DistanceKM *int    `json:"distance_km,omitempty"`
}

type ExtractionSummary struct {
	OverallConfidence float64  `json:"overall_confidence"`
	MissingFields     []string `json:"missing_fields,omitempty"`
	Notes             []string `json:"notes,omitempty"`
}

type Profile struct {
	SchemaVersion string            `json:"schema_version"`
	ProfileID     string            `json:"profile_id"`
	AppSource     string            `json:"app_source"`
	CapturedAt    time.Time         `json:"captured_at"`
	Media         []MediaItem        `json:"media"`
	TextBlocks    []TextBlock        `json:"text_blocks"`
	Metadata      Metadata           `json:"metadata"`
	Extraction    ExtractionSummary  `json:"extraction"`
}