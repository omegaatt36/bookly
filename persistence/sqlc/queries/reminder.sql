-- name: CreateReminder :one
INSERT INTO reminders (
    recurring_transaction_id, reminder_date
) VALUES (
    $1, $2
) RETURNING *;

-- name: DeleteReminder :exec
UPDATE reminders
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetRemindersByRecurringTransactionID :many
SELECT * FROM reminders
WHERE recurring_transaction_id = $1 AND deleted_at IS NULL
ORDER BY reminder_date ASC;

-- name: GetActiveRemindersByUserID :many
SELECT r.*
FROM reminders r
JOIN recurring_transactions rt ON r.recurring_transaction_id = rt.id
WHERE rt.user_id = $1 AND r.is_read = FALSE AND r.reminder_date <= $2 AND r.deleted_at IS NULL AND rt.deleted_at IS NULL
ORDER BY r.reminder_date ASC;

-- name: GetReminderByID :one
SELECT * FROM reminders
WHERE id = $1 AND deleted_at IS NULL;

-- name: MarkReminderAsRead :one
UPDATE reminders
SET
    is_read = TRUE,
    read_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: GetUpcomingReminders :many
SELECT r.*, rt.name AS transaction_name, rt.amount, rt.type
FROM reminders r
JOIN recurring_transactions rt ON r.recurring_transaction_id = rt.id
WHERE rt.user_id = $1
  AND r.is_read = FALSE
  AND r.reminder_date BETWEEN $2 AND $3
  AND r.deleted_at IS NULL
  AND rt.deleted_at IS NULL
ORDER BY r.reminder_date ASC;
