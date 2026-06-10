package httpx_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func TestMaxBodyLimitsRequestBody(t *testing.T) {
	handler := httpx.MaxBody(10)(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			writer.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader("this body is longer than ten bytes")
	req := httptest.NewRequest("POST", "/", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestMaxBodyAllowsSmallBody(t *testing.T) {
	handler := httpx.MaxBody(1024)(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		data, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(data)
	}))

	body := strings.NewReader("small")
	req := httptest.NewRequest("POST", "/", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "small", rec.Body.String())
}
