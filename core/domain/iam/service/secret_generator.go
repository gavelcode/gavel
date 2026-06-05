package service

import (
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
)

type SecretGenerator interface {
	NewSessionToken() (session.Token, error)
	NewAPITokenSecret() (apitoken.Secret, error)
}
