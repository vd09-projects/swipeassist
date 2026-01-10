package policies

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
)

// ApparentGenderProbabilityPolicyConfig controls the like:pass weighting based on detected apparent gender.
type ApparentGenderProbabilityPolicyConfig struct {
	FemaleLikeWeight  int
	FemalePassWeight  int
	UnknownLikeWeight int
	UnknownPassWeight int
	ScoreLike         int
	ScorePass         int
}

// DefaultApparentGenderProbabilityConfig returns a conservative fallback for unknown gender
// and a like-leaning ratio for detected female profiles.
func DefaultApparentGenderProbabilityConfig() ApparentGenderProbabilityPolicyConfig {
	return ApparentGenderProbabilityPolicyConfig{
		FemaleLikeWeight:  9,
		FemalePassWeight:  2,
		UnknownLikeWeight: 4,
		UnknownPassWeight: 6,
		ScoreLike:         70,
		ScorePass:         40,
	}
}

type genderRand interface {
	Intn(n int) int
}

// ApparentGenderProbabilityPolicy implements:
// - female: randomized like/pass using female weights
// - male: always pass
// - unknown/missing: randomized like/pass using fallback weights
type ApparentGenderProbabilityPolicy struct {
	cfg  ApparentGenderProbabilityPolicyConfig
	rand genderRand
	mu   sync.Mutex
}

const ApparentGenderProbabilityPolicyName PolicyName = "apparent_gender_probability_v1"

func NewApparentGenderProbabilityPolicy(cfg ApparentGenderProbabilityPolicyConfig) (*ApparentGenderProbabilityPolicy, error) {
	return newApparentGenderProbabilityPolicy(cfg, rand.New(rand.NewSource(time.Now().UnixNano())))
}

func newApparentGenderProbabilityPolicy(cfg ApparentGenderProbabilityPolicyConfig, r genderRand) (*ApparentGenderProbabilityPolicy, error) {
	if r == nil {
		return nil, fmt.Errorf("rand source is nil")
	}
	if err := validateApparentGenderConfig(cfg); err != nil {
		return nil, err
	}
	return &ApparentGenderProbabilityPolicy{
		cfg:  cfg,
		rand: r,
	}, nil
}

func validateApparentGenderConfig(cfg ApparentGenderProbabilityPolicyConfig) error {
	if cfg.FemaleLikeWeight <= 0 || cfg.FemalePassWeight <= 0 {
		return fmt.Errorf("female like and pass weights must be > 0 (got like=%d, pass=%d)", cfg.FemaleLikeWeight, cfg.FemalePassWeight)
	}
	if cfg.UnknownLikeWeight <= 0 || cfg.UnknownPassWeight <= 0 {
		return fmt.Errorf("unknown like and pass weights must be > 0 (got like=%d, pass=%d)", cfg.UnknownLikeWeight, cfg.UnknownPassWeight)
	}
	return nil
}

func (p *ApparentGenderProbabilityPolicy) Name() PolicyName {
	return ApparentGenderProbabilityPolicyName
}

func (p *ApparentGenderProbabilityPolicy) Decide(ctx context.Context, dc *DecisionContext) (*Decision, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if dc == nil {
		return nil, fmt.Errorf("DecisionContext is nil")
	}

	gender := extractApparentGender(dc.PhotoPersona)

	switch gender {
	case "male":
		return &Decision{
			App: dc.App,
			Action: domain.AppAction{
				Kind: domain.AppActionPass,
			},
			Score:      p.cfg.ScorePass,
			Reason:     "apparent_gender indicates male; always passing.",
			PolicyName: p.Name(),
		}, nil
	case "female":
		return p.weightedDecision(dc.App, p.cfg.FemaleLikeWeight, p.cfg.FemalePassWeight,
			fmt.Sprintf("apparent_gender indicates female; randomizing with like:pass ratio %d:%d.", p.cfg.FemaleLikeWeight, p.cfg.FemalePassWeight))
	default:
		return p.weightedDecision(dc.App, p.cfg.UnknownLikeWeight, p.cfg.UnknownPassWeight,
			fmt.Sprintf("apparent_gender missing or unknown; fallback like:pass ratio %d:%d.", p.cfg.UnknownLikeWeight, p.cfg.UnknownPassWeight))
	}
}

func (p *ApparentGenderProbabilityPolicy) weightedDecision(app domain.AppName, likeWeight, passWeight int, reason string) (*Decision, error) {
	total := likeWeight + passWeight
	if total <= 0 {
		return nil, fmt.Errorf("invalid weights: like=%d pass=%d", likeWeight, passWeight)
	}

	p.mu.Lock()
	roll := p.rand.Intn(total)
	p.mu.Unlock()

	action := domain.AppAction{Kind: domain.AppActionPass}
	score := p.cfg.ScorePass
	if roll < likeWeight {
		action.Kind = domain.AppActionLike
		score = p.cfg.ScoreLike
	}

	return &Decision{
		App:        app,
		Action:     action,
		Score:      score,
		Reason:     reason,
		PolicyName: p.Name(),
	}, nil
}

func extractApparentGender(bundle *domain.PhotoPersonaBundle) string {
	if bundle == nil || len(bundle.Images) == 0 {
		return ""
	}

	keys := make([]string, 0, len(bundle.Images))
	for k := range bundle.Images {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		profile := bundle.Images[key]
		if g := genderFromTraits(profile.Traits); g != "" {
			return g
		}
	}

	return ""
}

func genderFromTraits(traits map[string][]string) string {
	if len(traits) == 0 {
		return ""
	}
	signals, ok := traits["apparent_gender"]
	if !ok {
		return ""
	}
	for _, s := range signals {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "female":
			return "female"
		case "male":
			return "male"
		}
	}
	return ""
}
