package session

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const sessionTokenHashLen = 64

type TokenHash struct {
	value string
}

func NewTokenHash(raw string) (TokenHash, error) {
	if raw == "" {
		return TokenHash{}, fmt.Errorf("%w: session token hash must not be empty", ErrInvalid)
	}
	if len(raw) != sessionTokenHashLen {
		return TokenHash{}, fmt.Errorf("%w: session token hash must be %d hex characters", ErrInvalid, sessionTokenHashLen)
	}
	normalised := strings.ToLower(raw)
	if _, err := hex.DecodeString(normalised); err != nil {
		return TokenHash{}, fmt.Errorf("%w: session token hash must be hex-encoded: %v", ErrInvalid, err)
	}
	return TokenHash{value: normalised}, nil
}

func HashToken(token Token) TokenHash {
	sum := sha256.Sum256([]byte(token.String()))
	return TokenHash{value: hex.EncodeToString(sum[:])}
}

func (h TokenHash) String() string { return h.value }

func (h TokenHash) Equal(other TokenHash) bool { return h.value == other.value }
