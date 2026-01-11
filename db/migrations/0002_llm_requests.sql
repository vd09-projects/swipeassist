-- Store LLM extraction requests (with media) and their outputs.

CREATE TYPE extraction_kind AS ENUM ('behaviour', 'photo_persona');

CREATE TABLE llm_requests (
    id BIGSERIAL PRIMARY KEY,
    kind extraction_kind NOT NULL,
    profile_key TEXT,
    app TEXT NOT NULL,
    template_path TEXT NOT NULL,
    prompt_text TEXT NOT NULL,
    vars JSONB,
    model TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    error_message TEXT,
    parent_request_id BIGINT REFERENCES llm_requests (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX llm_requests_profile_key_idx ON llm_requests (profile_key, created_at);
CREATE INDEX llm_requests_kind_idx ON llm_requests (kind, created_at);
CREATE INDEX llm_requests_status_idx ON llm_requests (status);

CREATE TABLE llm_request_media (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT NOT NULL REFERENCES llm_requests (id) ON DELETE CASCADE,
    position INT NOT NULL,
    uri TEXT NOT NULL,
    media_type TEXT NOT NULL DEFAULT 'image',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (request_id, position)
);

CREATE TABLE behaviour_responses (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT UNIQUE NOT NULL REFERENCES llm_requests (id) ON DELETE CASCADE,
    traits_json JSONB NOT NULL,
    raw_response JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE photo_persona_responses (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT UNIQUE NOT NULL REFERENCES llm_requests (id) ON DELETE CASCADE,
    persona_json JSONB NOT NULL,
    raw_response JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
