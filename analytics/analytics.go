package analytics

import (
	"context"
	"sync"
	"time"

	"github.com/vd09-projects/swipeassist/domain"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

// Session is one execution of your engine for a specific app + policy.
type Session struct {
	start  time.Time
	app    domain.AppName
	policy any

	store *Store
	sinks []Sink

	mu     sync.Mutex
	closed bool
}

type Config struct {
	App    domain.AppName
	Policy any
	Sinks  []Sink
}

func NewSession(cfg Config) *Session {
	return &Session{
		start:  time.Now(),
		app:    cfg.App,
		policy: cfg.Policy,
		store:  NewStore(),
		sinks:  cfg.Sinks,
	}
}

// ---- High-level helpers (easy callsites) ----

func (r *Session) ProfileAttempt()                      { r.store.Inc("profile_attempts", 1) }
func (r *Session) ProfileComplete()                     { r.store.Inc("profiles_completed", 1) }
func (r *Session) AddScreenshots(n int)                 { r.store.Inc("screenshots", int64(n)) }
func (r *Session) RecordAction(a domain.AppActionType)  { r.store.Inc("action."+string(a), 1) }
func (r *Session) Inc(name string, n int64)             { r.store.Inc(name, n) }
func (r *Session) Observe(name string, d time.Duration) { r.store.Observe(name, d) }

// Close finalizes the Session and emits to sinks once.
func (r *Session) Close(ctx context.Context, err error) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	r.mu.Unlock()

	status := StatusSuccess
	errMsg := ""
	if err != nil {
		status = StatusError
		errMsg = err.Error()
	}

	snap := Snapshot{
		App:       r.app,
		Policy:    r.policy,
		Status:    status,
		Error:     errMsg,
		StartedAt: r.start,
		Runtime:   time.Since(r.start),
		Counters:  r.store.Counters(),
		Timers:    r.store.Timers(),
	}

	for _, s := range r.sinks {
		_ = s.Emit(ctx, snap) // best-effort; optionally collect errors
	}
}
