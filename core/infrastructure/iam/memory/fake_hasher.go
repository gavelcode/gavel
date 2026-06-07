package iam

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

const fakeHashPrefix = "$argon2id$v=19$m=65536,t=3,p=4$dGVzdHNhbHRzYWx0c2FsdA$"

type FakeHasher struct{}

var _ service.PasswordHasher = (*FakeHasher)(nil)

func NewFakeHasher() *FakeHasher { return &FakeHasher{} }

func (h *FakeHasher) Hash(plain string) (user.PasswordHash, error) {
	encoded := base64.RawStdEncoding.EncodeToString([]byte(plain))
	return user.NewPasswordHash(fakeHashPrefix + encoded)
}

const argon2HashParts = 6

func (h *FakeHasher) Verify(plain string, hash user.PasswordHash) (bool, error) {
	parts := strings.Split(hash.String(), "$")
	if len(parts) != argon2HashParts {
		return false, errors.New("fake hasher: malformed hash")
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("fake hasher: %w", err)
	}
	return string(expected) == plain, nil
}
