// STUB FEATURE — delete internal/features/tasks to start your project.

package domain

import (
	"time"

	"github.com/google/uuid"
)

// Task is the example entity for the tasks vertical slice.
type Task struct {
	ID        uuid.UUID
	Title     string
	Done      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Rename updates the task title and refreshes the updated-at timestamp.
func (t *Task) Rename(title string) {
	t.Title = title
	t.UpdatedAt = time.Now().UTC()
}
