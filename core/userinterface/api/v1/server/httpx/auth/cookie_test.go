package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func TestSessionCookieSet(t *testing.T) {
	cookie := auth.SessionCookie{Name: "session", Secure: true, TTL: 24 * time.Hour}
	rec := httptest.NewRecorder()

	cookie.Set(rec, "token-abc")

	result := rec.Result()
	require.NoError(t, result.Body.Close())
	require.Len(t, result.Cookies(), 1)
	got := result.Cookies()[0]
	assert.Equal(t, "session", got.Name)
	assert.Equal(t, "token-abc", got.Value)
	assert.Equal(t, "/", got.Path)
	assert.True(t, got.HttpOnly)
	assert.True(t, got.Secure)
	assert.Equal(t, http.SameSiteLaxMode, got.SameSite)
}

func TestSessionCookieClear(t *testing.T) {
	cookie := auth.SessionCookie{Name: "session", Secure: false}
	rec := httptest.NewRecorder()

	cookie.Clear(rec)

	result := rec.Result()
	require.NoError(t, result.Body.Close())
	require.Len(t, result.Cookies(), 1)
	got := result.Cookies()[0]
	assert.Equal(t, "session", got.Name)
	assert.Equal(t, "", got.Value)
	assert.Equal(t, -1, got.MaxAge)
}

func TestSessionCookieRead(t *testing.T) {
	cookie := auth.SessionCookie{Name: "session"}
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "my-token"})

	assert.Equal(t, "my-token", cookie.Read(req))
}

func TestSessionCookieReadMissing(t *testing.T) {
	cookie := auth.SessionCookie{Name: "session"}
	req := httptest.NewRequest("GET", "/", nil)

	assert.Equal(t, "", cookie.Read(req))
}
