// STUB FEATURE — delete internal/features/tasks to start your project.

package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the outbound port for Task persistence.
//
//go:generate go tool mockgen -source=repository.go -destination=../repository/mock/repository_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	List(ctx context.Context, limit, offset int32) ([]Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}
