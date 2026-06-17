//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package httphandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	httphandler "github.com/zercle/go-angular-spa-template/internal/features/tasks/handler/http"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/service/mock"
	sharederrors "github.com/zercle/go-angular-spa-template/internal/shared/errors"
)

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	sharederrors.RegisterSentinel(domain.ErrTaskNotFound, sharederrors.ErrNotFound)
	sharederrors.RegisterSentinel(domain.ErrInvalidTitle, sharederrors.ErrInvalidInput)
	sharederrors.RegisterSentinel(domain.ErrInvalidID, sharederrors.ErrInvalidInput)

	e := echo.New()
	e.Validator = newValidator(t)
	svc := mock.NewMockService(gomock.NewController(t))
	h := httphandler.New(svc)

	h.Register(e.Group("/api/v1"))

	return e, svc
}

func newValidator(t *testing.T) echo.Validator {
	t.Helper()
	return &validatorAdapter{v: validator.New()}
}

type validatorAdapter struct {
	v *validator.Validate
}

func (v *validatorAdapter) Validate(i any) error {
	return v.v.Struct(i)
}

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Create(ctx, "stub").Return(&domain.Task{ID: id, Title: "stub"}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/tasks", bytes.NewReader([]byte(`{"title":"stub"}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "stub")
}

func TestHandler_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id).Return(&domain.Task{ID: id, Title: "found"}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/tasks/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id).Return(nil, domain.ErrTaskNotFound)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/tasks/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "NOT_FOUND", body["error"])
}

func TestHandler_Create_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, _ := setupTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/tasks", bytes.NewReader([]byte(`{"title":""}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Create_ServiceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().Create(ctx, "stub").Return(nil, errors.New("boom"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/tasks", bytes.NewReader([]byte(`{"title":"stub"}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_List_NoQueryParams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().List(ctx, int32(0), int32(0)).Return([]domain.Task{{ID: uuid.New(), Title: "default"}}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/tasks", nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Update(ctx, id, "renamed", true).
		Return(&domain.Task{ID: id, Title: "renamed", Done: true}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPut, "/api/v1/tasks/"+id.String(),
		bytes.NewReader([]byte(`{"title":"renamed","done":true}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "renamed")
}

func TestHandler_Update_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, _ := setupTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPut, "/api/v1/tasks/not-a-uuid",
		bytes.NewReader([]byte(`{"title":"x","done":false}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Delete(ctx, id).Return(nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/tasks/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Delete(ctx, id).Return(domain.ErrTaskNotFound)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/tasks/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
