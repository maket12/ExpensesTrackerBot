-- name: CreateExpense :one
INSERT INTO expenses (id, user_id, amount, category, spent_at, comment, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at;

-- name: GetExpenseByID :one
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE id = $1;

-- name: UpdateExpense :one
UPDATE expenses
SET amount = $2,
    category = $3,
    spent_at = $4,
    comment = $5,
    updated_at = $6
WHERE id = $1
RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at;

-- name: DeleteExpense :execrows
DELETE FROM expenses
WHERE id = $1;

-- name: GetLatestExpensesByUser :many
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE user_id = $1
ORDER BY spent_at DESC, created_at DESC
LIMIT $2;

-- name: GetDistinctUserIDs :many
SELECT DISTINCT user_id
FROM expenses
ORDER BY user_id ASC;

-- name: ListExpensesFiltered :many
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE ($1::bigint IS NULL OR user_id = $1)
  AND ($2::varchar IS NULL OR category = $2)
  AND ($3::timestamptz IS NULL OR spent_at >= $3)
  AND ($4::timestamptz IS NULL OR spent_at <= $4)
ORDER BY spent_at DESC, created_at DESC
LIMIT $5 OFFSET $6;
