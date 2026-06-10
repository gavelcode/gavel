package iam

import (
	"context"
	"net/http"

	"github.com/usegavel/gavel/core/application/iam/createuser"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func (h *Handler) CreateUser(ctx context.Context, req gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.CreateUser401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.CreateUser400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	role := createuser.DefaultRole
	if req.Body.Role != nil && *req.Body.Role != "" {
		role = string(*req.Body.Role)
	}

	cmd, err := createuser.NewCommand(principal.TenantID, string(req.Body.Email), req.Body.DisplayName, role, req.Body.Password, true, h.deps.Now())
	if err != nil {
		return gen.CreateUser400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.CreateUser.Execute(ctx, cmd)
	if err != nil {
		switch apperr.Of(err) {
		case apperr.Conflict:
			return gen.CreateUser409JSONResponse(httpx.NewProblem(http.StatusConflict, "email already exists")), nil
		case apperr.Validation:
			return gen.CreateUser400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return createdUserFromResult(res), nil
}
