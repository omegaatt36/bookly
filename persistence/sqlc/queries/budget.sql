-- name: CreateBudget :one
INSERT INTO budgets (
    user_id,
    name,
    period,
    start_date,
    end_date,
    amount,
    category_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetBudgetByID :one
SELECT * FROM budgets
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListBudgetsByUserID :many
SELECT * FROM budgets
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY start_date DESC;

-- name: UpdateBudget :one
UPDATE budgets
SET
    name = $2,
    period = $3,
    start_date = $4,
    end_date = $5,
    amount = $6,
    category_id = $7,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteBudget :exec
UPDATE budgets
SET deleted_at = NOW()
WHERE id = $1;
