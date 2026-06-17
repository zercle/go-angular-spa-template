//go:build unit

// STUB FEATURE — delete internal/features/tasks to start your project.

package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
)

func TestItem_Rename(t *testing.T) {
	item := &domain.Task{
		ID:        uuid.New(),
		Title:     "original",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC().Add(-1 * time.Hour),
	}

	before := item.UpdatedAt
	item.Rename("renamed")

	assert.Equal(t, "renamed", item.Title)
	assert.True(t, item.UpdatedAt.After(before), "expected UpdatedAt to advance")
}

func TestSentinelErrors(t *testing.T) {
	assert.ErrorIs(t, domain.ErrTaskNotFound, domain.ErrTaskNotFound)
	assert.ErrorIs(t, domain.ErrInvalidTitle, domain.ErrInvalidTitle)
}
