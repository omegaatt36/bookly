-- name: CreateUser :one
INSERT INTO users (
    name,
    nickname
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetAllUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY id;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
    name = CASE WHEN sqlc.narg('name')::text IS NULL THEN name ELSE sqlc.narg('name') END,
    nickname = CASE WHEN sqlc.narg('nickname')::text IS NULL THEN nickname ELSE sqlc.narg('nickname') END,
    disabled = CASE WHEN sqlc.narg('disabled')::boolean IS NULL THEN disabled ELSE sqlc.narg('disabled') END,
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: DeactivateUserByID :one
UPDATE users
SET
    disabled = true,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteUser :one
UPDATE users
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: AddIdentity :one
INSERT INTO identities (
    user_id,
    provider,
    identifier,
    credential,
    last_used_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
RETURNING *;

-- name: GetUserByIdentity :one
SELECT
    u.id AS user_id,
    u.created_at AS user_created_at,
    u.updated_at AS user_updated_at,
    u.disabled AS user_disabled,
    u.name AS user_name,
    u.nickname AS user_nickname,
    i.id AS identity_id,
    i.user_id AS identity_user_id,
    i.provider AS identity_provider,
    i.identifier AS identity_identifier,
    i.credential AS identity_credential,
    i.last_used_at AS identity_last_used_at
FROM users u
JOIN identities i ON u.id = i.user_id
WHERE i.provider = $1 AND i.identifier = $2 AND u.deleted_at IS NULL
LIMIT 1;
