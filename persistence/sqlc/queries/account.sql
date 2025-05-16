-- name: CreateAccount :one
INSERT INTO accounts (
    user_id,
    name,
    currency,
    status,
    balance
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateAccount :one
UPDATE accounts
SET
    user_id = CASE WHEN sqlc.narg('user_id')::int IS NULL THEN user_id ELSE sqlc.narg('user_id') END,
    name = CASE WHEN sqlc.narg('name')::text IS NULL THEN name ELSE sqlc.narg('name') END,
    currency = CASE WHEN sqlc.narg('currency')::text IS NULL THEN currency ELSE sqlc.narg('currency') END,
    status = CASE WHEN sqlc.narg('status')::text IS NULL THEN status ELSE sqlc.narg('status') END,
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: DeactivateAccountByID :one
UPDATE accounts
SET
    status = sqlc.arg('status'),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: GetAllAccounts :many
SELECT * FROM accounts
WHERE deleted_at IS NULL
ORDER BY created_at;

-- name: GetAccountsByUserID :many
SELECT * FROM accounts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at;

-- name: IncreaseAccountBalance :one
UPDATE accounts
SET
    balance = balance + sqlc.arg('balance'),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAccount :one
UPDATE accounts
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;
