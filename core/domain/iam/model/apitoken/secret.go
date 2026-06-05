package apitoken

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	apiTokenSecretPrefix   = "gav_"
	apiTokenSecretBodyLen  = 43
	apiTokenSecretTotalLen = len(apiTokenSecretPrefix) + apiTokenSecretBodyLen
	apiTokenPrefixLen      = 12
	apiTokenHashLen        = 64
)

type Secret struct {
	value string
}

func NewSecret(raw string) (Secret, error) {
	if raw == "" {
		return Secret{}, fmt.Errorf("%w: secret must not be empty", ErrInvalid)
	}
	if len(raw) != apiTokenSecretTotalLen {
		return Secret{}, fmt.Errorf("%w: secret must be %d characters", ErrInvalid, apiTokenSecretTotalLen)
	}
	if !strings.HasPrefix(raw, apiTokenSecretPrefix) {
		return Secret{}, fmt.Errorf("%w: secret must start with %q", ErrInvalid, apiTokenSecretPrefix)
	}
	body := raw[len(apiTokenSecretPrefix):]
	if _, err := base64.RawURLEncoding.DecodeString(body); err != nil {
		return Secret{}, fmt.Errorf("%w: secret body must be url-safe base64 (no padding): %v", ErrInvalid, err)
	}
	return Secret{value: raw}, nil
}

func (s Secret) String() string { return s.value }

func (s Secret) Prefix() string {
	if len(s.value) < apiTokenPrefixLen {
		return s.value
	}
	return s.value[:apiTokenPrefixLen]
}

func (s Secret) Equal(other Secret) bool { return s.value == other.value }
