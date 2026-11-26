package moduledeps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/moduledeps"
)

func workspaceWith(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return root
}

func TestDetect_SingleLineBazelDep(t *testing.T) {
	root := workspaceWith(t, map[string]string{"MODULE.bazel": `
module(name = "app")
bazel_dep(name = "rules_rs", version = "0.2.4")
bazel_dep(name = "rules_go", version = "0.50.0", repo_name = "io_bazel_rules_go")
`})

	deps := moduledeps.Detect(root)

	assert.True(t, deps["rules_rs"])
	assert.True(t, deps["rules_go"])
	assert.False(t, deps["app"], "module() is not a dependency")
}

func TestDetect_MultiLineBazelDep(t *testing.T) {
	root := workspaceWith(t, map[string]string{"MODULE.bazel": "bazel_dep(\n    name = \"rules_rs\",\n    version = \"0.2.4\",\n    dev_dependency = True,\n)\n"})

	assert.True(t, moduledeps.Detect(root)["rules_rs"])
}

func TestDetect_IgnoresComments(t *testing.T) {
	root := workspaceWith(t, map[string]string{"MODULE.bazel": `
# bazel_dep(name = "rules_rs", version = "1")
bazel_dep(name = "rules_rust", version = "1") # bazel_dep(name = "rules_rs", version = "1")
bazel_dep(
    # name = "rules_rs",
    name = "rules_java",
    version = "1",
)
`})

	deps := moduledeps.Detect(root)

	assert.False(t, deps["rules_rs"])
	assert.True(t, deps["rules_rust"])
	assert.True(t, deps["rules_java"])
}

func TestDetect_HashInsideStringIsNotAComment(t *testing.T) {
	root := workspaceWith(t, map[string]string{"MODULE.bazel": `bazel_dep(name = "rules_rs", version = "1#x")` + "\n"})

	assert.True(t, moduledeps.Detect(root)["rules_rs"])
}

func TestDetect_FollowsIncludedFiles(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"MODULE.bazel":           `include("//:deps/rust.MODULE.bazel")` + "\n",
		"deps/rust.MODULE.bazel": `bazel_dep(name = "rules_rs", version = "1")` + "\n" + `include("//:deps/rust.MODULE.bazel")` + "\n",
	})

	assert.True(t, moduledeps.Detect(root)["rules_rs"], "includes are followed, and include cycles terminate")
}

func TestDetect_MissingModuleFileYieldsNoDeps(t *testing.T) {
	assert.Empty(t, moduledeps.Detect(t.TempDir()))
}
