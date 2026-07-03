package report_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/report"
)

func TestNewCommandUsesReportVerb(t *testing.T) {
	cmd := report.NewCommand()
	assert.Equal(t, "report", cmd.Use)
}

func TestNewCommandRegistersEverySpecFlag(t *testing.T) {
	cmd := report.NewCommand()
	for _, name := range []string{
		"to", "github-token", "repo", "commit", "check-name", "new-only", "project",
	} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "flag --%s must be registered", name)
	}
}

func TestNewCommandAppliesSpecDefaults(t *testing.T) {
	cmd := report.NewCommand()

	to := cmd.Flags().Lookup("to")
	require.NotNil(t, to)
	assert.Equal(t, "github-checks", to.DefValue)

	newOnly := cmd.Flags().Lookup("new-only")
	require.NotNil(t, newOnly)
	assert.Equal(t, "true", newOnly.DefValue)

	checkName := cmd.Flags().Lookup("check-name")
	require.NotNil(t, checkName)
	assert.Equal(t, "gavel", checkName.DefValue)
}

func TestRunEFailsUntilDeliveryIsImplemented(t *testing.T) {
	cmd := report.NewCommand()
	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
