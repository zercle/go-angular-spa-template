//go:build integration

// STUB FEATURE — delete internal/features/tasks to start your project.

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zercle/go-angular-spa-template/internal/config"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/repository"
	"github.com/zercle/go-angular-spa-template/internal/infrastructure/db"
	"github.com/zercle/go-angular-spa-template/internal/infrastructure/db/migrations"
	sqlcdb "github.com/zercle/go-angular-spa-template/internal/infrastructure/db/sqlc"
)

type RepositoryIntegrationSuite struct {
	suite.Suite
	pool   *pgxpool.Pool
	repo   *repository.Repository
	testDB string
}

func (s *RepositoryIntegrationSuite) SetupSuite() {
	t := s.T()
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err)

	// Connect to the configured database for integration tests.
	s.pool, err = db.NewPool(context.Background(), cfg)
	require.NoError(t, err)

	s.runMigrations(cfg)

	queries := sqlcdb.New(s.pool)
	s.repo = repository.NewRepository(queries)
}

func (s *RepositoryIntegrationSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *RepositoryIntegrationSuite) SetupTest() {
	t := s.T()
	t.Helper()

	_, err := s.pool.Exec(context.Background(), "TRUNCATE TABLE tasks RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func (s *RepositoryIntegrationSuite) TestCreateAndGetByID() {
	t := s.T()
	ctx := context.Background()

	item := &domain.Task{
		ID:        uuid.New(),
		Title:     "integration-item",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	err := s.repo.Create(ctx, item)
	require.NoError(t, err)

	got, err := s.repo.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, item.ID, got.ID)
	require.Equal(t, item.Title, got.Title)
}

func (s *RepositoryIntegrationSuite) TestGetByID_NotFound() {
	t := s.T()
	ctx := context.Background()

	got, err := s.repo.GetByID(ctx, uuid.New())
	require.Nil(t, got)
	require.True(t, errors.Is(err, domain.ErrTaskNotFound))
}

func (s *RepositoryIntegrationSuite) TestList() {
	t := s.T()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		item := &domain.Task{
			ID:        uuid.New(),
			Title:     "list-item",
			CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
		}
		require.NoError(t, s.repo.Create(ctx, item))
	}

	items, err := s.repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 3)
}

func (s *RepositoryIntegrationSuite) runMigrations(cfg *config.Config) {
	t := s.T()
	t.Helper()

	src, err := iofs.New(migrations.FS, ".")
	require.NoError(t, err)

	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.DBConnString())
	require.NoError(t, err)
	s.T().Cleanup(func() {
		_, _ = m.Close()
	})

	err = m.Up()
	if err == nil {
		return
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return
	}
	require.NoError(t, err)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryIntegrationSuite))
}
