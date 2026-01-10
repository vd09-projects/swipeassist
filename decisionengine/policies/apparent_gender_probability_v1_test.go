package policies

import (
	"context"
	"strings"
	"testing"

	"github.com/vd09-projects/swipeassist/domain"
)

type fixedRand struct {
	values []int
	idx    int
}

func (r *fixedRand) Intn(n int) int {
	if n <= 0 {
		panic("Intn called with non-positive n")
	}
	if len(r.values) == 0 {
		return 0
	}
	v := r.values[r.idx%len(r.values)]
	r.idx++
	if v < 0 {
		return 0
	}
	return v % n
}

func TestApparentGenderProbabilityPolicy_FemaleUsesConfiguredWeights(t *testing.T) {
	cfg := ApparentGenderProbabilityPolicyConfig{
		FemaleLikeWeight:  1,
		FemalePassWeight:  1,
		UnknownLikeWeight: 1,
		UnknownPassWeight: 1,
		ScoreLike:         70,
		ScorePass:         40,
	}
	pol, err := newApparentGenderProbabilityPolicy(cfg, &fixedRand{values: []int{0}})
	if err != nil {
		t.Fatalf("unexpected error creating policy: %v", err)
	}

	dec, err := pol.Decide(context.Background(), &DecisionContext{
		App:          domain.Bumble,
		PhotoPersona: personaWithGender("female"),
	})
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if dec.Action.Kind != domain.AppActionLike {
		t.Fatalf("expected LIKE, got %s", dec.Action.Kind)
	}
	if dec.Score != cfg.ScoreLike {
		t.Fatalf("unexpected score: got %d want %d", dec.Score, cfg.ScoreLike)
	}
	if dec.PolicyName != pol.Name() {
		t.Fatalf("unexpected policy name: %s", dec.PolicyName)
	}
	if !strings.Contains(dec.Reason, "female") {
		t.Fatalf("expected female reason, got %q", dec.Reason)
	}
}

func TestApparentGenderProbabilityPolicy_MaleAlwaysPasses(t *testing.T) {
	pol, err := newApparentGenderProbabilityPolicy(DefaultApparentGenderProbabilityConfig(), &fixedRand{values: []int{0}})
	if err != nil {
		t.Fatalf("unexpected error creating policy: %v", err)
	}

	dec, err := pol.Decide(context.Background(), &DecisionContext{
		App:          domain.Bumble,
		PhotoPersona: personaWithGender("male"),
	})
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if dec.Action.Kind != domain.AppActionPass {
		t.Fatalf("expected PASS for male, got %s", dec.Action.Kind)
	}
	if dec.Score != pol.cfg.ScorePass {
		t.Fatalf("unexpected pass score: got %d want %d", dec.Score, pol.cfg.ScorePass)
	}
	if !strings.Contains(dec.Reason, "male") {
		t.Fatalf("expected male reason, got %q", dec.Reason)
	}
}

func TestApparentGenderProbabilityPolicy_FallbackWhenUnknown(t *testing.T) {
	cfg := ApparentGenderProbabilityPolicyConfig{
		FemaleLikeWeight:  1,
		FemalePassWeight:  1,
		UnknownLikeWeight: 1,
		UnknownPassWeight: 1,
		ScoreLike:         70,
		ScorePass:         40,
	}
	pol, err := newApparentGenderProbabilityPolicy(cfg, &fixedRand{values: []int{1}})
	if err != nil {
		t.Fatalf("unexpected error creating policy: %v", err)
	}

	dec, err := pol.Decide(context.Background(), &DecisionContext{
		App:          domain.Bumble,
		PhotoPersona: nil,
	})
	if err != nil {
		t.Fatalf("Decide returned error: %v", err)
	}
	if dec.Action.Kind != domain.AppActionPass {
		t.Fatalf("expected PASS for unknown gender, got %s", dec.Action.Kind)
	}
	if !strings.Contains(dec.Reason, "fallback") {
		t.Fatalf("expected fallback reason, got %q", dec.Reason)
	}
}

func personaWithGender(g string) *domain.PhotoPersonaBundle {
	if g == "" {
		return nil
	}
	return &domain.PhotoPersonaBundle{
		Images: map[string]domain.PhotoPersonaProfile{
			"image_1": {
				Traits: map[string][]string{
					"apparent_gender": {g},
				},
			},
		},
	}
}
