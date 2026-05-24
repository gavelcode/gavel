package executor_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func TestRun_EchoCommand(t *testing.T) {
	cli := executor.NewWithBinary("echo", "")

	out, code, err := cli.Run(context.Background(), "hello")

	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "hello\n", string(out))
}

func TestRun_FailingCommand(t *testing.T) {
	cli := executor.NewWithBinary("false", "")

	_, code, err := cli.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, code)
}

func TestRun_FailingCommandWithStderr_ReturnsError(t *testing.T) {
	cli := executor.NewWithBinary("bash", "")

	_, code, err := cli.Run(context.Background(), "-c", "echo 'server not configured' >&2; exit 1")

	require.Error(t, err)
	assert.Equal(t, 1, code)
	assert.Contains(t, err.Error(), "server not configured")
}

func TestRun_FailingCommandWithStdout_ReturnsOutput(t *testing.T) {
	cli := executor.NewWithBinary("bash", "")

	out, code, err := cli.Run(context.Background(), "-c", `echo '{"verdict":"fail"}'; exit 1`)

	require.NoError(t, err)
	assert.Equal(t, 1, code)
	assert.Contains(t, string(out), "verdict")
}

func TestRun_NonexistentBinary(t *testing.T) {
	cli := executor.NewWithBinary("nonexistent-binary-xyz", "")

	_, _, err := cli.Run(context.Background())

	assert.Error(t, err)
}

func TestRunJSON_AppendsFlag(t *testing.T) {
	cli := executor.NewWithBinary("echo", "")

	out, code, err := cli.RunJSON(context.Background(), "judge")

	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "judge --json\n", string(out))
}

func TestRunIn_UsesProvidedGavelspace(t *testing.T) {
	cli := executor.NewWithBinary("pwd", "/tmp")

	out, code, err := cli.RunIn(context.Background(), "/", )

	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "/", strings.TrimSpace(string(out)))
}

func TestRunIn_FallsBackToDefaultWorkspace(t *testing.T) {
	cli := executor.NewWithBinary("pwd", "/tmp")

	out, code, err := cli.RunIn(context.Background(), "")

	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Contains(t, strings.TrimSpace(string(out)), "/tmp")
}

func TestRunInJSON_AppendsFlag(t *testing.T) {
	cli := executor.NewWithBinary("echo", "")

	out, code, err := cli.RunInJSON(context.Background(), "", "judge")

	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "judge --json\n", string(out))
}

func TestExtractJSON_CleanInput(t *testing.T) {
	input := []byte(`{"projects":[]}`)

	result := executor.ExtractJSON(input)

	assert.Equal(t, input, result)
}

func TestExtractJSON_LeadingGarbage(t *testing.T) {
	input := []byte("\x1b[38;2;212;160;23m  ⚠ server unreachable\x1b[0m\n{\"projects\":[]}")

	result := executor.ExtractJSON(input)

	assert.Equal(t, []byte(`{"projects":[]}`), result)
}

func TestExtractJSON_LeadingGarbageArray(t *testing.T) {
	input := []byte("some warning text\n[{\"commit\":\"abc\"}]")

	result := executor.ExtractJSON(input)

	assert.Equal(t, []byte(`[{"commit":"abc"}]`), result)
}

func TestExtractJSON_NoJSON(t *testing.T) {
	input := []byte("no json here at all")

	result := executor.ExtractJSON(input)

	assert.Equal(t, input, result)
}

func TestExtractJSON_EmptyInput(t *testing.T) {
	result := executor.ExtractJSON([]byte{})

	assert.Empty(t, result)
}

func TestExtractJSON_ArrayWithDash(t *testing.T) {
	input := []byte("warning\n[-1, -2]")
	result := executor.ExtractJSON(input)
	assert.Equal(t, []byte("[-1, -2]"), result)
}

func TestExtractJSON_ArrayWithTrue(t *testing.T) {
	input := []byte("prefix\n[true]")
	result := executor.ExtractJSON(input)
	assert.Equal(t, []byte("[true]"), result)
}

func TestExtractJSON_ArrayWithFalse(t *testing.T) {
	input := []byte("prefix\n[false]")
	result := executor.ExtractJSON(input)
	assert.Equal(t, []byte("[false]"), result)
}

func TestExtractJSON_ArrayWithNull(t *testing.T) {
	input := []byte("prefix\n[null]")
	result := executor.ExtractJSON(input)
	assert.Equal(t, []byte("[null]"), result)
}

func TestExtractJSON_ArrayStartsWithBracket(t *testing.T) {
	result := executor.ExtractJSON([]byte(`[{"a":1}]`))
	assert.Equal(t, []byte(`[{"a":1}]`), result)
}

func TestNew_ReturnsNonNil(t *testing.T) {
	cli := executor.New("/tmp")
	require.NotNil(t, cli)
}

func TestRunJSON_ErrorPropagates(t *testing.T) {
	cli := executor.NewWithBinary("bash", "")

	_, code, err := cli.RunJSON(context.Background(), "-c", "echo 'fail' >&2; exit 1")

	require.Error(t, err)
	assert.Equal(t, 1, code)
}

func TestRunInJSON_ErrorPropagates(t *testing.T) {
	cli := executor.NewWithBinary("bash", "")

	_, code, err := cli.RunInJSON(context.Background(), "", "-c", "echo 'fail' >&2; exit 1")

	require.Error(t, err)
	assert.Equal(t, 1, code)
}
