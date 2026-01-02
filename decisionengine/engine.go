package decisionengine

import (
	"context"

	"github.com/vd09-projects/swipeassist/decisionengine/policies"
)

type DecisionEngine struct {
	reg *Registry
}

func NewDecisionEngine(reg *Registry) *DecisionEngine {
	return &DecisionEngine{reg: reg}
}

func (e *DecisionEngine) Decide(ctx context.Context, dc *policies.DecisionContext) (*policies.Decision, error) {
	p, err := e.reg.Resolve(dc.App)
	if err != nil {
		return nil, err
	}
	return p.Decide(ctx, dc)
}
