package policies

import (
	"context"
	"strings"
	"testing"

	"github.com/vd09-projects/swipeassist/domain"
)

func TestQACyclePolicyNilDecisionContext(t *testing.T) {
	t.Parallel()

	p := NewQACyclePolicy(DefaultQACyclePolicyConfig())
	if _, err := p.Decide(context.Background(), nil); err == nil {
		t.Fatalf("expected error for nil decision context")
	}
}

func TestQACyclePolicyLessThanTwoQuestionsResetsCycle(t *testing.T) {
	t.Parallel()

	cfg := QACyclePolicyConfig{
		LikesBeforePass: 2,
		ScoreLike:       70,
		ScorePass:       40,
	}
	p := NewQACyclePolicy(cfg)

	withQuestions := &DecisionContext{
		App: domain.Bumble,
		BehaviourTraits: &domain.BehaviourTraits{
			QASections: &domain.QASectionsBlock{
				QA: map[string][]string{
					"q1": {"a"},
					"q2": {"b"},
				},
			},
		},
	}
	tooFewQuestions := &DecisionContext{
		App: domain.Bumble,
		BehaviourTraits: &domain.BehaviourTraits{
			QASections: &domain.QASectionsBlock{
				QA: map[string][]string{
					"only": {"a"},
				},
			},
		},
	}

	firstLike, err := p.Decide(context.Background(), withQuestions)
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if firstLike.Action.Kind != domain.AppActionLike || firstLike.Score != cfg.ScoreLike {
		t.Fatalf("expected like on first call, got %#v", firstLike)
	}

	passDecision, err := p.Decide(context.Background(), tooFewQuestions)
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if passDecision.Action.Kind != domain.AppActionPass || passDecision.Score != cfg.ScorePass {
		t.Fatalf("expected pass when fewer than 2 questions, got %#v", passDecision)
	}
	if !strings.Contains(passDecision.Reason, "Less than 2 Q&A questions detected") {
		t.Fatalf("unexpected reason: %q", passDecision.Reason)
	}

	afterReset, err := p.Decide(context.Background(), withQuestions)
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if afterReset.Action.Kind != domain.AppActionLike {
		t.Fatalf("expected cycle to restart after reset, got %#v", afterReset)
	}
}

func TestQACyclePolicyLikeCycle(t *testing.T) {
	t.Parallel()

	cfg := QACyclePolicyConfig{
		LikesBeforePass: 2,
		ScoreLike:       70,
		ScorePass:       40,
	}
	p := NewQACyclePolicy(cfg)

	ctx := context.Background()
	dc := &DecisionContext{
		App: domain.Bumble,
		BehaviourTraits: &domain.BehaviourTraits{
			QASections: &domain.QASectionsBlock{
				QA: map[string][]string{
					"q1": {"a"},
					"q2": {"b"},
					"q3": {"c"},
				},
			},
		},
	}

	for i := 0; i < cfg.LikesBeforePass+1; i++ {
		decision, err := p.Decide(ctx, dc)
		if err != nil {
			t.Fatalf("Decide returned error: %v", err)
		}
		if i < cfg.LikesBeforePass {
			if decision.Action.Kind != domain.AppActionLike || decision.Score != cfg.ScoreLike {
				t.Fatalf("expected like during cycle, got %#v", decision)
			}
		} else {
			if decision.Action.Kind != domain.AppActionPass || decision.Score != cfg.ScorePass {
				t.Fatalf("expected pass after like cycle, got %#v", decision)
			}
		}
		if decision.PolicyName != p.Name() {
			t.Fatalf("expected policy name %q, got %q", p.Name(), decision.PolicyName)
		}
	}

	nextCycle, err := p.Decide(ctx, dc)
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if nextCycle.Action.Kind != domain.AppActionLike {
		t.Fatalf("expected cycle to restart after pass, got %#v", nextCycle)
	}
}
