package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func TestNewServer_RegistersAllToolsAndResources(t *testing.T) {
	cli := executor.NewWithBinary("echo", "")

	assert.NotPanics(t, func() {
		srv := NewServer(cli, "test")
		assert.NotNil(t, srv)
	})
}

func TestNewCommand_ReturnsCommand(t *testing.T) {
	cmd := NewCommand("test")

	assert.Equal(t, "mcp", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Short, "MCP")
}

func TestImplementation_UsesProvidedVersion(t *testing.T) {
	impl := implementation("9.9.9")

	assert.Equal(t, "gavel", impl.Name)
	assert.Equal(t, "9.9.9", impl.Version)
}
