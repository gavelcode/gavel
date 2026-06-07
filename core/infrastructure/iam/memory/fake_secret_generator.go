package iam

import (
	"encoding/base64"
	"sync"
	"sync/atomic"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type FakeSecretGenerator struct {
	mu sync.Mutex
	n  atomic.Uint64
}

var _ service.SecretGenerator = (*FakeSecretGenerator)(nil)

func NewFakeSecretGenerator() *FakeSecretGenerator { return &FakeSecretGenerator{} }

func (g *FakeSecretGenerator) NewSessionToken() (session.Token, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	raw := encodeFakeBody(g.n.Add(1))
	return session.NewToken(raw)
}

func (g *FakeSecretGenerator) NewAPITokenSecret() (apitoken.Secret, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	raw := "gav_" + encodeFakeBody(g.n.Add(1))
	return apitoken.NewSecret(raw)
}

const (
	fakeTokenLength = 32
	byteMask        = 0xff
)

func encodeFakeBody(n uint64) string {
	buf := make([]byte, fakeTokenLength)
	for i := range buf {
		buf[i] = byte((n + uint64(i)) & byteMask)
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
