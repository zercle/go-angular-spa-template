// STUB FEATURE — delete internal/features/tasks to start your project.

package domain

import (
	"context"

	"github.com/google/uuid"
)

// Service is the inbound use-case port for Tasks.
//
//go:generate go tool mockgen -source=service.go -destination=../service/mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, title string) (*Task, error)
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	List(ctx context.Context, limit, offset int32) ([]Task, error)
	Update(ctx context.Context, id uuid.UUID, title string, done bool) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
