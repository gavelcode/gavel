package firstadmin

import (
	"fmt"
	"io"

	"github.com/usegavel/gavel/core/infrastructure/iam/crypto"
)

func ResolvePassword(configured string, rng io.Reader) (password string, generated bool, err error) {
	if configured != "" {
		return configured, false, nil
	}
	secret, err := crypto.NewSecretGenerator(rng).NewRandomSecret()
	if err != nil {
		return "", false, fmt.Errorf("generate initial admin password: %w", err)
	}
	return secret, true, nil
}
