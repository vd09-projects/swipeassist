package policies

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"github.com/vd09-projects/swipeassist/domain"
)

func TestProbabilisticRatioPolicyValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewProbabilisticRatioPolicy(ProbabilisticRatioPolicyConfig{LikeWeight: 0, PassWeight: 1}); err == nil {
		t.Fatalf("expected error when like weight is zero")
	}
}

func TestProbabilisticRatioPolicyNilDecisionContext(t *testing.T) {
	t.Parallel()

	p, err := NewProbabilisticRatioPolicy(DefaultProbabilisticRatioPolicyConfig())
	if err != nil {
		t.Fatalf("NewProbabilisticRatioPolicy returned error: %v", err)
	}

	if _, err := p.Decide(context.Background(), nil); err == nil {
		t.Fatalf("expected error for nil decision context")
	}
}

func TestProbabilisticRatioPolicyRespectsRatio(t *testing.T) {
	t.Parallel()

	cfg := ProbabilisticRatioPolicyConfig{
		LikeWeight: 3,
		PassWeight: 2,
		ScoreLike:  80,
		ScorePass:  30,
	}
	seededRand := rand.New(rand.NewSource(1))

	p, err := newProbabilisticRatioPolicy(cfg, seededRand)
	if err != nil {
		t.Fatalf("newProbabilisticRatioPolicy returned error: %v", err)
	}

	ctx := context.Background()
	dc := &DecisionContext{App: domain.Bumble}

	const trials = 1000
	likeCount := 0
	passCount := 0

	for i := 0; i < trials; i++ {
		decision, err := p.Decide(ctx, dc)
		if err != nil {
			t.Fatalf("Decide returned error: %v", err)
		}
		if decision.PolicyName != p.Name() {
			t.Fatalf("expected policy name %q, got %q", p.Name(), decision.PolicyName)
		}
		switch decision.Action.Kind {
		case domain.AppActionLike:
			likeCount++
			if decision.Score != cfg.ScoreLike {
				t.Fatalf("expected like score %d, got %d", cfg.ScoreLike, decision.Score)
			}
		case domain.AppActionPass:
			passCount++
			if decision.Score != cfg.ScorePass {
				t.Fatalf("expected pass score %d, got %d", cfg.ScorePass, decision.Score)
			}
		default:
			t.Fatalf("unexpected action kind: %s", decision.Action.Kind)
		}
	}

	total := likeCount + passCount
	if total != trials {
		t.Fatalf("expected %d total decisions, got %d", trials, total)
	}

	expectedRatio := float64(cfg.LikeWeight) / float64(cfg.LikeWeight+cfg.PassWeight)
	actualRatio := float64(likeCount) / float64(trials)
	if math.Abs(actualRatio-expectedRatio) > 0.05 {
		t.Fatalf("like ratio deviated too much: expected ~%.2f got %.2f", expectedRatio, actualRatio)
	}
}
