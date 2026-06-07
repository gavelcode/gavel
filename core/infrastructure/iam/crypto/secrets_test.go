package crypto_test

import (
	"crypto/rand"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/iam/crypto"
)

func TestSecretGeneratorYieldsValidUniqueTokens(t *testing.T) {
	g := crypto.NewSecretGenerator(rand.Reader)

	a, err := g.NewSessionToken()
	require.NoError(t, err)
	b, err := g.NewSessionToken()
	require.NoError(t, err)
	assert.NotEqual(t, a.String(), b.String(), "session tokens must be unique across calls")
	assert.Equal(t, 43, len(a.String()))
}

func TestSecretGeneratorYieldsValidUniqueAPITokens(t *testing.T) {
	g := crypto.NewSecretGenerator(rand.Reader)

	secretA, err := g.NewAPITokenSecret()
	require.NoError(t, err)
	b, err := g.NewAPITokenSecret()
	require.NoError(t, err)
	assert.NotEqual(t, secretA.String(), b.String())
	assert.Equal(t, 47, len(secretA.String()), "gav_ prefix + 43-char body")
	assert.Equal(t, "gav_", secretA.String()[:4])
}

func TestNewSessionTokenReturnsErrorOnReaderFailure(t *testing.T) {
	g := crypto.NewSecretGenerator(iotest.ErrReader(errors.New("rng broken")))

	_, err := g.NewSessionToken()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read random bytes")
}

func TestNewAPITokenSecretReturnsErrorOnReaderFailure(t *testing.T) {
	g := crypto.NewSecretGenerator(iotest.ErrReader(errors.New("rng broken")))

	_, err := g.NewAPITokenSecret()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read random bytes")
}
