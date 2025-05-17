-- name: CreateBankAccount :one
INSERT INTO bank_accounts (
    account_id,
    account_number,
    bank_name,
    branch_name,
    swift_code
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetBankAccountByID :one
SELECT * FROM bank_accounts
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetBankAccountByAccountID :one
SELECT * FROM bank_accounts
WHERE account_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateBankAccount :one
UPDATE bank_accounts
SET
    account_number = CASE WHEN sqlc.narg('account_number')::text IS NULL THEN account_number ELSE sqlc.narg('account_number') END,
    bank_name = CASE WHEN sqlc.narg('bank_name')::text IS NULL THEN bank_name ELSE sqlc.narg('bank_name') END,
    branch_name = CASE WHEN sqlc.narg('branch_name')::text IS NULL THEN branch_name ELSE sqlc.narg('branch_name') END,
    swift_code = CASE WHEN sqlc.narg('swift_code')::text IS NULL THEN swift_code ELSE sqlc.narg('swift_code') END,
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: DeleteBankAccount :one
UPDATE bank_accounts
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;