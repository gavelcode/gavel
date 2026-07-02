package collectevidence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
)

func TestNewCommand_Valid(t *testing.T) {
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)

	require.NoError(t, err)
	assert.Equal(t, "/ws", cmd.Workspace())
	assert.Equal(t, "//core/...", cmd.TargetPattern())
	assert.Equal(t, "core", cmd.ProjectName())
	assert.Equal(t, "main", cmd.DefaultBranch())
	assert.Equal(t, []string{"go"}, cmd.Languages())
	assert.False(t, cmd.Quick())
	assert.False(t, cmd.Absolute())
}

func TestNewCommand_EmptyWorkspace(t *testing.T) {
	_, err := collectevidence.NewCommand("", "//core/...", "core", "main", nil, false, false, nil)

	assert.ErrorIs(t, err, collectevidence.ErrInvalidCommand)
}

func TestNewCommand_EmptyPattern(t *testing.T) {
	_, err := collectevidence.NewCommand("/ws", "", "core", "main", nil, false, false, nil)

	assert.ErrorIs(t, err, collectevidence.ErrInvalidCommand)
}

func TestNewCommand_QuickMode(t *testing.T) {
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, true, false, nil)

	require.NoError(t, err)
	assert.True(t, cmd.Quick())
}

func TestNewCommand_WithScopedTargets(t *testing.T) {
	targets := []string{"//core/domain:model", "//core/application:app"}
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil,
		collectevidence.WithScopedTargets(targets),
	)

	require.NoError(t, err)
	assert.Equal(t, targets, cmd.ScopedTargets())
	assert.Equal(t, "//core/...", cmd.TargetPattern())
}

func TestNewCommand_WithToolSelection(t *testing.T) {
	selection := map[string][]string{"go": {"golangci-lint", "archtest"}}
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil,
		collectevidence.WithToolSelection(selection),
	)

	require.NoError(t, err)
	assert.Equal(t, selection, cmd.ToolSelection())
}

func TestNewCommand_ToolSelectionIsDefensivelyCopied(t *testing.T) {
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil,
		collectevidence.WithToolSelection(map[string][]string{"go": {"golangci-lint"}}),
	)
	require.NoError(t, err)

	mutated := cmd.ToolSelection()
	mutated["go"][0] = "tampered"

	assert.Equal(t, map[string][]string{"go": {"golangci-lint"}}, cmd.ToolSelection())
}

func TestNewCommand_WithoutScopedTargets_ReturnsNil(t *testing.T) {
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)

	require.NoError(t, err)
	assert.Nil(t, cmd.ScopedTargets())
}
