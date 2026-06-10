package iam

import (
	"context"
	"errors"
	"time"

	"github.com/usegavel/gavel/core/application/iam/changepassword"
	"github.com/usegavel/gavel/core/application/iam/createuser"
	"github.com/usegavel/gavel/core/application/iam/issuetoken"
	"github.com/usegavel/gavel/core/application/iam/listmytokens"
	"github.com/usegavel/gavel/core/application/iam/login"
	"github.com/usegavel/gavel/core/application/iam/logout"
	"github.com/usegavel/gavel/core/application/iam/revoketoken"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type Deps struct {
	Login          *login.Handler
	Logout         *logout.Handler
	ChangePassword *changepassword.Handler
	IssueToken     *issuetoken.Handler
	RevokeToken    *revoketoken.Handler
	ListMyTokens   *listmytokens.Handler
	CreateUser     *createuser.Handler
	Cookie         auth.SessionCookie
	DefaultTenant  string
	Now            func() time.Time
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	return &Handler{deps: deps}
}

func (h *Handler) CreateSession(ctx context.Context, req gen.CreateSessionRequestObject) (gen.CreateSessionResponseObject, error) {
	if req.Body == nil {
		return gen.CreateSession400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	tenant := h.deps.DefaultTenant
	if req.Body.Tenant != nil && *req.Body.Tenant != "" {
		tenant = *req.Body.Tenant
	}
	now := h.deps.Now()
	cmd, err := login.NewCommand(
		tenant,
		string(req.Body.Email),
		req.Body.Password,
		auth.UserAgentFromContext(ctx),
		auth.ClientIPFromContext(ctx),
		now,
		h.deps.Cookie.TTL,
	)
	if err != nil {
		return gen.CreateSession400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}

	res, err := h.deps.Login.Execute(ctx, cmd)
	if err != nil {
		if errors.Is(err, login.ErrInvalidCredentials) {
			return gen.CreateSession401JSONResponse{InvalidCredentialsJSONResponse: httpx.InvalidCredentials("invalid credentials")}, nil
		}
		return nil, err
	}

	return sessionCreated{
		me:     meFromLogin(res),
		cookie: h.deps.Cookie,
		token:  res.SessionToken,
	}, nil
}

func (h *Handler) DeleteCurrentSession(ctx context.Context, _ gen.DeleteCurrentSessionRequestObject) (gen.DeleteCurrentSessionResponseObject, error) {
	token := auth.SessionCookieFromContext(ctx, h.deps.Cookie.Name)
	if token != "" {
		cmd, err := logout.NewCommand(token, h.deps.Now())
		if err == nil {
			_, _ = h.deps.Logout.Execute(ctx, cmd)
		}
	}
	return sessionDeleted{cookie: h.deps.Cookie}, nil
}

func (h *Handler) GetMe(ctx context.Context, _ gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.GetMe401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	return gen.GetMe200JSONResponse(meFromPrincipal(principal)), nil
}

func (h *Handler) ChangeMyPassword(ctx context.Context, req gen.ChangeMyPasswordRequestObject) (gen.ChangeMyPasswordResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ChangeMyPassword401JSONResponse{CurrentPasswordIncorrectJSONResponse: httpx.CurrentPasswordIncorrect("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.ChangeMyPassword400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	cmd, err := changepassword.NewCommand(principal.UserID, req.Body.CurrentPassword, req.Body.NewPassword, h.deps.Now())
	if err != nil {
		return gen.ChangeMyPassword400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.ChangePassword.Execute(ctx, cmd); err != nil {
		switch {
		case errors.Is(err, changepassword.ErrCurrentPasswordWrong):
			return gen.ChangeMyPassword401JSONResponse{CurrentPasswordIncorrectJSONResponse: httpx.CurrentPasswordIncorrect("current password incorrect")}, nil
		case apperr.Of(err) == apperr.Validation:
			return gen.ChangeMyPassword400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return passwordChanged{cookie: h.deps.Cookie}, nil
}
