package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/domain"
)

// mockRepo is an in-memory domain.TaskRepository for tests.
type mockRepo struct {
	tasks  map[int64]domain.Task
	nextID int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{tasks: map[int64]domain.Task{}, nextID: 1}
}

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
	t.Title, t.Done, t.UpdatedAt = p.Title, p.Done, time.Now()
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

func TestCreate_RejectsBlankTitle(t *testing.T) {
	svc := NewTaskService(newMockRepo())
	for _, title := range []string{"", "   ", "\t"} {
		if _, err := svc.Create(context.Background(), domain.CreateTaskParams{Title: title}); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("title %q: want ErrInvalidInput, got %v", title, err)
		}
	}
}

func TestCreate_TrimsTitle(t *testing.T) {
	svc := NewTaskService(newMockRepo())
	task, err := svc.Create(context.Background(), domain.CreateTaskParams{Title: "  buy milk  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "buy milk" {
		t.Fatalf("want trimmed title %q, got %q", "buy milk", task.Title)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := NewTaskService(newMockRepo())
	_, err := svc.Update(context.Background(), 42, domain.UpdateTaskParams{Title: "x"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
