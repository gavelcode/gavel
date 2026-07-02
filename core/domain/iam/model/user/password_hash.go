package user

import (
	"fmt"
	"strings"
)

const argon2idPrefix = "$argon2id$"
const argon2idValidSegmentCount = 5

type PasswordHash struct {
	value string
}

func NewPasswordHash(encoded string) (PasswordHash, error) {
	if strings.TrimSpace(encoded) == "" {
		return PasswordHash{}, fmt.Errorf("%w: password hash must not be empty", ErrInvalidUser)
	}
	if !strings.HasPrefix(encoded, argon2idPrefix) {
		return PasswordHash{}, fmt.Errorf("%w: password hash must be argon2id-encoded", ErrInvalidUser)
	}
	if strings.Count(encoded, "$") != argon2idValidSegmentCount {
		return PasswordHash{}, fmt.Errorf("%w: password hash must have %d '$'-separated segments", ErrInvalidUser, argon2idValidSegmentCount)
	}
	return PasswordHash{value: encoded}, nil
}

func (h PasswordHash) String() string { return h.value }

func (h PasswordHash) Equal(other PasswordHash) bool { return h.value == other.value }
