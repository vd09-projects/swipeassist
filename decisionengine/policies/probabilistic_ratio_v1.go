package policies

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
)

// ProbabilisticRatioPolicyConfig sets the like:pass weighting for the random selector.
// The actual probability is LikeWeight / (LikeWeight + PassWeight).
type ProbabilisticRatioPolicyConfig struct {
	LikeWeight int // relative weight for LIKE
	PassWeight int // relative weight for PASS
	ScoreLike  int // score to emit when liking
	ScorePass  int // score to emit when passing
}

func DefaultProbabilisticRatioPolicyConfig() ProbabilisticRatioPolicyConfig {
	return ProbabilisticRatioPolicyConfig{
		LikeWeight: 5,
		PassWeight: 2,
		ScoreLike:  70,
		ScorePass:  40,
	}
}

type ProbabilisticRatioPolicy struct {
	cfg  ProbabilisticRatioPolicyConfig
	rand *rand.Rand
	mu   sync.Mutex
}

const ProbabilisticRatioPolicyName PolicyName = "probabilistic_ratio_v1"

func NewProbabilisticRatioPolicy(cfg ProbabilisticRatioPolicyConfig) (*ProbabilisticRatioPolicy, error) {
	return newProbabilisticRatioPolicy(cfg, rand.New(rand.NewSource(time.Now().UnixNano())))
}

func newProbabilisticRatioPolicy(cfg ProbabilisticRatioPolicyConfig, r *rand.Rand) (*ProbabilisticRatioPolicy, error) {
	if cfg.LikeWeight <= 0 || cfg.PassWeight <= 0 {
		return nil, fmt.Errorf("like and pass weights must both be > 0 (got like=%d, pass=%d)", cfg.LikeWeight, cfg.PassWeight)
	}
	if r == nil {
		return nil, fmt.Errorf("rand source is nil")
	}
	return &ProbabilisticRatioPolicy{
		cfg:  cfg,
		rand: r,
	}, nil
}

func (p *ProbabilisticRatioPolicy) Name() PolicyName { return ProbabilisticRatioPolicyName }

func (p *ProbabilisticRatioPolicy) Decide(ctx context.Context, dc *DecisionContext) (*Decision, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if dc == nil {
		return nil, fmt.Errorf("DecisionContext is nil")
	}

	total := p.cfg.LikeWeight + p.cfg.PassWeight

	p.mu.Lock()
	roll := p.rand.Intn(total)
	p.mu.Unlock()

	action := domain.AppAction{Kind: domain.AppActionPass}
	score := p.cfg.ScorePass
	if roll < p.cfg.LikeWeight {
		action.Kind = domain.AppActionLike
		score = p.cfg.ScoreLike
	}

	return &Decision{
		App:        dc.App,
		Action:     action,
		Score:      score,
		Reason:     fmt.Sprintf("Randomized decision using like:pass ratio %d:%d", p.cfg.LikeWeight, p.cfg.PassWeight),
		PolicyName: p.Name(),
	}, nil
}
