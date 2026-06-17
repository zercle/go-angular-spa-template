// STUB FEATURE — delete internal/features/tasks to start your project.

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
)

// defaultCacheTTL bounds how long individual tasks and list pages are cached.
const defaultCacheTTL = 60 * time.Second

const listGenKey = "tasks:list:gen"

// Metrics holds the OpenTelemetry counters the cached repository emits. They
// surface on /metrics via the Prometheus exporter.
type Metrics struct {
	Hits    metric.Int64Counter
	Misses  metric.Int64Counter
	Created metric.Int64Counter
}

// CachedRepository decorates a domain.Repository with a read-through Valkey
// cache, OpenTelemetry spans, and Prometheus counters. It implements
// domain.Repository, so the service layer is unaware of caching.
type CachedRepository struct {
	inner   domain.Repository
	cache   Cache
	tracer  trace.Tracer
	metrics Metrics
	ttl     time.Duration
}

// NewCachedRepository wraps inner with caching + instrumentation.
func NewCachedRepository(inner domain.Repository, cache Cache, tracer trace.Tracer, m Metrics) *CachedRepository {
	return &CachedRepository{inner: inner, cache: cache, tracer: tracer, metrics: m, ttl: defaultCacheTTL}
}

var _ domain.Repository = (*CachedRepository)(nil)

func taskKey(id uuid.UUID) string { return "task:" + id.String() }

// Create persists a task and invalidates cached list pages.
func (r *CachedRepository) Create(ctx context.Context, task *domain.Task) error {
	ctx, span := r.tracer.Start(ctx, "tasks.repository.Create")
	defer span.End()

	if err := r.inner.Create(ctx, task); err != nil {
		span.RecordError(err)
		return err //nolint:wrapcheck // decorator passes the domain error through unchanged
	}
	r.cache.Incr(ctx, listGenKey) // invalidate cached list pages
	r.metrics.Created.Add(ctx, 1)
	return nil
}

// GetByID returns a task, serving from the cache on a hit and populating it on a miss.
func (r *CachedRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	ctx, span := r.tracer.Start(ctx, "tasks.repository.GetByID",
		trace.WithAttributes(attribute.String("task.id", id.String())))
	defer span.End()

	if b, ok := r.cache.Get(ctx, taskKey(id)); ok {
		var cached domain.Task
		if json.Unmarshal(b, &cached) == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			r.metrics.Hits.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "get")))
			return &cached, nil
		}
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))
	r.metrics.Misses.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "get")))

	task, err := r.inner.GetByID(ctx, id)
	if err != nil {
		return nil, err //nolint:wrapcheck // pass through domain error (incl. ErrTaskNotFound); misses are not cached
	}
	if b, err := json.Marshal(task); err == nil {
		r.cache.Set(ctx, taskKey(id), b, r.ttl)
	}
	return task, nil
}

// List returns a page of tasks, serving from the cache on a hit and populating it on a miss.
func (r *CachedRepository) List(ctx context.Context, limit, offset int32) ([]domain.Task, error) {
	ctx, span := r.tracer.Start(ctx, "tasks.repository.List")
	defer span.End()

	key := r.listKey(ctx, limit, offset)
	if b, ok := r.cache.Get(ctx, key); ok {
		var cached []domain.Task
		if json.Unmarshal(b, &cached) == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			r.metrics.Hits.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "list")))
			return cached, nil
		}
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))
	r.metrics.Misses.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "list")))

	tasks, err := r.inner.List(ctx, limit, offset)
	if err != nil {
		return nil, err //nolint:wrapcheck // decorator passes the domain error through unchanged
	}
	if b, err := json.Marshal(tasks); err == nil {
		r.cache.Set(ctx, key, b, r.ttl)
	}
	return tasks, nil
}

// Update persists changes, evicts the task entry, and invalidates list pages.
func (r *CachedRepository) Update(ctx context.Context, task *domain.Task) error {
	ctx, span := r.tracer.Start(ctx, "tasks.repository.Update")
	defer span.End()

	if err := r.inner.Update(ctx, task); err != nil {
		span.RecordError(err)
		return err //nolint:wrapcheck // decorator passes the domain error through unchanged
	}
	r.cache.Delete(ctx, taskKey(task.ID))
	r.cache.Incr(ctx, listGenKey)
	return nil
}

// Delete removes a task, evicts the task entry, and invalidates list pages.
func (r *CachedRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "tasks.repository.Delete")
	defer span.End()

	if err := r.inner.Delete(ctx, id); err != nil {
		span.RecordError(err)
		return err //nolint:wrapcheck // decorator passes the domain error through unchanged
	}
	r.cache.Delete(ctx, taskKey(id))
	r.cache.Incr(ctx, listGenKey)
	return nil
}

// listKey namespaces list-page cache keys by a generation counter so a single
// INCR on write invalidates every cached page without scanning keys.
func (r *CachedRepository) listKey(ctx context.Context, limit, offset int32) string {
	gen := "0"
	if b, ok := r.cache.Get(ctx, listGenKey); ok {
		gen = string(b)
	}
	return fmt.Sprintf("tasks:list:%d:%d:%s", limit, offset, gen)
}
