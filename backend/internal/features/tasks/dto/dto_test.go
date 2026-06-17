//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package dto_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/dto"
)

func TestCreateTaskRequest_Validation(t *testing.T) {
	v := validator.New()

	valid := dto.CreateTaskRequest{Title: "valid name"}
	assert.NoError(t, v.Struct(valid))

	empty := dto.CreateTaskRequest{Title: ""}
	assert.Error(t, v.Struct(empty))

	long := dto.CreateTaskRequest{Title: string(make([]byte, 256))}
	assert.Error(t, v.Struct(long))
}

func TestListTasksRequest_Validation(t *testing.T) {
	v := validator.New()

	valid := dto.ListTasksRequest{Limit: 10, Offset: 0}
	assert.NoError(t, v.Struct(valid))

	defaultLimit := dto.ListTasksRequest{}
	assert.NoError(t, v.Struct(defaultLimit))

	highLimit := dto.ListTasksRequest{Limit: 101, Offset: 0}
	assert.Error(t, v.Struct(highLimit))

	negativeOffset := dto.ListTasksRequest{Limit: 10, Offset: -1}
	assert.Error(t, v.Struct(negativeOffset))
}
