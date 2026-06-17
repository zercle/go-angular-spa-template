package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/domain"
	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/usecase"
)

type mockRepo struct {
	tasks  map[int64]domain.Task
	nextID int64
}

func newMockRepo() *mockRepo { return &mockRepo{tasks: map[int64]domain.Task{}, nextID: 1} }

func (m *mockRepo) Create(_ context.Context, p domain.CreateTaskParams) (domain.Task, error) {
	t := domain.Task{ID: m.nextID, Title: p.Title, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.tasks[m.nextID] = t
	m.nextID++
	return t, nil
}
func (m *mockRepo) Get(_ context.Context, id int64) (domain.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return domain.Task{}, domain.ErrNotFound
	}
	return t, nil
}
func (m *mockRepo) List(_ context.Context) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		out = append(out, t)
	}
	return out, nil
}
func (m *mockRepo) Update(_ context.Context, id int64, p domain.UpdateTaskParams) (domain.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return domain.Task{}, domain.ErrNotFound
	}
	t.Title, t.Done = p.Title, p.Done
	m.tasks[id] = t
	return t, nil
}
func (m *mockRepo) Delete(_ context.Context, id int64) error {
	if _, ok := m.tasks[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func newTestApp() *fiber.App {
	svc := usecase.NewTaskService(newMockRepo())
	spa := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<html></html>")}}
	return NewRouter(svc, spa)
}

func TestCreateTask_Created(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", resp.StatusCode)
	}
}

func TestCreateTask_BlankTitle_BadRequest(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/999", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %d", resp.StatusCode)
	}
}

func TestSPAFallback_ServesIndex(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/some/client/route", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
}
