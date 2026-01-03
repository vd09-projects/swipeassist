package analytics

import (
	"sync"
	"time"
)

type Store struct {
	mu       sync.Mutex
	counters map[string]int64
	timers   map[string]TimerAgg
}

type TimerAgg struct {
	Count int64
	Total time.Duration
	Min   time.Duration
	Max   time.Duration
}

func NewStore() *Store {
	return &Store{
		counters: make(map[string]int64),
		timers:   make(map[string]TimerAgg),
	}
}

func (s *Store) Inc(name string, n int64) {
	if n == 0 || name == "" {
		return
	}
	s.mu.Lock()
	s.counters[name] += n
	s.mu.Unlock()
}

func (s *Store) Observe(name string, d time.Duration) {
	if name == "" {
		return
	}
	s.mu.Lock()
	agg := s.timers[name]
	agg.Count++
	agg.Total += d
	if agg.Count == 1 || d < agg.Min {
		agg.Min = d
	}
	if d > agg.Max {
		agg.Max = d
	}
	s.timers[name] = agg
	s.mu.Unlock()
}

func (s *Store) Counters() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		out[k] = v
	}
	return out
}

func (s *Store) Timers() map[string]TimerAgg {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]TimerAgg, len(s.timers))
	for k, v := range s.timers {
		out[k] = v
	}
	return out
}
