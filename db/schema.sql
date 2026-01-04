-- Schema for storing swipe decisions emitted by the decision engine.

CREATE TYPE app_action_kind AS ENUM ('PASS', 'LIKE', 'SUPERSWIPE');

CREATE TABLE decisions (
    id BIGSERIAL PRIMARY KEY,
    profile_key TEXT,
    app TEXT NOT NULL,
    policy_name TEXT NOT NULL,
    action_kind app_action_kind NOT NULL,
    action_message TEXT,
    score INTEGER NOT NULL DEFAULT 0,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX decisions_profile_key_idx ON decisions (profile_key);
CREATE INDEX decisions_created_at_idx ON decisions (created_at);

CREATE TABLE behaviour_traits (
    id BIGSERIAL PRIMARY KEY,
    profile_key TEXT,
    app TEXT NOT NULL,
    traits JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX behaviour_traits_profile_key_idx ON behaviour_traits (profile_key);
CREATE INDEX behaviour_traits_created_at_idx ON behaviour_traits (created_at);
