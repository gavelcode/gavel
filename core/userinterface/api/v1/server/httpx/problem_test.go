package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func TestNewProblem(t *testing.T) {
	prob := httpx.NewProblem(http.StatusBadRequest, "invalid input")

	assert.Equal(t, "about:blank", prob.Type)
	assert.Equal(t, "invalid input", prob.Title)
	assert.Equal(t, int32(http.StatusBadRequest), prob.Status)
}

func TestBadRequest(t *testing.T) {
	prob := httpx.BadRequest("missing field")

	assert.Equal(t, int32(http.StatusBadRequest), prob.Status)
	assert.Equal(t, "missing field", prob.Title)
}

func TestUnauthorized(t *testing.T) {
	prob := httpx.Unauthorized("token expired")

	assert.Equal(t, int32(http.StatusUnauthorized), prob.Status)
}

func TestInvalidCredentials(t *testing.T) {
	prob := httpx.InvalidCredentials("wrong password")

	assert.Equal(t, int32(http.StatusUnauthorized), prob.Status)
}

func TestCurrentPasswordIncorrect(t *testing.T) {
	prob := httpx.CurrentPasswordIncorrect("password mismatch")

	assert.Equal(t, int32(http.StatusUnauthorized), prob.Status)
}

func TestNotFound(t *testing.T) {
	prob := httpx.NotFound("resource not found")

	assert.Equal(t, int32(http.StatusNotFound), prob.Status)
}

func TestWriteProblem(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteProblem(rec, http.StatusForbidden, "access denied")

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "Forbidden")
	assert.Contains(t, rec.Body.String(), "access denied")
}

func TestWriteProblemEmptyDetail(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteProblem(rec, http.StatusInternalServerError, "")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Internal Server Error")
}
