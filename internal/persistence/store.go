package persistence

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/vd09-projects/swipeassist/decisionengine/policies"
	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/internal/dbgen"
)

// Store persists profile-level artifacts such as behaviour traits and decisions.
// Call Close when finished.
type Store interface {
	SaveBehaviour(ctx context.Context, profileKey string, app domain.AppName, traits *domain.BehaviourTraits) error
	SaveDecision(ctx context.Context, profileKey string, decision *policies.Decision) error
	Close(ctx context.Context) error
}

// NewStore creates a Store backed by Postgres when dbURL is set, or a no-op store otherwise.
func NewStore(ctx context.Context, dbURL string) (Store, error) {
	if dbURL == "" {
		return NoopStore{}, nil
	}
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	return &DBStore{
		conn:    conn,
		queries: dbgen.New(conn),
	}, nil
}

// NoopStore drops all writes and keeps the pipeline running when persistence is disabled.
type NoopStore struct{}

func (NoopStore) SaveBehaviour(_ context.Context, _ string, _ domain.AppName, _ *domain.BehaviourTraits) error {
	return nil
}
func (NoopStore) SaveDecision(_ context.Context, _ string, _ *policies.Decision) error { return nil }
func (NoopStore) Close(_ context.Context) error                                        { return nil }

// DBStore persists entities through sqlc-generated queries.
type DBStore struct {
	conn    *pgx.Conn
	queries *dbgen.Queries
}

func (s *DBStore) SaveBehaviour(ctx context.Context, profileKey string, app domain.AppName, traits *domain.BehaviourTraits) error {
	if traits == nil {
		return nil
	}
	_, err := s.queries.InsertBehaviourTraits(ctx, dbgen.InsertBehaviourTraitsParams{
		ProfileKey: stringPtr(profileKey),
		App:        app,
		Traits:     *traits,
	})
	return err
}

func (s *DBStore) SaveDecision(ctx context.Context, profileKey string, decision *policies.Decision) error {
	if decision == nil {
		return nil
	}
	var msgPtr *string
	if decision.Action.Message != "" {
		msgPtr = stringPtr(decision.Action.Message)
	}
	_, err := s.queries.InsertDecision(ctx, dbgen.InsertDecisionParams{
		ProfileKey:    stringPtr(profileKey),
		App:           decision.App,
		PolicyName:    string(decision.PolicyName),
		ActionKind:    decision.Action.Kind,
		ActionMessage: msgPtr,
		Score:         decision.Score,
		Reason:        decision.Reason,
	})
	return err
}

func (s *DBStore) Close(ctx context.Context) error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close(ctx)
}

func stringPtr(s string) *string { return &s }
