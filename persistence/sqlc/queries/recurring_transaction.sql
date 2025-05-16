-- name: CreateRecurringTransaction :one
INSERT INTO recurring_transactions (
    user_id, account_id, name, type, amount, note,
    start_date, end_date, recur_type, status, frequency,
    day_of_week, day_of_month, month_of_year, next_due
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetRecurringTransactionByID :one
SELECT * FROM recurring_transactions
WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetRecurringTransactionsByUserID :many
SELECT * FROM recurring_transactions
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY next_due ASC;

-- name: GetActiveRecurringTransactionsDue :many
SELECT * FROM recurring_transactions
WHERE status = 'active' AND next_due <= $1 AND deleted_at IS NULL
ORDER BY next_due ASC;

-- name: UpdateRecurringTransaction :one
UPDATE recurring_transactions
SET
    updated_at = NOW(),
    name = CASE WHEN sqlc.narg('name')::text IS NULL THEN name ELSE sqlc.narg('name') END,
    type = CASE WHEN sqlc.narg('type')::text IS NULL THEN type ELSE sqlc.narg('type') END,
    amount = CASE WHEN sqlc.narg('amount')::decimal IS NULL THEN amount ELSE sqlc.narg('amount') END,
    note = CASE WHEN sqlc.narg('note')::text IS NULL THEN note ELSE sqlc.narg('note') END,
    end_date = CASE WHEN sqlc.narg('end_date')::timestamptz IS NULL THEN end_date ELSE sqlc.narg('end_date') END,
    recur_type = CASE WHEN sqlc.narg('recur_type')::text IS NULL THEN recur_type ELSE sqlc.narg('recur_type') END,
    status = CASE WHEN sqlc.narg('status')::text IS NULL THEN status ELSE sqlc.narg('status') END,
    frequency = CASE WHEN sqlc.narg('frequency')::int IS NULL THEN frequency ELSE sqlc.narg('frequency') END,
    day_of_week = CASE WHEN sqlc.narg('day_of_week')::int IS NULL THEN day_of_week ELSE sqlc.narg('day_of_week') END,
    day_of_month = CASE WHEN sqlc.narg('day_of_month')::int IS NULL THEN day_of_month ELSE sqlc.narg('day_of_month') END,
    month_of_year = CASE WHEN sqlc.narg('month_of_year')::int IS NULL THEN month_of_year ELSE sqlc.narg('month_of_year') END
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateRecurringTransactionExecution :one
UPDATE recurring_transactions
SET
    updated_at = NOW(),
    last_executed = sqlc.arg('last_executed'),
    next_due = sqlc.arg('next_due')
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: DeleteRecurringTransaction :one
UPDATE recurring_transactions
SET
    updated_at = NOW(),
    status = 'cancelled',
    deleted_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;
