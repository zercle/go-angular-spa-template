// STUB FEATURE — delete internal/features/tasks to start your project.

package dto

// CreateTaskRequest is the payload for creating a new task.
type CreateTaskRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

// TaskResponse is the JSON representation of a task.
type TaskResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
