package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromRunfiles_ReadsPublishedCatalog(t *testing.T) {
	parsed, err := loadFromRunfiles()

	require.NoError(t, err)
	assert.NotEmpty(t, parsed.languages)
	assert.NotEmpty(t, parsed.aspectsBzl)
	assert.NotEmpty(t, parsed.languages["go"], "the published catalog must list go tools")
}

func TestActive_LoadsLazilyFromRunfiles(t *testing.T) {
	loaded = nil
	t.Cleanup(func() { loaded = nil })

	got := active()

	assert.NotEmpty(t, got.languages)
}
