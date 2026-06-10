package search_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/search"
)

const testQuery = "core"

type fakeFinder struct {
	items []searchquery.SearchResult
	err   error
}

func (f *fakeFinder) Search(_ context.Context, _ string, _ int) ([]searchquery.SearchResult, error) {
	return f.items, f.err
}

func newTestHandler(finder *fakeFinder) *search.Handler {
	return search.New(search.Deps{
		Search: searchquery.NewHandler(finder),
	})
}

func TestSearchEmptyQueryReturnsEmptyResults(t *testing.T) {
	handler := newTestHandler(&fakeFinder{})

	resp, err := handler.Search(context.Background(), gen.SearchRequestObject{
		Params: gen.SearchParams{},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.Search200JSONResponse)
	require.True(t, ok)
	assert.Empty(t, jsonResp.Results)
}

func TestSearchWhitespaceOnlyQueryReturnsEmptyResults(t *testing.T) {
	handler := newTestHandler(&fakeFinder{})
	query := "   "

	resp, err := handler.Search(context.Background(), gen.SearchRequestObject{
		Params: gen.SearchParams{Q: &query},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.Search200JSONResponse)
	require.True(t, ok)
	assert.Empty(t, jsonResp.Results)
}

func TestSearchReturnsItems(t *testing.T) {
	finder := &fakeFinder{items: []searchquery.SearchResult{
		{Type: "project", ID: "1", Title: "Core", Subtitle: "core project", URL: "/projects/core"},
	}}
	handler := newTestHandler(finder)
	query := testQuery

	resp, err := handler.Search(context.Background(), gen.SearchRequestObject{
		Params: gen.SearchParams{Q: &query},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.Search200JSONResponse)
	require.True(t, ok)
	require.Len(t, jsonResp.Results, 1)
	assert.Equal(t, "project", jsonResp.Results[0].Type)
	assert.Equal(t, "Core", jsonResp.Results[0].Title)
	assert.Equal(t, "/projects/core", jsonResp.Results[0].Url)
}

func TestSearchWithCustomLimit(t *testing.T) {
	finder := &fakeFinder{items: []searchquery.SearchResult{
		{Type: "project", ID: "1", Title: "Core"},
	}}
	handler := newTestHandler(finder)
	query := testQuery
	limit := 5

	resp, err := handler.Search(context.Background(), gen.SearchRequestObject{
		Params: gen.SearchParams{Q: &query, Limit: &limit},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.Search200JSONResponse)
	require.True(t, ok)
	assert.Len(t, jsonResp.Results, 1)
}

func TestSearchFinderError(t *testing.T) {
	finder := &fakeFinder{err: errors.New("db error")}
	handler := newTestHandler(finder)
	query := testQuery

	resp, err := handler.Search(context.Background(), gen.SearchRequestObject{
		Params: gen.SearchParams{Q: &query},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}
