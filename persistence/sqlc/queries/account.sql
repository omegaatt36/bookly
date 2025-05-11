-- name: CreateAccount :exec
INSERT INTO accounts (
    user_id,
    name,
    currency,
    status,
    balance
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateAccount :exec
UPDATE accounts
SET
    user_id = CASE WHEN sqlc.narg('user_id')::uuid IS NULL THEN user_id ELSE sqlc.narg('user_id') END,
    name = CASE WHEN sqlc.narg('name')::text IS NULL THEN name ELSE sqlc.narg('name') END,
    currency = CASE WHEN sqlc.narg('currency')::text IS NULL THEN currency ELSE sqlc.narg('currency') END,
    status = CASE WHEN sqlc.narg('status')::text IS NULL THEN status ELSE sqlc.narg('status') END,
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: DeactivateAccountByID :exec
UPDATE accounts
SET
    status = sqlc.arg('status'),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetAllAccounts :many
SELECT * FROM accounts
WHERE deleted_at IS NULL
ORDER BY created_at;

-- name: GetAccountsByUserID :many
SELECT * FROM accounts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at;

-- name: IncreaseAccountBalance :exec
UPDATE accounts
SET
    balance = balance + sqlc.arg('balance'),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: DeleteAccount :exec
UPDATE accounts
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
