-- name: CreateTask :exec
INSERT INTO tasks (id, title, done, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: UpdateTask :execrows
UPDATE tasks
SET title = $2, done = $3, updated_at = $4
WHERE id = $1;

-- name: DeleteTask :execrows
DELETE FROM tasks WHERE id = $1;
