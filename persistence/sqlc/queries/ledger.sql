-- name: CreateLedger :one
INSERT INTO ledgers (
    account_id,
    date,
    type,
    amount,
    note,
    is_adjustment,
    adjusted_from,
    category_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetLedgerByID :one
SELECT
    l.*,
    a.currency
FROM ledgers l
JOIN accounts a ON l.account_id = a.id
WHERE l.id = $1 AND l.deleted_at IS NULL AND a.deleted_at IS NULL
LIMIT 1;

-- name: DeleteLedger :one
UPDATE ledgers
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: GetLedgersByAccountID :many
SELECT
    l.*,
    a.currency
FROM ledgers l
JOIN accounts a ON l.account_id = a.id
WHERE l.account_id = $1 AND l.deleted_at IS NULL AND a.deleted_at IS NULL
ORDER BY l.date DESC, l.updated_at DESC;

-- name: UpdateLedger :one
UPDATE ledgers
SET
    date = CASE WHEN sqlc.narg('date')::timestamptz IS NULL THEN date ELSE sqlc.narg('date') END,
    type = CASE WHEN sqlc.narg('type')::text IS NULL THEN type ELSE sqlc.narg('type') END,
    amount = CASE WHEN sqlc.narg('amount')::decimal IS NULL THEN amount ELSE sqlc.narg('amount') END,
    note = CASE WHEN sqlc.narg('note')::text IS NULL THEN note ELSE sqlc.narg('note') END,
    category_id = CASE WHEN sqlc.narg('category_id')::int IS NULL THEN category_id ELSE sqlc.narg('category_id') END,
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: VoidLedger :one
UPDATE ledgers
SET
    is_voided = true,
    voided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: GetLedgerAmount :one
SELECT amount FROM ledgers
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
LIMIT 1;

-- name: ListLedgersByUserIDDateRangeCategory :many
SELECT
    l.*,
    a.currency -- Include currency from account
FROM ledgers l
JOIN accounts a ON l.account_id = a.id
WHERE a.user_id = $1
  AND l.date >= $2
  AND l.date <= $3
  AND l.category_id = $4
  AND l.deleted_at IS NULL
  AND a.deleted_at IS NULL
ORDER BY l.date DESC;
