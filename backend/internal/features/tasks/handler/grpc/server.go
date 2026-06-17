// STUB FEATURE — delete internal/features/tasks to start your project.

package grpchandler

import (
	"context"
	"time"

	"github.com/google/uuid"

	pb "github.com/zercle/go-angular-spa-template/api/pb/tasks/v1"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	sharederrors "github.com/zercle/go-angular-spa-template/internal/shared/errors"
)

// nolint:wrapcheck // gRPC handlers return the shared mapper error directly.

// Server implements the tasks.v1.TaskService gRPC contract.
type Server struct {
	pb.UnimplementedTaskServiceServer
	service domain.Service
}

// NewServer returns a gRPC handler for the tasks feature.
func NewServer(service domain.Service) *Server {
	return &Server{service: service}
}

// CreateTask creates a new task.
func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	task, err := s.service.Create(ctx, req.Title)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}
	return mapDomainToPB(task), nil
}

// GetTask retrieves a task by ID.
func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, sharederrors.GRPCErr(domain.ErrInvalidID)
	}
	task, err := s.service.Get(ctx, id)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}
	return mapDomainToPB(task), nil
}

// ListTasks returns a paginated list of tasks.
func (s *Server) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tasks, err := s.service.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}
	resp := &pb.ListTasksResponse{Tasks: make([]*pb.Task, len(tasks))}
	for i := range tasks {
		resp.Tasks[i] = mapDomainToPB(&tasks[i])
	}
	return resp, nil
}

// UpdateTask updates an existing task.
func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, sharederrors.GRPCErr(domain.ErrInvalidID)
	}
	task, err := s.service.Update(ctx, id, req.Title, req.Done)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}
	return mapDomainToPB(task), nil
}

// DeleteTask removes a task.
func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, sharederrors.GRPCErr(domain.ErrInvalidID)
	}
	if err := s.service.Delete(ctx, id); err != nil {
		return nil, sharederrors.GRPCErr(err)
	}
	return &pb.DeleteTaskResponse{}, nil
}

func mapDomainToPB(task *domain.Task) *pb.Task {
	if task == nil {
		return nil
	}
	return &pb.Task{
		Id:        task.ID.String(),
		Title:     task.Title,
		Done:      task.Done,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
		UpdatedAt: task.UpdatedAt.Format(time.RFC3339),
	}
}
