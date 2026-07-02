package installer

import (
	"os"
	"regexp"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGavelToolsVersionMatchesModuleBazel(t *testing.T) {
	path, err := runfiles.Rlocation("_main/MODULE.bazel")
	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	match := regexp.MustCompile(`bazel_dep\(name = "gavel_tools", version = "([^"]+)"\)`).FindStringSubmatch(string(data))
	require.NotNil(t, match, "MODULE.bazel must declare a gavel_tools bazel_dep")

	assert.Equal(t, match[1], gavelToolsVersion,
		"gavelToolsVersion must match the gavel_tools version gavel itself depends on in MODULE.bazel")
}
