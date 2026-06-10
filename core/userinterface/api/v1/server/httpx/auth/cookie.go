package auth

import (
	"net/http"
	"time"
)

type SessionCookie struct {
	Name   string
	Secure bool
	TTL    time.Duration
}

func (c SessionCookie) Set(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(c.TTL),
	})
}

func (c SessionCookie) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.Name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (c SessionCookie) Read(r *http.Request) string {
	cookie, err := r.Cookie(c.Name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
