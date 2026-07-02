package catalog

import (
	"errors"
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

func TestLoadCatalog_ResolveError(t *testing.T) {
	_, err := loadCatalog(
		func(string) (string, error) { return "", errors.New("missing runfile") },
		func(string) ([]byte, error) { return nil, nil },
	)

	assert.ErrorContains(t, err, "locate")
}

func TestLoadCatalog_ReadError(t *testing.T) {
	_, err := loadCatalog(
		func(string) (string, error) { return "/some/path", nil },
		func(string) ([]byte, error) { return nil, errors.New("read failed") },
	)

	assert.ErrorContains(t, err, "read")
}

func TestActive_PanicsWhenTheCatalogCannotLoad(t *testing.T) {
	loaded = nil
	original := loader
	loader = func() (*Catalog, error) { return nil, errors.New("boom") }
	t.Cleanup(func() {
		loaded = nil
		loader = original
	})

	assert.Panics(t, func() { active() })
}
