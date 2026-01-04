-- Store and fetch BehaviourTraits snapshots.

-- name: InsertBehaviourTraits :one
INSERT INTO behaviour_traits (
    profile_key,
    app,
    traits
) VALUES ($1, $2, $3)
RETURNING id, profile_key, app, traits, created_at;

-- name: GetBehaviourTraits :one
SELECT
    id,
    profile_key,
    app,
    traits,
    created_at
FROM behaviour_traits
WHERE id = $1;

-- name: ListBehaviourTraitsByProfile :many
SELECT
    id,
    profile_key,
    app,
    traits,
    created_at
FROM behaviour_traits
WHERE profile_key = $1
ORDER BY created_at DESC;

-- name: ListRecentBehaviourTraits :many
SELECT
    id,
    profile_key,
    app,
    traits,
    created_at
FROM behaviour_traits
ORDER BY created_at DESC
LIMIT $1;
