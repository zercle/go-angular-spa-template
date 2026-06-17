//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/repository"
	sqlcdb "github.com/zercle/go-angular-spa-template/internal/infrastructure/db/sqlc"
)

// mockDBTX implements sqlc.DBTX for in-memory repository tests.
type mockDBTX struct {
	items     []sqlcdb.Task
	listItems []sqlcdb.Task
	listErr   error
	err       error
}

func (m *mockDBTX) Exec(_ context.Context, _ string, args ...interface{}) (pgconn.CommandTag, error) {
	if m.err != nil {
		return pgconn.CommandTag{}, m.err
	}
	// CreateTask: id, title, done, created_at, updated_at
	if len(args) >= 5 {
		m.items = append(m.items, sqlcdb.Task{
			ID:        args[0].(uuid.UUID),
			Title:     args[1].(string),
			Done:      args[2].(bool),
			CreatedAt: args[3].(time.Time),
			UpdatedAt: args[4].(time.Time),
		})
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (m *mockDBTX) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return &mockRows{items: m.listItems}, nil
}

func (m *mockDBTX) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return &mockRow{item: m.items, err: m.err}
}

// mockRow implements pgx.Row for the fake DBTX.
type mockRow struct {
	item []sqlcdb.Task
	err  error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(r.item) == 0 {
		return pgx.ErrNoRows
	}
	i := r.item[0]
	*dest[0].(*uuid.UUID) = i.ID
	*dest[1].(*string) = i.Title
	*dest[2].(*bool) = i.Done
	*dest[3].(*time.Time) = i.CreatedAt
	*dest[4].(*time.Time) = i.UpdatedAt
	return nil
}

// mockRows implements pgx.Rows for the fake DBTX.
type mockRows struct {
	items []sqlcdb.Task
	idx   int
}

func (r *mockRows) Next() bool {
	if r.idx >= len(r.items) {
		return false
	}
	r.idx++
	return true
}

func (r *mockRows) Scan(dest ...interface{}) error {
	i := r.items[r.idx-1]
	*dest[0].(*uuid.UUID) = i.ID
	*dest[1].(*string) = i.Title
	*dest[2].(*bool) = i.Done
	*dest[3].(*time.Time) = i.CreatedAt
	*dest[4].(*time.Time) = i.UpdatedAt
	return nil
}

func (r *mockRows) Close()                                       {}
func (r *mockRows) Err() error                                   { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

func TestRepository_Create(t *testing.T) {
	dbtx := &mockDBTX{}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	task := &domain.Task{
		ID:        uuid.New(),
		Title:     "repo-task",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := repo.Create(context.Background(), task)
	require.NoError(t, err)
	assert.Len(t, dbtx.items, 1)
}

func TestRepository_Create_Error(t *testing.T) {
	dbtx := &mockDBTX{err: errors.New("exec failed")}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	task := &domain.Task{ID: uuid.New(), Title: "x", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	err := repo.Create(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec failed")
}

func TestRepository_GetByID(t *testing.T) {
	id := uuid.New()
	dbtx := &mockDBTX{items: []sqlcdb.Task{{ID: id, Title: "found", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	dbtx := &mockDBTX{}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	got, err := repo.GetByID(context.Background(), uuid.New())
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, domain.ErrTaskNotFound))
}

func TestRepository_List(t *testing.T) {
	id := uuid.New()
	dbtx := &mockDBTX{listItems: []sqlcdb.Task{{ID: id, Title: "listed", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	tasks, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, id, tasks[0].ID)
}

func TestRepository_List_Error(t *testing.T) {
	dbtx := &mockDBTX{listErr: errors.New("query failed")}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	tasks, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, tasks)
}

func TestRepository_Update(t *testing.T) {
	id := uuid.New()
	dbtx := &mockDBTX{}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	err := repo.Update(context.Background(), &domain.Task{ID: id, Title: "updated", Done: true, UpdatedAt: time.Now().UTC()})
	require.NoError(t, err)
}

func TestRepository_Delete(t *testing.T) {
	dbtx := &mockDBTX{}
	repo := repository.NewRepository(sqlcdb.New(dbtx))

	err := repo.Delete(context.Background(), uuid.New())
	require.NoError(t, err)
}
