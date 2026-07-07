package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/gavelspace/get"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "validName", input: "my-monorepo", wantErr: false},
		{name: "emptyName", input: "", wantErr: true},
		{name: "whitespaceName", input: "  ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := get.NewQuery(testTenant, testCase.input)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.input, query.Name())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { get.NewHandler(nil) })
}

func TestExecuteReturnsResult(t *testing.T) {
	now := time.Now()
	expected := &get.GavelspaceDetail{
		Name: "my-monorepo",
		Projects: []get.ProjectRefView{
			{ID: "p-1", Key: "backend", Name: "Backend Service", LatestVerdict: "pass"},
			{ID: "p-2", Key: "frontend", Name: "Frontend App", LatestVerdict: "fail"},
		},
		CreatedAt: now,
	}
	fake := &fakeGavelspaceGetter{result: expected}
	h := get.NewHandler(fake)

	query, err := get.NewQuery(testTenant, "my-monorepo")
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, "my-monorepo", result.Name)
	require.Len(t, result.Projects, 2)
	assert.Equal(t, "backend", result.Projects[0].Key)
	assert.Equal(t, "pass", result.Projects[0].LatestVerdict)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("gavelspace not found")
	fake := &fakeGavelspaceGetter{err: queryErr}
	h := get.NewHandler(fake)

	query, err := get.NewQuery(testTenant, "missing-repo")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := get.NewQuery(testTenant, "")
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}

func TestNewQueryRejectsEmptyTenant(t *testing.T) {
	_, err := get.NewQuery("", "my-monorepo")
	require.Error(t, err)
}

func TestExecuteInvalidTenant(t *testing.T) {
	h := get.NewHandler(&fakeGavelspaceGetter{})
	query, err := get.NewQuery("not-a-uuid", "my-monorepo")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant id")
}
