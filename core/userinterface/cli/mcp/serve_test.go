package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func TestNewServer_RegistersAllToolsAndResources(t *testing.T) {
	cli := executor.NewWithBinary("echo", "")

	assert.NotPanics(t, func() {
		srv := NewServer(cli)
		assert.NotNil(t, srv)
	})
}

func TestNewCommand_ReturnsCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "mcp", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Short, "MCP")
}
