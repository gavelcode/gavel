package search_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/supporting/search"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		limit   int
		wantErr bool
	}{
		{name: "validQuery", query: "findbugs", limit: 10, wantErr: false},
		{name: "emptyQuery", query: "", limit: 10, wantErr: true},
		{name: "whitespaceQuery", query: "  ", limit: 10, wantErr: true},
		{name: "zeroLimit", query: "findbugs", limit: 0, wantErr: true},
		{name: "negativeLimit", query: "findbugs", limit: -1, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := search.NewQuery(testCase.query, testCase.limit)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.query, query.Text())
			assert.Equal(t, testCase.limit, query.Limit())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { search.NewHandler(nil) })
}

func TestExecuteReturnsResults(t *testing.T) {
	expected := []search.SearchResult{
		{Type: "project", ID: "p-1", Title: "My Project", Subtitle: "key-1", URL: "/projects/key-1"},
	}
	fake := &fakeSearchQuery{results: expected}
	h := search.NewHandler(fake)

	query, err := search.NewQuery("my", 10)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "project", result.Items[0].Type)
	assert.Equal(t, "My Project", result.Items[0].Title)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("search engine down")
	fake := &fakeSearchQuery{err: queryErr}
	h := search.NewHandler(fake)

	query, err := search.NewQuery("findbugs", 10)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := search.NewQuery("", 10)
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
