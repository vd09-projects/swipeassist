package decisionengine

import (
	"context"

	"github.com/vd09-projects/swipeassist/decisionengine/policies"
)

type DecisionEngine struct {
	reg        *Registry
	policyName policies.PolicyName
}

func NewDecisionEngine(reg *Registry, policyName policies.PolicyName) *DecisionEngine {
	return &DecisionEngine{
		reg:        reg,
		policyName: policyName,
	}
}

func (e *DecisionEngine) Decide(ctx context.Context, dc *policies.DecisionContext) (*policies.Decision, error) {
	p, err := e.reg.Resolve(e.policyName)
	if err != nil {
		return nil, err
	}
	return p.Decide(ctx, dc)
}
