-- name: CreateCategory :one
INSERT INTO categories (
    user_id,
    name
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCategoriesByUserID :many
SELECT * FROM categories
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET
    name = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE id = $1;
