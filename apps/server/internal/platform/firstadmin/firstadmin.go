// Package firstadmin resolves the credential used to seed the first admin on a
// fresh database. It is a small testable unit deliberately kept out of the
// composition root: the root is untestable DI wiring, so folding this branching
// logic into main.go would drag the whole wiring file into the coverage
// denominator. Here it stands alone and is covered directly.
package firstadmin

import (
	"fmt"
	"io"

	"github.com/usegavel/gavel/core/infrastructure/iam/crypto"
)

// ResolvePassword returns the admin password to seed: the operator-supplied one
// when non-empty, otherwise a freshly generated secret. generated is true only
// for the generated case, so the caller can log it once after the write commits
// and never for an operator-supplied value. It is used both by first-boot
// (GAVEL_ADMIN_PASSWORD) and by tenant provisioning (--admin-password).
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
