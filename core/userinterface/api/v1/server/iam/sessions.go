package iam

import (
	"net/http"

	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type sessionCreated struct {
	me     gen.Me
	cookie auth.SessionCookie
	token  string
}

func (r sessionCreated) VisitCreateSessionResponse(w http.ResponseWriter) error {
	r.cookie.Set(w, r.token)
	return gen.CreateSession200JSONResponse(r.me).VisitCreateSessionResponse(w)
}

type sessionDeleted struct {
	cookie auth.SessionCookie
}

func (r sessionDeleted) VisitDeleteCurrentSessionResponse(w http.ResponseWriter) error {
	r.cookie.Clear(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type passwordChanged struct {
	cookie auth.SessionCookie
}

func (r passwordChanged) VisitChangeMyPasswordResponse(w http.ResponseWriter) error {
	r.cookie.Clear(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}
