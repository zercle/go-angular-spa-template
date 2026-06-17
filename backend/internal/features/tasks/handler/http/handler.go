// STUB FEATURE — delete internal/features/tasks to start your project.

package httphandler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/dto"
	sharederrors "github.com/zercle/go-angular-spa-template/internal/shared/errors"
)

// Handler exposes the tasks domain service over HTTP.
type Handler struct {
	service domain.Service
}

// New returns an HTTP handler for the tasks feature.
func New(service domain.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts the tasks routes on the provided echo group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/tasks", func(c *echo.Context) error { return h.Create(c) })
	g.GET("/tasks", func(c *echo.Context) error { return h.List(c) })
	g.GET("/tasks/:id", func(c *echo.Context) error { return h.Get(c) })
	g.PUT("/tasks/:id", func(c *echo.Context) error { return h.Update(c) })
	g.DELETE("/tasks/:id", func(c *echo.Context) error { return h.Delete(c) })
}

// Create handles POST /tasks.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Create(c *echo.Context) error {
	var req dto.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	task, err := h.service.Create(c.Request().Context(), req.Title)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusCreated, mapTaskToResponse(task))
}

// Get handles GET /tasks/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Get(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, body := sharederrors.HTTPError(domain.ErrInvalidID)
		return c.JSON(status, body)
	}

	task, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, mapTaskToResponse(task))
}

// List handles GET /tasks.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) List(c *echo.Context) error {
	var req dto.ListTasksRequest
	if err := c.Bind(&req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	tasks, err := h.service.List(c.Request().Context(), req.Limit, req.Offset)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, mapTasksToResponse(tasks))
}

// Update handles PUT /tasks/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Update(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, body := sharederrors.HTTPError(domain.ErrInvalidID)
		return c.JSON(status, body)
	}

	var req dto.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	task, err := h.service.Update(c.Request().Context(), id, req.Title, req.Done)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, mapTaskToResponse(task))
}

// Delete handles DELETE /tasks/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Delete(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, body := sharederrors.HTTPError(domain.ErrInvalidID)
		return c.JSON(status, body)
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapTaskToResponse(task *domain.Task) dto.TaskResponse {
	if task == nil {
		return dto.TaskResponse{}
	}
	return dto.TaskResponse{
		ID:        task.ID.String(),
		Title:     task.Title,
		Done:      task.Done,
		CreatedAt: task.CreatedAt.Format(timeFormat),
		UpdatedAt: task.UpdatedAt.Format(timeFormat),
	}
}

func mapTasksToResponse(tasks []domain.Task) dto.ListTasksResponse {
	resp := dto.ListTasksResponse{Tasks: make([]dto.TaskResponse, len(tasks))}
	for i := range tasks {
		resp.Tasks[i] = mapTaskToResponse(&tasks[i])
	}
	return resp
}
