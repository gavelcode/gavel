package apitoken

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type SecretHash struct {
	value string
}

func NewSecretHash(raw string) (SecretHash, error) {
	if raw == "" {
		return SecretHash{}, fmt.Errorf("%w: hash must not be empty", ErrInvalid)
	}
	if len(raw) != apiTokenHashLen {
		return SecretHash{}, fmt.Errorf("%w: hash must be %d hex characters", ErrInvalid, apiTokenHashLen)
	}
	normalised := strings.ToLower(raw)
	if _, err := hex.DecodeString(normalised); err != nil {
		return SecretHash{}, fmt.Errorf("%w: hash must be hex-encoded: %v", ErrInvalid, err)
	}
	return SecretHash{value: normalised}, nil
}

func HashSecret(secret Secret) SecretHash {
	sum := sha256.Sum256([]byte(secret.String()))
	return SecretHash{value: hex.EncodeToString(sum[:])}
}

func (h SecretHash) String() string { return h.value }

func (h SecretHash) Equal(other SecretHash) bool { return h.value == other.value }
