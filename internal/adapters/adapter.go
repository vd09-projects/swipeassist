package adapters

import (
	"context"

	"github.com/vd09-projects/swipeassist/internal/domain"
)

type Action string

const (
	ActionPass      Action = "PASS"
	ActionLike      Action = "LIKE"
	ActionSuperLike Action = "SUPERLIKE"
	ActionMessage   Action = "MESSAGE"
)

type AppCapabilities struct {
	ActionsSupported []Action `json:"actions_supported"`
	MessageOptional  bool     `json:"message_optional"`
}

type ImageRef struct {
	Path   string
	Sha256 string // "sha256:<hex>"
	Bytes  []byte // loaded content
}

type ExtractInput struct {
	Images []ImageRef
	Hints  map[string]string // optional: age, location, etc.
}

type ExtractOutput struct {
	Profile domain.Profile
	Report  ExtractionReport
}

type ExtractionReport struct {
	AdapterName       string   `json:"adapter_name"`
	OverallConfidence float64  `json:"overall_confidence"`
	Warnings          []string `json:"warnings,omitempty"`
	Errors            []string `json:"errors,omitempty"`
}

type AppAdapter interface {
	Name() string
	Capabilities() AppCapabilities
	Extract(ctx context.Context, in ExtractInput) (ExtractOutput, error)
}
