// Package domain holds the core business entities and the ports (interfaces)
// that the outer layers implement. It imports no framework code.
package domain

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a task does not exist.
var ErrNotFound = errors.New("task not found")

// ErrInvalidInput is returned when input fails domain validation.
var ErrInvalidInput = errors.New("invalid input")

// Task is the core entity of the demo vertical slice.
type Task struct {
	ID        int64
	Title     string
	Done      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateTaskParams are the inputs required to create a task.
type CreateTaskParams struct {
	Title string
}

// UpdateTaskParams are the inputs required to update a task.
type UpdateTaskParams struct {
	Title string
	Done  bool
}

// TaskRepository is the persistence port implemented by the infrastructure layer.
type TaskRepository interface {
	Create(ctx context.Context, p CreateTaskParams) (Task, error)
	Get(ctx context.Context, id int64) (Task, error)
	List(ctx context.Context) ([]Task, error)
	Update(ctx context.Context, id int64, p UpdateTaskParams) (Task, error)
	Delete(ctx context.Context, id int64) error
}
