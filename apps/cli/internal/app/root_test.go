package app_test

import (
	"log/slog"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/apps/cli/internal/app"
)

func TestNewRootCommandHasExpectedSubcommands(t *testing.T) {
	deps := minimalDeps()
	root := app.NewRootCommand(deps)

	expected := []string{"judge", "init", "validate", "watch", "config", "projects", "trends", "mcp"}
	var names []string
	for _, cmd := range root.Commands() {
		names = append(names, cmd.Name())
	}
	assert.ElementsMatch(t, expected, names)
}

func TestNewRootCommandHasVerboseFlag(t *testing.T) {
	deps := minimalDeps()
	root := app.NewRootCommand(deps)

	flag := root.PersistentFlags().Lookup("verbose")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestNewRootCommandUsageAndShort(t *testing.T) {
	deps := minimalDeps()
	root := app.NewRootCommand(deps)

	assert.Equal(t, "gavel", root.Use)
	assert.NotEmpty(t, root.Short)
}

func TestVerboseFlagSetsDebugLogLevel(t *testing.T) {
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelWarn)

	deps := minimalDeps()
	deps.LogLevel = logLevel

	root := app.NewRootCommand(deps)
	noop := &cobra.Command{Use: "noop", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(noop)
	root.SetArgs([]string{"--verbose", "noop"})
	require.NoError(t, root.Execute())

	assert.Equal(t, slog.LevelDebug, logLevel.Level())
}

func minimalDeps() app.Deps {
	logLevel := new(slog.LevelVar)
	return app.Deps{
		WorkspaceResolver: func() (string, error) { return "/tmp", nil },
		Logger:            slog.Default(),
		LogLevel:          logLevel,
		Verifier:          stubVerifier{},
		ConfigInstaller:   stubInstaller{},
		ToolCatalog:       stubCatalog{},
	}
}

type stubVerifier struct{}

func (stubVerifier) VerifyStructure(string) ([]string, error) { return nil, nil }

type stubInstaller struct{}

func (stubInstaller) Install(string, []string) (map[string]bool, error) { return nil, nil }

type stubCatalog struct{}

func (stubCatalog) Catalog([]string) ([]string, []string) { return nil, nil }
