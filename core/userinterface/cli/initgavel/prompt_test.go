package initgavel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNotBlank_Empty(t *testing.T) {
	validate := validateNotBlank("project name")
	err := validate("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project name is required")
}

func TestValidateNotBlank_Whitespace(t *testing.T) {
	validate := validateNotBlank("field")
	err := validate("   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field is required")
}

func TestValidateNotBlank_Valid(t *testing.T) {
	validate := validateNotBlank("field")
	err := validate("some-value")
	require.NoError(t, err)
}
