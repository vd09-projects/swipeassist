package decisionengine

import (
	"context"

	"github.com/vd09-projects/swipeassist/decisionengine/policies"
)

// Policy is a pluggable "brain" strategy.
type Policy interface {
	Name() policies.PolicyName
	Decide(ctx context.Context, dc *policies.DecisionContext) (*policies.Decision, error)
}
