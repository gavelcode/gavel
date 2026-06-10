package iam

import (
	"github.com/oapi-codegen/runtime/types"

	iamcreateuser "github.com/usegavel/gavel/core/application/iam/createuser"
	iamissuetoken "github.com/usegavel/gavel/core/application/iam/issuetoken"
	iamlistmytokens "github.com/usegavel/gavel/core/application/iam/listmytokens"
	"github.com/usegavel/gavel/core/application/iam/login"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func meFromLogin(res login.Result) gen.Me {
	return gen.Me{
		Id:                 httpx.ParseUUIDOrZero(res.UserID),
		Email:              types.Email(res.Email),
		DisplayName:        res.DisplayName,
		Role:               gen.MeRole(res.Role),
		TenantId:           httpx.ParseUUIDOrZero(res.TenantID),
		MustChangePassword: res.MustChangePassword,
	}
}

func meFromPrincipal(principal *auth.Principal) gen.Me {
	return gen.Me{
		Id:                 httpx.ParseUUIDOrZero(principal.UserID),
		Email:              types.Email(principal.Email),
		DisplayName:        principal.DisplayName,
		Role:               gen.MeRole(principal.Role),
		TenantId:           httpx.ParseUUIDOrZero(principal.TenantID),
		MustChangePassword: principal.MustChangePassword,
	}
}

func tokenSummaryFromView(token iamlistmytokens.TokenSummary) gen.TokenSummary {
	return gen.TokenSummary{
		Id:         httpx.ParseUUIDOrZero(token.ID),
		Name:       token.Name,
		Prefix:     token.Prefix,
		Scopes:     token.Scopes,
		CreatedAt:  token.CreatedAt,
		LastUsedAt: token.LastUsedAt,
		ExpiresAt:  token.ExpiresAt,
		IsRevoked:  token.IsRevoked,
		IsExpired:  token.IsExpired,
	}
}

func createdTokenFromIssue(res iamissuetoken.Result) gen.CreateMyToken201JSONResponse {
	return gen.CreateMyToken201JSONResponse{
		Id:     httpx.ParseUUIDOrZero(res.TokenID),
		Name:   res.Name,
		Scopes: res.Scopes,
		Token:  res.PlainSecret,
		Prefix: res.TokenPrefix,
	}
}

func createdUserFromResult(res iamcreateuser.Result) gen.CreateUser201JSONResponse {
	return gen.CreateUser201JSONResponse{
		Id:    httpx.ParseUUIDOrZero(res.UserID),
		Email: types.Email(res.Email),
		Role:  gen.CreatedUserRole(res.Role),
	}
}
