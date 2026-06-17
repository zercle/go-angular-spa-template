//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package repository_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/repository"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/repository/mock"
)

// fakeCache is an in-memory Cache for testing the decorator.
type fakeCache struct {
	data map[string][]byte
}

func newFakeCache() *fakeCache { return &fakeCache{data: map[string][]byte{}} }

func (f *fakeCache) Get(_ context.Context, key string) ([]byte, bool) {
	b, ok := f.data[key]
	return b, ok
}

func (f *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) {
	f.data[key] = value
}
func (f *fakeCache) Delete(_ context.Context, key string) { delete(f.data, key) }
func (f *fakeCache) Incr(_ context.Context, key string) {
	n := int64(0)
	if b, ok := f.data[key]; ok {
		n, _ = strconv.ParseInt(string(b), 10, 64)
	}
	f.data[key] = []byte(strconv.FormatInt(n+1, 10))
}

func newCached(t *testing.T, inner domain.Repository, cache repository.Cache) *repository.CachedRepository {
	t.Helper()
	meter := mnoop.NewMeterProvider().Meter("test")
	hits, _ := meter.Int64Counter("h")
	misses, _ := meter.Int64Counter("m")
	created, _ := meter.Int64Counter("c")
	return repository.NewCachedRepository(inner, cache, tnoop.NewTracerProvider().Tracer("test"),
		repository.Metrics{Hits: hits, Misses: misses, Created: created})
}

func TestCachedRepository_GetByID_MissThenHit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	inner := mock.NewMockRepository(ctrl)
	id := uuid.New()
	task := &domain.Task{ID: id, Title: "cached"}

	// inner is consulted exactly once; the second call must be served from cache.
	inner.EXPECT().GetByID(gomock.Any(), id).Return(task, nil).Times(1)

	repo := newCached(t, inner, newFakeCache())

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "cached", got.Title)

	got2, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "cached", got2.Title)
}

func TestCachedRepository_GetByID_NotFoundNotCached(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	inner := mock.NewMockRepository(ctrl)
	id := uuid.New()

	// Not-found must hit inner every time (misses are never cached).
	inner.EXPECT().GetByID(gomock.Any(), id).Return(nil, domain.ErrTaskNotFound).Times(2)

	repo := newCached(t, inner, newFakeCache())
	for range 2 {
		_, err := repo.GetByID(context.Background(), id)
		require.ErrorIs(t, err, domain.ErrTaskNotFound)
	}
}

func TestCachedRepository_List_MissThenHit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	inner := mock.NewMockRepository(ctrl)
	want := []domain.Task{{ID: uuid.New(), Title: "a"}}

	inner.EXPECT().List(gomock.Any(), int32(10), int32(0)).Return(want, nil).Times(1)

	repo := newCached(t, inner, newFakeCache())

	got, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, got, 1)

	got2, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, got2, 1)
}

func TestCachedRepository_Create_InvalidatesLists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	inner := mock.NewMockRepository(ctrl)
	inner.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	cache := newFakeCache()
	repo := newCached(t, inner, cache)

	require.NoError(t, repo.Create(context.Background(), &domain.Task{ID: uuid.New(), Title: "x"}))
	assert.Equal(t, []byte("1"), cache.data["tasks:list:gen"], "create should bump the list generation")
}

func TestCachedRepository_Update_EvictsAndInvalidates(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	inner := mock.NewMockRepository(ctrl)
	id := uuid.New()
	inner.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	cache := newFakeCache()
	cache.data["task:"+id.String()] = []byte(`{"cached":true}`)
	repo := newCached(t, inner, cache)

	require.NoError(t, repo.Update(context.Background(), &domain.Task{ID: id, Title: "y"}))
	_, present := cache.data["task:"+id.String()]
	assert.False(t, present, "update should evict the cached task")
	assert.Equal(t, []byte("1"), cache.data["tasks:list:gen"])
}

func TestCachedRepository_Delete_EvictsAndInvalidates(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	inner := mock.NewMockRepository(ctrl)
	id := uuid.New()
	inner.EXPECT().Delete(gomock.Any(), id).Return(nil)

	cache := newFakeCache()
	cache.data["task:"+id.String()] = []byte(`{}`)
	repo := newCached(t, inner, cache)

	require.NoError(t, repo.Delete(context.Background(), id))
	_, present := cache.data["task:"+id.String()]
	assert.False(t, present)
	assert.Equal(t, []byte("1"), cache.data["tasks:list:gen"])
}
