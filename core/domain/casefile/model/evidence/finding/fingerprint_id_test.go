package finding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func TestNewFingerprint(t *testing.T) {
	fp, err := finding.NewFingerprintID("abc123")

	require.NoError(t, err)
	assert.Equal(t, "abc123", fp.Value())
}

func TestNewFingerprintEmptyString(t *testing.T) {
	_, err := finding.NewFingerprintID("")

	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrInvalidFingerprintID)
}

func TestNewFingerprintWhitespaceOnly(t *testing.T) {
	_, err := finding.NewFingerprintID("   ")

	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrInvalidFingerprintID)
}

func TestFingerprintValue(t *testing.T) {
	fp, err := finding.NewFingerprintID("sha256:deadbeef")

	require.NoError(t, err)
	assert.Equal(t, "sha256:deadbeef", fp.Value())
}

func TestFingerprintEqualSameValue(t *testing.T) {
	fingerprintID, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)

	sameFingerprintID, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)

	assert.True(t, fingerprintID.Equal(sameFingerprintID))
}

func TestFingerprintEqualDifferentValue(t *testing.T) {
	fingerprintID, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)

	otherFingerprintID, err := finding.NewFingerprintID("xyz789")
	require.NoError(t, err)

	assert.False(t, fingerprintID.Equal(otherFingerprintID))
}
