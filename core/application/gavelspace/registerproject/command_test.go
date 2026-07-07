package registerproject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/gavelspace/registerproject"
)

func TestNewCommandRejectsEmptyGavelspaceName(t *testing.T) {
	_, err := registerproject.NewCommand(testTenant, "", "proj-1", "//svc/...")
	require.Error(t, err)
	assert.ErrorIs(t, err, registerproject.ErrInvalidCommand)
}

func TestNewCommandRejectsEmptyProjectID(t *testing.T) {
	_, err := registerproject.NewCommand(testTenant, "alpha", "", "//svc/...")
	require.Error(t, err)
	assert.ErrorIs(t, err, registerproject.ErrInvalidCommand)
}

func TestNewCommandRejectsEmptyTargetPattern(t *testing.T) {
	_, err := registerproject.NewCommand(testTenant, "alpha", "proj-1", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, registerproject.ErrInvalidCommand)
}

func TestNewCommandRejectsEmptyTenant(t *testing.T) {
	_, err := registerproject.NewCommand("", "alpha", "proj-1", "//svc/...")
	require.Error(t, err)
	assert.ErrorIs(t, err, registerproject.ErrInvalidCommand)
}
