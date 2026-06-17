-- name: CreateTask :one
INSERT INTO tasks (title)
VALUES ($1)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY created_at DESC, id DESC;

-- name: UpdateTask :one
UPDATE tasks
SET title = $1,
    done = $2,
    updated_at = now()
WHERE id = $3
RETURNING *;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE id = $1;
