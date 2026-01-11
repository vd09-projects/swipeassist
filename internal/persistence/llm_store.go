package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/internal/dbgen"
)

type LLMRequestStatus string

const (
	LLMRequestStatusPending   LLMRequestStatus = "pending"
	LLMRequestStatusSucceeded LLMRequestStatus = "succeeded"
	LLMRequestStatusFailed    LLMRequestStatus = "failed"
)

// MediaItem represents one media attachment for a request.
type MediaItem struct {
	URI       string
	MediaType string // optional; defaults to "image"
}

// CreateLLMRequestInput bundles the data needed to persist an LLM request.
type CreateLLMRequestInput struct {
	Kind            dbgen.ExtractionKind
	ProfileKey      string
	App             domain.AppName
	TemplatePath    string
	PromptText      string
	Vars            any // marshalled to JSON; nil allowed
	Model           string
	Status          LLMRequestStatus
	ErrorMessage    *string
	ParentRequestID *int64
	Media           []MediaItem
}

// LLMStore hides the dbgen wiring for storing LLM requests and responses.
type LLMStore struct {
	queries *dbgen.Queries
}

func NewLLMStore(queries *dbgen.Queries) *LLMStore {
	return &LLMStore{queries: queries}
}

// CreateRequest inserts the request row and its media in order, returning the stored request.
func (s *LLMStore) CreateRequest(ctx context.Context, in CreateLLMRequestInput) (dbgen.LlmRequest, error) {
	if s.queries == nil {
		return dbgen.LlmRequest{}, fmt.Errorf("queries is nil")
	}
	if in.Kind == "" {
		return dbgen.LlmRequest{}, fmt.Errorf("kind is required")
	}
	if in.App == "" {
		return dbgen.LlmRequest{}, fmt.Errorf("app is required")
	}
	if in.TemplatePath == "" {
		return dbgen.LlmRequest{}, fmt.Errorf("template_path is required")
	}
	if in.PromptText == "" {
		return dbgen.LlmRequest{}, fmt.Errorf("prompt_text is required")
	}
	if in.Model == "" {
		return dbgen.LlmRequest{}, fmt.Errorf("model is required")
	}

	status := in.Status
	if status == "" {
		status = LLMRequestStatusPending
	}

	var varsJSON []byte
	if in.Vars != nil {
		b, err := json.Marshal(in.Vars)
		if err != nil {
			return dbgen.LlmRequest{}, fmt.Errorf("marshal vars: %w", err)
		}
		varsJSON = b
	}

	req, err := s.queries.InsertLLMRequest(ctx, dbgen.InsertLLMRequestParams{
		Kind:            in.Kind,
		ProfileKey:      nullableString(in.ProfileKey),
		App:             in.App,
		TemplatePath:    in.TemplatePath,
		PromptText:      in.PromptText,
		Vars:            varsJSON,
		Model:           in.Model,
		Status:          string(status),
		ErrorMessage:    in.ErrorMessage,
		ParentRequestID: in.ParentRequestID,
	})
	if err != nil {
		return dbgen.LlmRequest{}, err
	}

	for idx, media := range in.Media {
		position := idx + 1
		mediaType := media.MediaType
		if mediaType == "" {
			mediaType = "image"
		}
		if _, err := s.queries.InsertLLMRequestMedia(ctx, dbgen.InsertLLMRequestMediaParams{
			RequestID: req.ID,
			Position:  int32(position),
			Uri:       media.URI,
			MediaType: mediaType,
		}); err != nil {
			return req, fmt.Errorf("insert media %d: %w", position, err)
		}
	}

	return req, nil
}

// MarkRequestStatus updates status/error and stamps completion time (defaults to now when zero).
func (s *LLMStore) MarkRequestStatus(ctx context.Context, requestID int64, status LLMRequestStatus, errMsg *string, completedAt time.Time) error {
	if s.queries == nil {
		return fmt.Errorf("queries is nil")
	}
	if status == "" {
		return fmt.Errorf("status is required")
	}
	ts := pgtype.Timestamptz{}
	if completedAt.IsZero() {
		ts = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	} else {
		ts = pgtype.Timestamptz{Time: completedAt, Valid: true}
	}
	return s.queries.UpdateLLMRequestStatus(ctx, dbgen.UpdateLLMRequestStatusParams{
		ID:           requestID,
		Status:       string(status),
		ErrorMessage: errMsg,
		CompletedAt:  ts,
	})
}

func (s *LLMStore) SaveBehaviourResponse(ctx context.Context, requestID int64, traits domain.BehaviourTraits, raw json.RawMessage) (dbgen.BehaviourResponse, error) {
	if s.queries == nil {
		return dbgen.BehaviourResponse{}, fmt.Errorf("queries is nil")
	}
	return s.queries.InsertBehaviourResponse(ctx, dbgen.InsertBehaviourResponseParams{
		RequestID:   requestID,
		TraitsJson:  traits,
		RawResponse: raw,
	})
}

func (s *LLMStore) SavePhotoPersonaResponse(ctx context.Context, requestID int64, persona domain.PhotoPersonaBundle, raw json.RawMessage) (dbgen.PhotoPersonaResponse, error) {
	if s.queries == nil {
		return dbgen.PhotoPersonaResponse{}, fmt.Errorf("queries is nil")
	}
	return s.queries.InsertPhotoPersonaResponse(ctx, dbgen.InsertPhotoPersonaResponseParams{
		RequestID:   requestID,
		PersonaJson: persona,
		RawResponse: raw,
	})
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
