package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

type APITokenRepository interface {
	Save(ctx context.Context, token apitoken.APIToken) error
	ByID(ctx context.Context, id apitoken.APITokenID) (apitoken.APIToken, error)
	ByTokenHash(ctx context.Context, hash apitoken.SecretHash) (apitoken.APIToken, error)
	ListByUser(ctx context.Context, userID user.UserID) ([]apitoken.APIToken, error)
}
