package list_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/gavelspace/list"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		wantErr bool
	}{
		{name: "validQuery", limit: 10, offset: 0, wantErr: false},
		{name: "validQueryWithOffset", limit: 5, offset: 20, wantErr: false},
		{name: "zeroLimit", limit: 0, offset: 0, wantErr: true},
		{name: "negativeLimit", limit: -1, offset: 0, wantErr: true},
		{name: "negativeOffset", limit: 10, offset: -1, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := list.NewQuery(testTenant, testCase.limit, testCase.offset)
			if testCase.wantErr {
				assert.Error(t, err)
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
	expected := []list.GavelspaceSummary{
		{Name: "my-monorepo", ProjectCount: 3, CreatedAt: now},
		{Name: "other-repo", ProjectCount: 1, CreatedAt: now},
	}
	fake := &fakeGavelspaceLister{items: expected, total: 2}
	h := list.NewHandler(fake)

	query, err := list.NewQuery(testTenant, 10, 0)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "my-monorepo", result.Items[0].Name)
	assert.Equal(t, 3, result.Items[0].ProjectCount)
	assert.Equal(t, 2, result.Total)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("database unavailable")
	fake := &fakeGavelspaceLister{err: queryErr}
	h := list.NewHandler(fake)

	query, err := list.NewQuery(testTenant, 10, 0)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := list.NewQuery(testTenant, 0, 0)
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}

func TestNewQueryRejectsEmptyTenant(t *testing.T) {
	_, err := list.NewQuery("", 10, 0)
	require.Error(t, err)
}

func TestExecuteInvalidTenant(t *testing.T) {
	h := list.NewHandler(&fakeGavelspaceLister{})
	query, err := list.NewQuery("not-a-uuid", 10, 0)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant id")
}
