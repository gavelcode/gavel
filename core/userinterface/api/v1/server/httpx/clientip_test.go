package httpx_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func TestClientIPFromXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	assert.Equal(t, "1.2.3.4", httpx.ClientIP(req))
}

func TestClientIPFromXForwardedForMultiple(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 9.10.11.12")

	assert.Equal(t, "1.2.3.4", httpx.ClientIP(req))
}

func TestClientIPFromXForwardedForWithSpaces(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "  1.2.3.4  ")

	assert.Equal(t, "1.2.3.4", httpx.ClientIP(req))
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	assert.Equal(t, "192.168.1.1:12345", httpx.ClientIP(req))
}
