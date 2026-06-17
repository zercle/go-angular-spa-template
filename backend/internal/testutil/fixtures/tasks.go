// Package fixtures provides sample domain objects for tests.
package fixtures

import (
	"time"

	"github.com/google/uuid"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
)

// NewTask returns a sample Task with the given title. It uses a deterministic
// UUID and fixed timestamps so tests can assert against known values.
func NewTask(title string) domain.Task {
	return domain.Task{
		ID:        uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
		Title:     title,
		Done:      false,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}
