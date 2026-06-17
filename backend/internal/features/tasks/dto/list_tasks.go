// STUB FEATURE — delete internal/features/tasks to start your project.

package dto

// ListTasksRequest carries pagination parameters for listing tasks.
type ListTasksRequest struct {
	Limit  int32 `json:"limit" query:"limit" validate:"omitempty,min=0,max=100"`
	Offset int32 `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

// ListTasksResponse wraps a page of tasks.
type ListTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}
