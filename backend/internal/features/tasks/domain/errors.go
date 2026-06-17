// STUB FEATURE — delete internal/features/tasks to start your project.

package domain

import "errors"

// Domain sentinel errors for the tasks feature.
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTitle = errors.New("task title is invalid")
	ErrInvalidID    = errors.New("task id is invalid")
)
