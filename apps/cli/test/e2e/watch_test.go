//go:build e2e

package e2e_test

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatch_EmitsStartedEvent(t *testing.T) {
	t.Cleanup(func() { cleanWorkspace(t) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workspace := examplesGoRepo(t)

	cmd := exec.CommandContext(ctx, gavelBinary(t), "watch", "--workspace="+workspace)
	cmd.Dir = workspace
	cmd.Env = append(os.Environ(), "GAVEL_LOG_LEVEL=error")

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())

	scanner := bufio.NewScanner(stdout)
	require.True(t, scanner.Scan(), "expected at least one JSONL event on stdout")

	var event map[string]any
	require.NoError(t, json.Unmarshal(scanner.Bytes(), &event),
		"first line should be valid JSON, got: %s", scanner.Text())

	assert.Equal(t, "started", event["event"],
		"first event should be 'started', got: %v", event["event"])
	assert.NotEmpty(t, event["ts"], "started event should have a timestamp")
	assert.NotEmpty(t, event["workspace"], "started event should have workspace path")

	cancel()
	_ = cmd.Wait()
}

func TestWatch_EmitsChangedOnFileTouch(t *testing.T) {
	t.Cleanup(func() { cleanWorkspace(t) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workspace := examplesGoRepo(t)

	cmd := exec.CommandContext(ctx, gavelBinary(t), "watch", "--workspace="+workspace, "--debounce=200ms")
	cmd.Dir = workspace
	cmd.Env = append(os.Environ(), "GAVEL_LOG_LEVEL=error")

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())

	scanner := bufio.NewScanner(stdout)

	require.True(t, scanner.Scan(), "expected started event")
	var started map[string]any
	require.NoError(t, json.Unmarshal(scanner.Bytes(), &started))
	require.Equal(t, "started", started["event"])

	goFiles, err := filepath.Glob(filepath.Join(workspace, "internal", "domain", "order", "*.go"))
	require.NoError(t, err)
	require.NotEmpty(t, goFiles, "expected Go files in examples/go-repo/internal/domain/order/")

	target := goFiles[0]
	now := time.Now()
	require.NoError(t, os.Chtimes(target, now, now), "failed to touch file %s", target)

	deadline := time.After(15 * time.Second)
	var foundChanged bool
	for !foundChanged {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for 'changed' event after touching a file")
		default:
		}

		if !scanner.Scan() {
			break
		}
		var ev map[string]any
		if json.Unmarshal(scanner.Bytes(), &ev) != nil {
			continue
		}
		if ev["event"] == "changed" {
			foundChanged = true
			files, ok := ev["files"].([]any)
			assert.True(t, ok, "changed event should have a 'files' array")
			assert.NotEmpty(t, files, "changed event 'files' should not be empty")
		}
	}
	assert.True(t, foundChanged, "should have received a 'changed' event")

	cancel()
	_ = cmd.Wait()
}
