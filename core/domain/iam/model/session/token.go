package session

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const sessionTokenLen = 43

type Token struct {
	value string
}

func NewToken(raw string) (Token, error) {
	if strings.TrimSpace(raw) == "" {
		return Token{}, fmt.Errorf("%w: session token must not be empty", ErrInvalid)
	}
	if len(raw) != sessionTokenLen {
		return Token{}, fmt.Errorf("%w: session token must be %d characters", ErrInvalid, sessionTokenLen)
	}
	if _, err := base64.RawURLEncoding.DecodeString(raw); err != nil {
		return Token{}, fmt.Errorf("%w: session token must be url-safe base64 (no padding): %v", ErrInvalid, err)
	}
	return Token{value: raw}, nil
}

func (t Token) String() string { return t.value }

func (t Token) Equal(other Token) bool { return t.value == other.value }
