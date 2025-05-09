-- name: CreateLedger :exec
INSERT INTO ledgers (
    account_id,
    date,
    type,
    amount,
    note,
    is_adjustment,
    adjusted_from
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetLedgerByID :one
SELECT 
    l.*,
    a.currency
FROM ledgers l
JOIN accounts a ON l.account_id = a.id
WHERE l.id = $1
LIMIT 1;

-- name: GetLedgersByAccountID :many
SELECT 
    l.*,
    a.currency
FROM ledgers l
JOIN accounts a ON l.account_id = a.id
WHERE l.account_id = $1
ORDER BY l.date DESC, l.updated_at DESC;

-- name: UpdateLedger :exec
UPDATE ledgers 
SET 
    date = CASE WHEN sqlc.narg('date')::timestamptz IS NULL THEN date ELSE sqlc.narg('date') END,
    type = CASE WHEN sqlc.narg('type')::text IS NULL THEN type ELSE sqlc.narg('type') END,
    amount = CASE WHEN sqlc.narg('amount')::decimal IS NULL THEN amount ELSE sqlc.narg('amount') END,
    note = CASE WHEN sqlc.narg('note')::text IS NULL THEN note ELSE sqlc.narg('note') END,
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: VoidLedger :exec
UPDATE ledgers 
SET 
    is_voided = true,
    voided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: GetLedgerAmount :one
SELECT amount FROM ledgers
WHERE id = sqlc.arg('id')
LIMIT 1;