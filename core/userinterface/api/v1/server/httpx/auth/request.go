package auth

import (
	"context"
	"net/http"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

type rawRequestKey struct{}

func ExposeRawRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), rawRequestKey{}, r)))
	})
}

func UserAgentFromContext(ctx context.Context) string {
	if r, ok := ctx.Value(rawRequestKey{}).(*http.Request); ok {
		return r.UserAgent()
	}
	return ""
}

func ClientIPFromContext(ctx context.Context) string {
	if r, ok := ctx.Value(rawRequestKey{}).(*http.Request); ok {
		return httpx.ClientIP(r)
	}
	return ""
}

func SessionCookieFromContext(ctx context.Context, cookieName string) string {
	r, ok := ctx.Value(rawRequestKey{}).(*http.Request)
	if !ok {
		return ""
	}
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
