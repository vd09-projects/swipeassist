package analytics

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
)

type Snapshot struct {
	App       domain.AppName      `json:"app"`
	Policy    any                 `json:"policy"`
	Status    Status              `json:"status"`
	Error     string              `json:"error,omitempty"`
	StartedAt time.Time           `json:"started_at"`
	Runtime   time.Duration       `json:"runtime"`
	Counters  map[string]int64    `json:"counters"`
	Timers    map[string]TimerAgg `json:"timers,omitempty"`
}

type Sink interface {
	Emit(ctx context.Context, snap Snapshot) error
}

// LogSink prints a stable, readable line (sorted keys).
type LogSink struct{}

func NewLogSink() *LogSink { return &LogSink{} }

func (s *LogSink) Emit(_ context.Context, snap Snapshot) error {
	keys := make([]string, 0, len(snap.Counters))
	for k := range snap.Counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// You can keep it structured (JSON) or key=value.
	b, _ := json.Marshal(snap)
	log.Printf("analytics: %s", string(b))

	// If you prefer key=value, build it using sorted keys.
	_ = keys
	return nil
}
