package listfindings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	list "github.com/usegavel/gavel/core/application/casefile/listfindings"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		filters list.Filters
		limit   int
		offset  int
		wantErr bool
	}{
		{name: "validQuery", filters: list.Filters{ProjectID: "proj-1"}, limit: 10, offset: 0, wantErr: false},
		{name: "emptyFilters", filters: list.Filters{}, limit: 10, offset: 0, wantErr: false},
		{name: "zeroLimit", filters: list.Filters{}, limit: 0, offset: 0, wantErr: true},
		{name: "negativeLimit", filters: list.Filters{}, limit: -1, offset: 0, wantErr: true},
		{name: "negativeOffset", filters: list.Filters{}, limit: 10, offset: -1, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := list.NewQuery("22222222-2222-2222-2222-222222222222", testCase.filters, testCase.limit, testCase.offset)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.filters, query.Filters())
			assert.Equal(t, testCase.limit, query.Limit())
			assert.Equal(t, testCase.offset, query.Offset())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { list.NewHandler(nil) })
}

func TestExecuteReturnsResult(t *testing.T) {
	expected := []list.FindingView{
		{
			Tool:          "pmd",
			RuleID:        "UnusedVariable",
			Severity:      "warning",
			FilePath:      "src/Main.java",
			Line:          42,
			Message:       "Unused variable 'x'",
			FingerprintID: "fp-1",
			Status:        "new",
			Source:        "static-analysis",
		},
		{
			Tool:          "spotbugs",
			RuleID:        "NP_NULL_ON_SOME_PATH",
			Severity:      "error",
			FilePath:      "src/Service.java",
			Line:          17,
			Message:       "Possible null pointer dereference",
			FingerprintID: "fp-2",
			Status:        "baseline",
			Source:        "static-analysis",
		},
	}
	fake := &fakeFindingLister{items: expected, total: 25}
	h := list.NewHandler(fake)

	query, err := list.NewQuery("22222222-2222-2222-2222-222222222222", list.Filters{ProjectID: "proj-1"}, 10, 0)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, expected, result.Items)
	assert.Equal(t, 25, result.Total)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("connection failed")
	fake := &fakeFindingLister{err: queryErr}
	h := list.NewHandler(fake)

	query, err := list.NewQuery("22222222-2222-2222-2222-222222222222", list.Filters{}, 10, 0)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := list.NewQuery("22222222-2222-2222-2222-222222222222", list.Filters{}, 0, 0)
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
