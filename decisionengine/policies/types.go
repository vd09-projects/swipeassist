package policies

import "github.com/vd09-projects/swipeassist/domain"

// Recommendation is the generic output from the chooser engine.
type Decision struct {
	App        domain.AppName   `json:"app"`
	Action     domain.AppAction `json:"action"`
	Score      int              `json:"score"` // 0-100 (keep; optional now)
	Reason     string           `json:"reason"`
	PolicyName string           `json:"policy_name"`
}

// DecisionContext replaces the old "Input".
// It’s explicit, stable, and future-proof.
type DecisionContext struct {
	App domain.AppName `json:"app"`

	BehaviourTraits *domain.BehaviourTraits    `json:"behaviour_traits,omitempty"`
	PhotoPersona    *domain.PhotoPersonaBundle `json:"photo_persona_bundle,omitempty"`

	// optional: useful for logs, replay, debugging
	ProfileKey string `json:"profile_key,omitempty"`
}
