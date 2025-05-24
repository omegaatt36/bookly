-- name: CreateBudget :one
INSERT INTO budgets (
    user_id,
    name,
    category,
    amount,
    period_type,
    start_date,
    end_date,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetBudgetByID :one
SELECT * FROM budgets
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetBudgetsByUserID :many
SELECT * FROM budgets
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetActiveBudgetsByUserID :many
SELECT * FROM budgets
WHERE user_id = $1 AND deleted_at IS NULL AND is_active = true
ORDER BY created_at DESC;

-- name: GetBudgetsByUserIDAndCategory :many
SELECT * FROM budgets
WHERE user_id = $1 AND category = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetActiveBudgetByUserIDCategoryAndPeriod :one
SELECT * FROM budgets
WHERE user_id = $1 
AND category = $2 
AND period_type = $3
AND is_active = true
AND deleted_at IS NULL
AND start_date <= $4
AND (end_date IS NULL OR end_date >= $4)
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateBudget :one
UPDATE budgets
SET
    name = COALESCE($2, name),
    category = COALESCE($3, category),
    amount = COALESCE($4, amount),
    period_type = COALESCE($5, period_type),
    start_date = COALESCE($6, start_date),
    end_date = COALESCE($7, end_date),
    is_active = COALESCE($8, is_active),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteBudget :exec
UPDATE budgets
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetLedgersByCategoryAndDateRange :many
SELECT * FROM ledgers
WHERE account_id IN (
    SELECT id FROM accounts WHERE user_id = $1 AND deleted_at IS NULL
)
AND category = $2
AND date >= $3
AND date <= $4
AND deleted_at IS NULL
AND is_voided = false
ORDER BY date DESC;

-- name: GetLedgersSumByCategoryAndDateRange :one
SELECT COALESCE(SUM(
    CASE 
        WHEN type = 'expense' THEN amount
        WHEN type = 'income' THEN -amount
        ELSE 0
    END
), 0)::DECIMAL(20, 2) as total_expense
FROM ledgers
WHERE account_id IN (
    SELECT id FROM accounts WHERE user_id = $1 AND deleted_at IS NULL
)
AND category = $2
AND date >= $3
AND date <= $4
AND deleted_at IS NULL
AND is_voided = false;