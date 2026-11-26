package gavelconfig_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/gavelspace/gavelconfig"
)

func bridgesConfig(bridges string) []byte {
	return []byte(`
name: my-monorepo
projects:
  - name: payments
    pattern: //payments/...
    tooling:
      go: [golangci-lint]
    bridges:
` + bridges)
}

func TestParseShouldProduceProjectWithBridges(t *testing.T) {
	config, err := gavelconfig.Parse(bridgesConfig(`
      - name: shellcheck
        aspect: //tools/lint:shellcheck.bzl%shellcheck_gavel
        sarif_suffix: .shellcheck.sarif
        build_flags: ["--some=flag"]
      - name: todo
        aspect: //tools/lint:todo.bzl%todo_gavel
        sarif_suffix: .todo.sarif
`))

	require.NoError(t, err)
	bridges := config.Projects()[0].Bridges()
	require.Len(t, bridges, 2)
	assert.Equal(t, "shellcheck", bridges[0].Name())
	assert.Equal(t, "//tools/lint:shellcheck.bzl%shellcheck_gavel", bridges[0].Aspect())
	assert.Equal(t, ".shellcheck.sarif", bridges[0].SARIFSuffix())
	assert.Equal(t, []string{"--some=flag"}, bridges[0].BuildFlags())
	assert.Empty(t, bridges[1].BuildFlags())
}

func TestParseShouldRejectInvalidBridges(t *testing.T) {
	tests := []struct {
		name    string
		bridges string
		message string
	}{
		{
			name:    "shouldRejectMissingName",
			bridges: "      - {aspect: '//t:a.bzl%a', sarif_suffix: .a.sarif}",
			message: "name",
		},
		{
			name:    "shouldRejectMissingAspect",
			bridges: "      - {name: a, sarif_suffix: .a.sarif}",
			message: "aspect",
		},
		{
			name:    "shouldRejectMissingSuffix",
			bridges: "      - {name: a, aspect: '//t:a.bzl%a'}",
			message: "sarif_suffix",
		},
		{
			name:    "shouldRejectMalformedLabel",
			bridges: "      - {name: a, aspect: '//t:a.bzl', sarif_suffix: .a.sarif}",
			message: "%",
		},
		{
			name:    "shouldRejectSuffixNotEndingInSarif",
			bridges: "      - {name: a, aspect: '//t:a.bzl%a', sarif_suffix: .a.json}",
			message: ".sarif",
		},
		{
			name: "shouldRejectDuplicateNames",
			bridges: `      - {name: a, aspect: '//t:a.bzl%a', sarif_suffix: .a.sarif}
      - {name: a, aspect: '//t:b.bzl%b', sarif_suffix: .b.sarif}`,
			message: "duplicate",
		},
		{
			name:    "shouldRejectClashWithCatalogToolOfTheProject",
			bridges: "      - {name: golangci-lint, aspect: '//t:a.bzl%a', sarif_suffix: .a.sarif}",
			message: "golangci-lint",
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := gavelconfig.Parse(bridgesConfig(tcase.bridges))

			require.Error(t, err)
			assert.ErrorIs(t, err, gavelconfig.ErrParseConfig)
			assert.Contains(t, err.Error(), "bridges[")
			assert.Contains(t, err.Error(), tcase.message)
		})
	}
}

func TestParseShouldRejectUnknownBridgeField(t *testing.T) {
	_, err := gavelconfig.Parse(bridgesConfig("      - {name: a, aspect: '//t:a.bzl%a', sarif_suffix: .a.sarif, nope: 1}"))

	require.Error(t, err)
}
