package mcp_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"

	gavelmcp "github.com/usegavel/gavel/core/userinterface/cli/mcp"
)

func fakeGavelPath(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script fakes not supported on windows")
	}
	path := filepath.Join(testdataDir(t), "fake-gavel.sh")
	info, err := os.Stat(path)
	require.NoError(t, err, "fake-gavel.sh not found in testdata")
	require.True(t, info.Mode()&0o111 != 0, "fake-gavel.sh must be executable")
	return path
}

func testdataDir(t *testing.T) string {
	t.Helper()
	if srcDir := os.Getenv("TEST_SRCDIR"); srcDir != "" {
		workspace := os.Getenv("TEST_WORKSPACE")
		if workspace == "" {
			workspace = "_main"
		}
		p := filepath.Join(srcDir, workspace, "apps/cli/test/e2e/mcp/testdata")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		candidate := filepath.Join(dir, "apps/cli/test/e2e/mcp/testdata")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find testdata directory")
		}
		dir = parent
	}
}

func connectServerAndClient(t *testing.T, cli *executor.CLI) *mcpsdk.ClientSession {
	t.Helper()
	ctx := context.Background()

	server := gavelmcp.NewServer(cli)
	cTransport, sTransport := mcpsdk.NewInMemoryTransports()

	ss, err := server.Connect(ctx, sTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ss.Close()) })

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	cs, err := client.Connect(ctx, cTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cs.Close()) })

	return cs
}

func TestListTools(t *testing.T) {
	cli := executor.NewWithBinary(fakeGavelPath(t), t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.ListTools(context.Background(), &mcpsdk.ListToolsParams{})
	require.NoError(t, err)

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	assert.Contains(t, names, "gavel_judge")
	assert.Contains(t, names, "gavel_lint_file")
	assert.Contains(t, names, "gavel_coverage")
	assert.Contains(t, names, "gavel_validate")
	assert.Contains(t, names, "gavel_init")
	assert.Contains(t, names, "gavel_trends")
	assert.Contains(t, names, "gavel_arch")
	assert.Contains(t, names, "gavel_findings")
	assert.Len(t, result.Tools, 8)
}

func TestListResources(t *testing.T) {
	cli := executor.NewWithBinary(fakeGavelPath(t), t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.ListResources(context.Background(), &mcpsdk.ListResourcesParams{})
	require.NoError(t, err)

	uris := make([]string, 0, len(result.Resources))
	for _, r := range result.Resources {
		uris = append(uris, r.URI)
	}
	assert.Contains(t, uris, "gavel://config")
	assert.Contains(t, uris, "gavel://projects")
	assert.Contains(t, uris, "gavel://architecture")
}

func TestCallTool_Judge(t *testing.T) {
	cli := executor.NewWithBinary(fakeGavelPath(t), t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "gavel_judge",
		Arguments: map[string]any{"quick": true},
	})
	require.NoError(t, err)
	require.False(t, result.IsError, "tool call should not be an error")
	require.NotEmpty(t, result.Content)

	text := result.Content[0].(*mcpsdk.TextContent).Text
	assert.Contains(t, text, "core")
	assert.Contains(t, text, "PASS")
	assert.Contains(t, text, "Findings: 3")
	assert.Contains(t, text, "Coverage: 92.5%")
}

func TestCallTool_Validate(t *testing.T) {
	cli := executor.NewWithBinary(fakeGavelPath(t), t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "gavel_validate",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.NotEmpty(t, result.Content)

	text := result.Content[0].(*mcpsdk.TextContent).Text
	assert.Contains(t, text, "valid")
}

func TestReadResource_Config(t *testing.T) {
	cli := executor.NewWithBinary(fakeGavelPath(t), t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{
		URI: "gavel://config",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Contents)

	text := result.Contents[0].Text
	assert.Contains(t, text, "Gavel Configuration")
	assert.Contains(t, text, "core")
	assert.Contains(t, text, "code_quality")
	assert.Contains(t, text, "coverage")
}

func TestReadResource_Projects(t *testing.T) {
	cli := executor.NewWithBinary(fakeGavelPath(t), t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{
		URI: "gavel://projects",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Contents)

	text := result.Contents[0].Text
	assert.Contains(t, text, "core")
}

func TestCallTool_Judge_ExecutorError(t *testing.T) {
	cli := executor.NewWithBinary("/nonexistent/binary/gavel", t.TempDir())
	session := connectServerAndClient(t, cli)

	result, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "gavel_judge",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError, "tool call should report an error when executor fails")
}
