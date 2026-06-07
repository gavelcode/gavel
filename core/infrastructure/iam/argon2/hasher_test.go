package argon2_test

import (
	"crypto/rand"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/infrastructure/iam/argon2"
)

func TestHasherRoundTrip(t *testing.T) {
	hasher := argon2.NewWithConfig(rand.Reader, argon2.Config{Time: 1, Memory: 8 * 1024, Threads: 1, KeyLen: 16, SaltLen: 8})

	hash, err := hasher.Hash("hunter22")
	require.NoError(t, err)
	assert.NotEmpty(t, hash.String())

	matched, err := hasher.Verify("hunter22", hash)
	require.NoError(t, err)
	assert.True(t, matched)

	matched, err = hasher.Verify("wrong", hash)
	require.NoError(t, err)
	assert.False(t, matched)
}

func TestHasherHashRejectsInvalidConfig(t *testing.T) {
	h := argon2.NewWithConfig(rand.Reader, argon2.Config{SaltLen: 0, KeyLen: 16, Time: 1, Memory: 8 * 1024, Threads: 1})

	_, err := h.Hash("password")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestHasherVerifyRejectsBadVersion(t *testing.T) {
	h := argon2.New(rand.Reader)
	hash, err := user.NewPasswordHash("$argon2id$vNaN$m=65536,t=3,p=4$dGVzdA$dGVzdA")
	require.NoError(t, err)

	_, err = h.Verify("password", hash)

	require.Error(t, err)
}

func TestHasherVerifyRejectsBadParams(t *testing.T) {
	h := argon2.New(rand.Reader)
	hash, err := user.NewPasswordHash("$argon2id$v=19$bad-params$dGVzdA$dGVzdA")
	require.NoError(t, err)

	_, err = h.Verify("password", hash)

	require.Error(t, err)
}

func TestHasherVerifyRejectsBadSaltBase64(t *testing.T) {
	h := argon2.New(rand.Reader)
	hash, err := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$!!!invalid!!!$dGVzdA")
	require.NoError(t, err)

	_, err = h.Verify("password", hash)

	require.Error(t, err)
}

func TestHasherVerifyRejectsBadKeyBase64(t *testing.T) {
	h := argon2.New(rand.Reader)
	hash, err := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$dGVzdA$!!!invalid!!!")
	require.NoError(t, err)

	_, err = h.Verify("password", hash)

	require.Error(t, err)
}

func TestHasherDifferentSaltsProduceDifferentHashes(t *testing.T) {
	hasher := argon2.NewWithConfig(rand.Reader, argon2.Config{Time: 1, Memory: 8 * 1024, Threads: 1, KeyLen: 16, SaltLen: 8})

	hashA, err := hasher.Hash("samepass")
	require.NoError(t, err)
	hashB, err := hasher.Hash("samepass")
	require.NoError(t, err)
	assert.NotEqual(t, hashA.String(), hashB.String(), "different salts must yield different hashes")

	ok, _ := hasher.Verify("samepass", hashA)
	assert.True(t, ok)
	ok, _ = hasher.Verify("samepass", hashB)
	assert.True(t, ok)
}

func TestHasherHashReturnsErrorOnReaderFailure(t *testing.T) {
	h := argon2.NewWithConfig(iotest.ErrReader(errors.New("rng broken")), argon2.Config{
		Time: 1, Memory: 8 * 1024, Threads: 1, KeyLen: 16, SaltLen: 8,
	})

	_, err := h.Hash("password")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read salt")
}
