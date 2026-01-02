package policies

import (
	"context"
	"fmt"
	"sync"

	"github.com/vd09-projects/swipeassist/domain"
)

type QACyclePolicyConfig struct {
	LikesBeforePass int // 2
	ScoreLike       int // e.g. 70
	ScorePass       int // e.g. 40
}

func DefaultQACyclePolicyConfig() QACyclePolicyConfig {
	return QACyclePolicyConfig{
		LikesBeforePass: 3,
		ScoreLike:       70,
		ScorePass:       40,
	}
}

// QACyclePolicy implements:
// - if QA has >= 2 questions: Like, Like, Pass repeating
// - if QA has < 2 questions: Pass + reset cycle
type QACyclePolicy struct {
	cfg QACyclePolicyConfig

	mu        sync.Mutex
	likeCount int
}

func NewQACyclePolicy(cfg QACyclePolicyConfig) *QACyclePolicy {
	return &QACyclePolicy{cfg: cfg}
}

func (p *QACyclePolicy) Name() string { return "qa_cycle_v1" }

func (p *QACyclePolicy) Decide(ctx context.Context, dc *DecisionContext) (*Decision, error) {
	if dc == nil {
		return nil, fmt.Errorf("DecisionContext is nil")
	}
	qCount := countQuestions(dc.BehaviourTraits)

	// Reset rule: <2 questions => Pass + reset
	if qCount < 2 {
		p.mu.Lock()
		p.likeCount = 0
		p.mu.Unlock()

		return &Decision{
			App: dc.App,
			Action: domain.AppAction{
				Kind: domain.AppActionPass,
			},
			Score:      p.cfg.ScorePass,
			Reason:     "Less than 2 Q&A questions detected; passing and restarting cycle.",
			PolicyName: p.Name(),
		}, nil
	}

	// Cycle: Like Like Pass
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.likeCount < p.cfg.LikesBeforePass {
		p.likeCount++
		return &Decision{
			App: dc.App,
			Action: domain.AppAction{
				Kind: domain.AppActionLike,
			},
			Score:      p.cfg.ScoreLike,
			Reason:     "Q&A has at least 2 questions; following cycle: like, like, pass.",
			PolicyName: p.Name(),
		}, nil
	}

	// Pass step then reset.
	p.likeCount = 0
	return &Decision{
		App: dc.App,
		Action: domain.AppAction{
			Kind: domain.AppActionPass,
		},
		Score:      p.cfg.ScorePass,
		Reason:     "Q&A has at least 2 questions; cycle reached pass step (like, like, pass).",
		PolicyName: p.Name(),
	}, nil
}

func countQuestions(bt *domain.BehaviourTraits) int {
	if bt == nil || bt.QASections == nil || bt.QASections.QA == nil {
		return 0
	}
	return len(bt.QASections.QA)
}
