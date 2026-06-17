// STUB FEATURE — delete internal/features/tasks to start your project.

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
)

const (
	defaultPageSize int32 = 20
	maxPageSize     int32 = 100
	maxTitleLength        = 255
)

// Service implements the domain.Service inbound use-case port.
type Service struct {
	repo domain.Repository
}

// NewService returns a Service backed by the provided repository.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// Create validates the title and persists a new task.
func (s *Service) Create(ctx context.Context, title string) (*domain.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > maxTitleLength {
		return nil, domain.ErrInvalidTitle
	}

	now := time.Now().UTC()
	task := &domain.Task{
		ID:        uuid.New(),
		Title:     title,
		Done:      false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

// Get retrieves a task by ID, passing through domain.ErrTaskNotFound.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}

// List returns a paginated list of tasks. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0.
func (s *Service) List(ctx context.Context, limit, offset int32) ([]domain.Task, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	tasks, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	return tasks, nil
}

// Update validates the title and updates an existing task's title and done flag.
func (s *Service) Update(ctx context.Context, id uuid.UUID, title string, done bool) (*domain.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > maxTitleLength {
		return nil, domain.ErrInvalidTitle
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	existing.Title = title
	existing.Done = done
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("update task: %w", err)
	}

	return existing, nil
}

// Delete removes a task by ID, passing through domain.ErrTaskNotFound.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return domain.ErrTaskNotFound
		}
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}
