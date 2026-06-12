package v1integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
)

func TestSearch_EmptyQuery(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/search", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		Results []any `json:"results"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Empty(t, body.Results)
}

func TestSearch_WithResults(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	fixture.searches.results = append(fixture.searches.results, searchquery.SearchResult{
		Type: "project", ID: "p-1", Title: "Core", Subtitle: "go", URL: "/projects/core",
	})

	res := fixture.do(t, http.MethodGet, "/search?q=core", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		Results []struct {
			Title string `json:"title"`
		} `json:"results"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Len(t, body.Results, 1)
	require.Equal(t, "Core", body.Results[0].Title)
}

func TestGetProjectSource_RejectsMissingParams(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/projects/core/source", nil, cookie)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetProjectSource_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	id := seedProject(f, "core", "Core")
	f.blobs.put(id, "abc123", "main.go", []byte("package main\n"))

	res := f.do(t, http.MethodGet, "/projects/core/source?commit=abc123&path=main.go", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	require.Equal(t, "package main\n", res.Body.String())
}
