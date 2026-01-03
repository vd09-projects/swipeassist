package decisionengine

import (
	"context"
	"testing"

	"github.com/vd09-projects/swipeassist/decisionengine/policies"
	"github.com/vd09-projects/swipeassist/domain"
)

type fakePolicy struct {
	t        *testing.T
	name     policies.PolicyName
	decision *policies.Decision
	err      error

	callCount int
	lastCtx   *policies.DecisionContext
}

func (f *fakePolicy) Name() policies.PolicyName { return f.name }

func (f *fakePolicy) Decide(ctx context.Context, dc *policies.DecisionContext) (*policies.Decision, error) {
	f.callCount++
	if ctx == nil {
		f.t.Fatalf("expected context, got nil")
	}
	f.lastCtx = dc
	return f.decision, f.err
}

func TestRegistryResolveRegistered(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	p := &fakePolicy{t: t, name: policies.PolicyName("fake")}
	reg.Register(p.name, p)

	got, err := reg.Resolve(p.name)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != p {
		t.Fatalf("Resolve returned unexpected policy: %#v", got)
	}
}

func TestRegistryResolveMissing(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	got, err := reg.Resolve(policies.PolicyName("missing"))
	if err == nil {
		t.Fatalf("expected error when resolving missing policy name")
	}
	if got != nil {
		t.Fatalf("expected nil policy when missing, got %#v", got)
	}
}

func TestDecisionEngineUsesPolicyFromRegistry(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	wantDecision := &policies.Decision{
		App: domain.Bumble,
		Action: domain.AppAction{
			Kind: domain.AppActionLike,
		},
		Score:      90,
		Reason:     "ok",
		PolicyName: "fake",
	}

	fp := &fakePolicy{t: t, name: policies.PolicyName("fake"), decision: wantDecision}
	reg.Register(fp.name, fp)

	engine := NewDecisionEngine(reg, fp.name)
	dc := &policies.DecisionContext{
		App:        domain.Bumble,
		ProfileKey: "profile-key",
	}

	got, err := engine.Decide(context.Background(), dc)
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if got != wantDecision {
		t.Fatalf("Decide returned unexpected decision: %#v", got)
	}
	if fp.callCount != 1 {
		t.Fatalf("expected policy decide called once, got %d", fp.callCount)
	}
	if fp.lastCtx != dc {
		t.Fatalf("policy received unexpected decision context: %#v", fp.lastCtx)
	}
}

func TestDecisionEnginePropagatesResolveError(t *testing.T) {
	t.Parallel()

	engine := NewDecisionEngine(NewRegistry(), policies.PolicyName("missing"))
	_, err := engine.Decide(context.Background(), &policies.DecisionContext{App: domain.Bumble})
	if err == nil {
		t.Fatalf("expected error when no policy registered")
	}
}
