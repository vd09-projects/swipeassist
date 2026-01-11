-- Store and fetch LLM extraction requests and their outputs.

-- name: InsertLLMRequest :one
INSERT INTO llm_requests (
    kind,
    profile_key,
    app,
    template_path,
    prompt_text,
    vars,
    model,
    status,
    error_message,
    parent_request_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING
    id,
    kind,
    profile_key,
    app,
    template_path,
    prompt_text,
    vars,
    model,
    status,
    error_message,
    parent_request_id,
    created_at,
    completed_at;

-- name: UpdateLLMRequestStatus :exec
UPDATE llm_requests
SET
    status = $2,
    error_message = $3,
    completed_at = $4
WHERE id = $1;

-- name: GetLLMRequest :one
SELECT
    id,
    kind,
    profile_key,
    app,
    template_path,
    prompt_text,
    vars,
    model,
    status,
    error_message,
    parent_request_id,
    created_at,
    completed_at
FROM llm_requests
WHERE id = $1;

-- name: InsertLLMRequestMedia :one
INSERT INTO llm_request_media (
    request_id,
    position,
    uri,
    media_type
) VALUES ($1, $2, $3, $4)
RETURNING
    id,
    request_id,
    position,
    uri,
    media_type,
    created_at;

-- name: ListLLMRequestMedia :many
SELECT
    id,
    request_id,
    position,
    uri,
    media_type,
    created_at
FROM llm_request_media
WHERE request_id = $1
ORDER BY position;

-- name: InsertBehaviourResponse :one
INSERT INTO behaviour_responses (
    request_id,
    traits_json,
    raw_response
) VALUES ($1, $2, $3)
RETURNING
    id,
    request_id,
    traits_json,
    raw_response,
    created_at;

-- name: InsertPhotoPersonaResponse :one
INSERT INTO photo_persona_responses (
    request_id,
    persona_json,
    raw_response
) VALUES ($1, $2, $3)
RETURNING
    id,
    request_id,
    persona_json,
    raw_response,
    created_at;
