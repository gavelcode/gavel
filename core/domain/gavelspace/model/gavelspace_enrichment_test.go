package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
)

func TestGavelspace_ServerConfig(t *testing.T) {
	gspace, err := model.NewGavelspace("myrepo")
	require.NoError(t, err)

	assert.False(t, gspace.ServerConfig().IsConfigured())

	gspace.SetServerConfig(model.NewServerConfig("https://gavel.example.com", "tok"))

	assert.True(t, gspace.ServerConfig().IsConfigured())
	assert.Equal(t, "https://gavel.example.com", gspace.ServerConfig().URL())
	assert.Equal(t, "tok", gspace.ServerConfig().Token())
}

func TestGavelspace_FindingsSource(t *testing.T) {
	gspace, err := model.NewGavelspace("myrepo")
	require.NoError(t, err)

	assert.Equal(t, "", gspace.FindingsSource())

	gspace.SetFindingsSource("rules_lint")

	assert.Equal(t, "rules_lint", gspace.FindingsSource())
}

func TestCoverageOptions(t *testing.T) {
	opts := model.NewCoverageOptions("small,medium", "-docker", "//core[/:]")

	assert.Equal(t, "small,medium", opts.TestSizeFilters())
	assert.Equal(t, "-docker", opts.TestTagFilters())
	assert.Equal(t, "//core[/:]", opts.InstrumentationFilter())
}

func TestServerConfig_Empty(t *testing.T) {
	sc := model.NewServerConfig("", "")

	assert.False(t, sc.IsConfigured())
	assert.Equal(t, "", sc.URL())
}
