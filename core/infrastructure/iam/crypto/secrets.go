package crypto

import (
	"encoding/base64"
	"fmt"
	"io"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

const randomBodyBytes = 32

type SecretGenerator struct {
	rng io.Reader
}

var _ service.SecretGenerator = (*SecretGenerator)(nil)

func NewSecretGenerator(rng io.Reader) *SecretGenerator {
	return &SecretGenerator{rng: rng}
}

func (g *SecretGenerator) NewSessionToken() (session.Token, error) {
	body, err := g.randomBase64Body()
	if err != nil {
		return session.Token{}, err
	}
	return session.NewToken(body)
}

func (g *SecretGenerator) NewAPITokenSecret() (apitoken.Secret, error) {
	body, err := g.randomBase64Body()
	if err != nil {
		return apitoken.Secret{}, err
	}
	return apitoken.NewSecret("gav_" + body)
}

func (g *SecretGenerator) randomBase64Body() (string, error) {
	buf := make([]byte, randomBodyBytes)
	if _, err := io.ReadFull(g.rng, buf); err != nil {
		return "", fmt.Errorf("crypto: read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
