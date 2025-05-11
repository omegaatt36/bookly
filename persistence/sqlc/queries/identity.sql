-- name: GetIdentitiesByUserID :many
SELECT * FROM identities
WHERE user_id = $1;

-- name: GetIdentityByProviderAndIdentifier :one
SELECT * FROM identities
WHERE provider = $1 AND identifier = $2
LIMIT 1;

-- name: UpdateIdentityCredential :exec
UPDATE identities
SET
    credential = $3,
    last_used_at = NOW()
WHERE provider = $1 AND identifier = $2;

-- name: UpdateIdentityLastUsed :exec
UPDATE identities
SET
    last_used_at = NOW()
WHERE provider = $1 AND identifier = $2;

-- name: DeleteIdentity :exec
DELETE FROM identities
WHERE provider = $1 AND identifier = $2;
