package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

func TestFakeHasherRoundTrip(t *testing.T) {
	hasher := memiam.NewFakeHasher()
	hash, err := hasher.Hash("hunter2")
	require.NoError(t, err)

	matched, err := hasher.Verify("hunter2", hash)
	require.NoError(t, err)
	assert.True(t, matched)

	matched, err = hasher.Verify("wrong", hash)
	require.NoError(t, err)
	assert.False(t, matched)
}

func TestFakeHasherVerifyRejectsInvalidBase64(t *testing.T) {
	h := memiam.NewFakeHasher()
	hash, err := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$salt$!!!not-base64!!!")
	require.NoError(t, err)

	_, err = h.Verify("password", hash)

	require.Error(t, err)
}

func TestFakeSecretGeneratorYieldsUniqueTokens(t *testing.T) {
	generator := memiam.NewFakeSecretGenerator()

	a, err := generator.NewSessionToken()
	require.NoError(t, err)
	b, err := generator.NewSessionToken()
	require.NoError(t, err)
	assert.NotEqual(t, a.String(), b.String())

	c, err := generator.NewAPITokenSecret()
	require.NoError(t, err)
	d, err := generator.NewAPITokenSecret()
	require.NoError(t, err)
	assert.NotEqual(t, c.String(), d.String())
}
