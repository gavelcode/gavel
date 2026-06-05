package service

import "github.com/usegavel/gavel/core/domain/iam/model/user"

type PasswordHasher interface {
	Hash(plain string) (user.PasswordHash, error)
	Verify(plain string, hash user.PasswordHash) (bool, error)
}
