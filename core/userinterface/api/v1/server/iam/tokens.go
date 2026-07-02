package iam

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/usegavel/gavel/core/application/iam/issuetoken"
	"github.com/usegavel/gavel/core/application/iam/listmytokens"
	"github.com/usegavel/gavel/core/application/iam/revoketoken"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func (h *Handler) ListMyTokens(ctx context.Context, _ gen.ListMyTokensRequestObject) (gen.ListMyTokensResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListMyTokens401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}

	q, err := listmytokens.NewQuery(principal.UserID, h.deps.Now())
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListMyTokens.Execute(ctx, q)
	if err != nil {
		return nil, err
	}

	items := make([]gen.TokenSummary, 0, len(res.Tokens))
	for _, t := range res.Tokens {
		items = append(items, tokenSummaryFromView(t))
	}
	return gen.ListMyTokens200JSONResponse{Items: items, NextCursor: nil}, nil
}

func (h *Handler) CreateMyToken(ctx context.Context, req gen.CreateMyTokenRequestObject) (gen.CreateMyTokenResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.CreateMyToken401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.CreateMyToken400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	scopes := make([]string, 0, len(req.Body.Scopes))
	for _, sc := range req.Body.Scopes {
		scopes = append(scopes, string(sc))
	}
	if slices.Contains(scopes, issuetoken.ScopeAdmin) && principal.Role != issuetoken.RoleAdmin {
		return gen.CreateMyToken403JSONResponse(httpx.NewProblem(http.StatusForbidden, "only admins can create admin-scoped tokens")), nil
	}

	now := h.deps.Now()
	expiresAt := time.Time{}
	if req.Body.ExpiresInDays != nil && *req.Body.ExpiresInDays > 0 {
		expiresAt = now.Add(time.Duration(*req.Body.ExpiresInDays) * 24 * time.Hour)
	}

	cmd, err := issuetoken.NewCommand(principal.UserID, req.Body.Name, scopes, now, expiresAt)
	if err != nil {
		return gen.CreateMyToken400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.IssueToken.Execute(ctx, cmd)
	if err != nil {
		if apperr.Of(err) == apperr.Validation {
			return gen.CreateMyToken400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		}
		return nil, err
	}
	return createdTokenFromIssue(res), nil
}

func (h *Handler) DeleteMyToken(ctx context.Context, req gen.DeleteMyTokenRequestObject) (gen.DeleteMyTokenResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.DeleteMyToken401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}

	cmd, err := revoketoken.NewCommand(req.Id.String(), principal.UserID, h.deps.Now())
	if err != nil {
		return gen.DeleteMyToken400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.RevokeToken.Execute(ctx, cmd); err != nil {
		if errors.Is(err, revoketoken.ErrUnauthorized) {
			return gen.DeleteMyToken403JSONResponse(httpx.NewProblem(http.StatusForbidden, "not your token")), nil
		}
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.DeleteMyToken404JSONResponse{NotFoundJSONResponse: httpx.NotFound("token not found")}, nil
		case apperr.Validation:
			return gen.DeleteMyToken400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.DeleteMyToken204Response{}, nil
}
