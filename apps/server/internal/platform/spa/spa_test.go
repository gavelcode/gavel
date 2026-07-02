package spa_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/apps/server/internal/platform/spa"
)

func TestHandlerServesExistingFile(t *testing.T) {
	frontend := fstest.MapFS{
		"index.html":       {Data: []byte("<html>app</html>")},
		"assets/style.css": {Data: []byte("body{}")},
	}
	handler := spa.Handler(frontend)

	req := httptest.NewRequest(http.MethodGet, "/assets/style.css", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "body{}")
}

func TestHandlerFallsBackToIndex(t *testing.T) {
	frontend := fstest.MapFS{
		"index.html": {Data: []byte("<html>app</html>")},
	}
	handler := spa.Handler(frontend)

	req := httptest.NewRequest(http.MethodGet, "/some/deep/route", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>app</html>")
}

func TestHandlerServesRootAsIndex(t *testing.T) {
	frontend := fstest.MapFS{
		"index.html": {Data: []byte("<html>root</html>")},
	}
	handler := spa.Handler(frontend)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>root</html>")
}
