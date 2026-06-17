// STUB FEATURE — delete internal/features/tasks to start your project.

package dto

// UpdateTaskRequest is the payload for updating a task.
type UpdateTaskRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
	Done  bool   `json:"done"`
}
