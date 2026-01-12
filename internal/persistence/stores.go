package persistence

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/vd09-projects/swipeassist/internal/dbgen"
)

// Stores groups the decision/behaviour store with the LLM request store.
type Stores struct {
	Decisions Store
	LLM       LLMPersister
	closeFn   func(ctx context.Context) error
}

// NewStores returns no-op stores when dbURL is empty, else a Postgres-backed pair.
func NewStores(ctx context.Context, dbURL string) (*Stores, error) {
	if dbURL == "" {
		return &Stores{
			Decisions: NoopStore{},
			LLM:       NewNoopLLMStore(),
			closeFn:   func(context.Context) error { return nil },
		}, nil
	}

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	queries := dbgen.New(conn)

	return &Stores{
		Decisions: &DBStore{conn: conn, queries: queries},
		LLM:       NewLLMStore(queries),
		closeFn: func(c context.Context) error {
			return conn.Close(c)
		},
	}, nil
}

// Close closes the underlying DB connection when present.
func (s *Stores) Close(ctx context.Context) error {
	if s == nil || s.closeFn == nil {
		return nil
	}
	return s.closeFn(ctx)
}
