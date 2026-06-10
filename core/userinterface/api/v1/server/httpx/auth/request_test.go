package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func TestExposeRawRequestAndUserAgent(t *testing.T) {
	var capturedUA string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedUA = auth.UserAgentFromContext(req.Context())
	})

	handler := auth.ExposeRawRequest(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "TestBot/1.0")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "TestBot/1.0", capturedUA)
}

func TestUserAgentFromContextNoRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, "", auth.UserAgentFromContext(req.Context()))
}

func TestClientIPFromContextWithRequest(t *testing.T) {
	var capturedIP string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedIP = auth.ClientIPFromContext(req.Context())
	})

	handler := auth.ExposeRawRequest(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "10.0.0.1", capturedIP)
}

func TestClientIPFromContextNoRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, "", auth.ClientIPFromContext(req.Context()))
}

func TestSessionCookieFromContextWithCookie(t *testing.T) {
	var capturedValue string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedValue = auth.SessionCookieFromContext(req.Context(), "session")
	})

	handler := auth.ExposeRawRequest(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token-xyz"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "token-xyz", capturedValue)
}

func TestSessionCookieFromContextNoCookie(t *testing.T) {
	var capturedValue string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedValue = auth.SessionCookieFromContext(req.Context(), "session")
	})

	handler := auth.ExposeRawRequest(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", capturedValue)
}

func TestSessionCookieFromContextNoExposedRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, "", auth.SessionCookieFromContext(req.Context(), "session"))
}
