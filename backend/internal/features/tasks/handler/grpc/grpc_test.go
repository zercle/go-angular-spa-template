//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package grpchandler_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pb "github.com/zercle/go-angular-spa-template/api/pb/tasks/v1"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	grpchandler "github.com/zercle/go-angular-spa-template/internal/features/tasks/handler/grpc"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/service/mock"
)

func TestServer_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	item := &domain.Task{ID: uuid.New(), Title: "grpc-item", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	svc.EXPECT().Create(gomock.Any(), "grpc-item").Return(item, nil)

	resp, err := server.CreateTask(context.Background(), &pb.CreateTaskRequest{Title: "grpc-item"})
	require.NoError(t, err)
	assert.Equal(t, item.ID.String(), resp.Id)
	assert.Equal(t, item.Title, resp.Title)
}

func TestServer_CreateTask_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	svc.EXPECT().Create(gomock.Any(), "bad").Return(nil, domain.ErrInvalidTitle)

	resp, err := server.CreateTask(context.Background(), &pb.CreateTaskRequest{Title: "bad"})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_GetTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	item := &domain.Task{ID: id, Title: "grpc-item", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	svc.EXPECT().Get(gomock.Any(), id).Return(item, nil)

	resp, err := server.GetTask(context.Background(), &pb.GetTaskRequest{Id: id.String()})
	require.NoError(t, err)
	assert.Equal(t, id.String(), resp.Id)
}

func TestServer_GetTask_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	svc.EXPECT().Get(gomock.Any(), id).Return(nil, domain.ErrTaskNotFound)

	resp, err := server.GetTask(context.Background(), &pb.GetTaskRequest{Id: id.String()})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_ListTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	items := []domain.Task{{ID: id, Title: "grpc-item", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}
	svc.EXPECT().List(gomock.Any(), int32(10), int32(0)).Return(items, nil)

	resp, err := server.ListTasks(context.Background(), &pb.ListTasksRequest{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, resp.Tasks, 1)
}

func TestServer_GetTask_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	resp, err := server.GetTask(context.Background(), &pb.GetTaskRequest{Id: "not-a-uuid"})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_UpdateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	svc.EXPECT().Update(gomock.Any(), id, "renamed", true).
		Return(&domain.Task{ID: id, Title: "renamed", Done: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}, nil)

	resp, err := server.UpdateTask(context.Background(), &pb.UpdateTaskRequest{Id: id.String(), Title: "renamed", Done: true})
	require.NoError(t, err)
	assert.Equal(t, "renamed", resp.Title)
	assert.True(t, resp.Done)
}

func TestServer_UpdateTask_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	resp, err := server.UpdateTask(context.Background(), &pb.UpdateTaskRequest{Id: "nope", Title: "x"})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_DeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	svc.EXPECT().Delete(gomock.Any(), id).Return(nil)

	resp, err := server.DeleteTask(context.Background(), &pb.DeleteTaskRequest{Id: id.String()})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestServer_DeleteTask_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	resp, err := server.DeleteTask(context.Background(), &pb.DeleteTaskRequest{Id: "nope"})
	require.Error(t, err)
	assert.Nil(t, resp)
}
