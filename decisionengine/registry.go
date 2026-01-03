package decisionengine

import (
	"fmt"

	"github.com/vd09-projects/swipeassist/decisionengine/policies"
)

type Registry struct {
	byName map[policies.PolicyName]Policy
}

func NewRegistry() *Registry {
	reg := &Registry{byName: make(map[policies.PolicyName]Policy)}
	if pol, err := policies.NewProbabilisticRatioPolicy(policies.DefaultProbabilisticRatioPolicyConfig()); err == nil {
		reg.Register(policies.ProbabilisticRatioPolicyName, pol)
	} else {
		fmt.Println("ERROR: Error in creating Probabilistic Ratio Policy")
		// return nil, fmt.Errorf("Error in creating Probabilistic Ratio Policy")
	}
	reg.Register(policies.QACyclePolicyName, policies.NewQACyclePolicy(policies.DefaultQACyclePolicyConfig()))
	return reg
}

func (r *Registry) Register(name policies.PolicyName, p Policy) {
	r.byName[name] = p
}

func (r *Registry) Resolve(name policies.PolicyName) (Policy, error) {
	p, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("no decision policy registered for name=%s", name)
	}
	return p, nil
}
