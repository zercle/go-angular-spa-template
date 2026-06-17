//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/repository/mock"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/service"
)

func TestService_Create_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchItemName("stub")).Return(nil)

	svc := service.NewService(repo)
	item, err := svc.Create(ctx, "stub")

	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, "stub", item.Title)
	require.NotEqual(t, uuid.Nil, item.ID)
	require.False(t, item.CreatedAt.IsZero())
	require.False(t, item.UpdatedAt.IsZero())
}

func TestService_Create_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := service.NewService(repo)

	item, err := svc.Create(ctx, "")

	require.ErrorIs(t, err, domain.ErrInvalidTitle)
	require.Nil(t, item)
}

func TestService_Create_WhitespaceName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := service.NewService(repo)
	item, err := svc.Create(ctx, "   ")
	require.ErrorIs(t, err, domain.ErrInvalidTitle)
	require.Nil(t, item)
}

func TestService_Get_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	expected := &domain.Task{ID: id, Title: "found"}
	repo.EXPECT().GetByID(ctx, id).Return(expected, nil)

	svc := service.NewService(repo)
	item, err := svc.Get(ctx, id)

	require.NoError(t, err)
	require.Equal(t, expected, item)
}

func TestService_Get_MapsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrTaskNotFound)

	svc := service.NewService(repo)
	item, err := svc.Get(ctx, id)

	require.ErrorIs(t, err, domain.ErrTaskNotFound)
	require.Nil(t, item)
}

func TestService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	expected := []domain.Task{{ID: uuid.New(), Title: "one"}}
	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(expected, nil)

	svc := service.NewService(repo)
	items, err := svc.List(ctx, 10, 5)

	require.NoError(t, err)
	require.Equal(t, expected, items)
}

func TestService_List_AppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	expected := []domain.Task{{ID: uuid.New(), Title: "default"}}
	repo.EXPECT().List(ctx, int32(20), int32(5)).Return(expected, nil)

	svc := service.NewService(repo)
	items, err := svc.List(ctx, 0, 5)

	require.NoError(t, err)
	require.Equal(t, expected, items)
}

func TestService_List_ClampsOverMaxLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	expected := []domain.Task{{ID: uuid.New(), Title: "clamped"}}
	repo.EXPECT().List(ctx, int32(100), int32(0)).Return(expected, nil)

	svc := service.NewService(repo)
	items, err := svc.List(ctx, 999, -5)

	require.NoError(t, err)
	require.Equal(t, expected, items)
}

func TestService_Create_RepositoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchItemName("stub")).Return(errors.New("boom"))

	svc := service.NewService(repo)
	item, err := svc.Create(ctx, "stub")

	require.Error(t, err)
	require.Nil(t, item)
}

func matchItemName(name string) any {
	return matchItemByName{name: name}
}

type matchItemByName struct {
	name string
}

func (m matchItemByName) Matches(x any) bool {
	item, ok := x.(*domain.Task)
	return ok && item.Title == m.name
}

func (m matchItemByName) String() string {
	return "is item named " + m.name
}

func TestService_Update_Happy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(&domain.Task{ID: id, Title: "old", Done: false}, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	svc := service.NewService(repo)
	got, err := svc.Update(ctx, id, "new", true)

	require.NoError(t, err)
	require.Equal(t, "new", got.Title)
	require.True(t, got.Done)
}

func TestService_Update_InvalidTitle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := service.NewService(repo)

	got, err := svc.Update(ctx, uuid.New(), "   ", false)
	require.ErrorIs(t, err, domain.ErrInvalidTitle)
	require.Nil(t, got)
}

func TestService_Update_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()
	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrTaskNotFound)

	svc := service.NewService(repo)
	got, err := svc.Update(ctx, id, "new", false)
	require.ErrorIs(t, err, domain.ErrTaskNotFound)
	require.Nil(t, got)
}

func TestService_Delete_Happy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()
	repo.EXPECT().Delete(ctx, id).Return(nil)

	svc := service.NewService(repo)
	require.NoError(t, svc.Delete(ctx, id))
}

func TestService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()
	repo.EXPECT().Delete(ctx, id).Return(domain.ErrTaskNotFound)

	svc := service.NewService(repo)
	require.ErrorIs(t, svc.Delete(ctx, id), domain.ErrTaskNotFound)
}
