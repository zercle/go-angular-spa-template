package rest

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/domain"
	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/usecase"
)

// TaskHandler exposes the tasks resource over HTTP.
type TaskHandler struct {
	svc      *usecase.TaskService
	validate *validator.Validate
}

// NewTaskHandler builds a handler for the given service.
func NewTaskHandler(svc *usecase.TaskService) *TaskHandler {
	return &TaskHandler{
		svc:      svc,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// Register mounts the task routes onto the given router (the /api/v1 group).
func (h *TaskHandler) Register(r fiber.Router) {
	r.Get("/tasks", h.list)
	r.Post("/tasks", h.create)
	r.Get("/tasks/:id", h.get)
	r.Put("/tasks/:id", h.update)
	r.Delete("/tasks/:id", h.remove)
}

type createTaskRequest struct {
	Title string `json:"title" validate:"required,max=200"`
}

type updateTaskRequest struct {
	Title string `json:"title" validate:"required,max=200"`
	Done  bool   `json:"done"`
}

type taskResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toResponse(t domain.Task) taskResponse {
	return taskResponse{
		ID:        t.ID,
		Title:     t.Title,
		Done:      t.Done,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (h *TaskHandler) list(c fiber.Ctx) error {
	tasks, err := h.svc.List(c.Context())
	if err != nil {
		return h.mapError(c, err)
	}
	out := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toResponse(t))
	}
	return ok(c, fiber.StatusOK, out)
}

func (h *TaskHandler) get(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid id")
	}
	t, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return h.mapError(c, err)
	}
	return ok(c, fiber.StatusOK, toResponse(t))
}

func (h *TaskHandler) create(c fiber.Ctx) error {
	var req createTaskRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	t, err := h.svc.Create(c.Context(), domain.CreateTaskParams{Title: req.Title})
	if err != nil {
		return h.mapError(c, err)
	}
	return ok(c, fiber.StatusCreated, toResponse(t))
}

func (h *TaskHandler) update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid id")
	}
	var req updateTaskRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	t, err := h.svc.Update(c.Context(), id, domain.UpdateTaskParams{Title: req.Title, Done: req.Done})
	if err != nil {
		return h.mapError(c, err)
	}
	return ok(c, fiber.StatusOK, toResponse(t))
}

func (h *TaskHandler) remove(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return h.mapError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseID(c fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("id"), 10, 64)
}

func (h *TaskHandler) mapError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return fail(c, fiber.StatusNotFound, "task not found")
	case errors.Is(err, domain.ErrInvalidInput):
		return fail(c, fiber.StatusBadRequest, err.Error())
	default:
		return fail(c, fiber.StatusInternalServerError, "internal server error")
	}
}
