package list_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/project/list"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		wantErr bool
	}{
		{name: "validQuery", limit: 10, offset: 0, wantErr: false},
		{name: "zeroLimit", limit: 0, offset: 0, wantErr: true},
		{name: "negativeLimit", limit: -1, offset: 0, wantErr: true},
		{name: "negativeOffset", limit: 10, offset: -1, wantErr: true},
		{name: "validWithOffset", limit: 5, offset: 20, wantErr: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := list.NewQuery(testTenant.String(), testCase.limit, testCase.offset)
			if testCase.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, list.ErrInvalidQuery)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.limit, query.Limit())
			assert.Equal(t, testCase.offset, query.Offset())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { list.NewHandler(nil) })
}

func TestExecuteReturnsResult(t *testing.T) {
	now := time.Now()
	expected := []list.ProjectSummary{
		{
			ID:            "proj-1",
			Key:           "my-project",
			Name:          "My Project",
			DefaultBranch: "main",
			LatestVerdict: "pass",
			TotalFindings: 10,
			CreatedAt:     now,
		},
	}
	fake := &fakeProjectLister{items: expected, total: 42}
	h := list.NewHandler(fake)

	query, err := list.NewQuery(testTenant.String(), 10, 0)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, expected, result.Items)
	assert.Equal(t, 42, result.Total)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("connection failed")
	fake := &fakeProjectLister{err: queryErr}
	h := list.NewHandler(fake)

	query, err := list.NewQuery(testTenant.String(), 10, 0)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := list.NewQuery(testTenant.String(), 0, 0)
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
