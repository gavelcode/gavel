package catalog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedCatalog_IsBundledIntoTheBinary(t *testing.T) {
	assert.NotEmpty(t, embeddedCatalog, "catalog.yaml must be embedded so the standalone binary carries it")
}

func TestLoadEmbedded_ParsesTheBundledCatalog(t *testing.T) {
	parsed, err := loadEmbedded()

	require.NoError(t, err)
	assert.NotEmpty(t, parsed.languages)
	assert.NotEmpty(t, parsed.aspectsBzl)
	assert.NotEmpty(t, parsed.languages["go"], "the bundled catalog must list go tools")
	assert.NotEmpty(t, parsed.languages["rust"], "the bundled catalog must list rust tools")
}

func TestActive_LoadsLazily(t *testing.T) {
	previous := loaded
	loaded = nil
	t.Cleanup(func() { loaded = previous })

	got := active()

	assert.NotEmpty(t, got.languages)
}

func TestActive_PanicsWhenTheCatalogCannotLoad(t *testing.T) {
	previous := loaded
	loaded = nil
	original := loader
	loader = func() (*Catalog, error) { return nil, errors.New("boom") }
	t.Cleanup(func() {
		loaded = previous
		loader = original
	})

	assert.Panics(t, func() { active() })
}
