//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
)

const (
	adminEmail    = "admin@gavel.local"
	adminPassword = testkit.SeedAdminPassword
	newPassword   = "e2e-test-password-123!"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); dir != "" {
		return dir
	}
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "MODULE.bazel")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}

func serverBinary(t *testing.T) string {
	t.Helper()
	if bin := os.Getenv("GAVEL_SERVER_BINARY"); bin != "" {
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
		t.Skipf("GAVEL_SERVER_BINARY not found: %s", bin)
	}
	root := projectRoot(t)
	out, err := exec.Command("bazel", "cquery", "//apps/server/cmd/gavel-server", "--output=files").Output()
	if err == nil {
		bin := filepath.Join(root, strings.TrimSpace(string(out)))
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
	}
	t.Skip("gavel-server binary not found")
	return ""
}

func gavelBinary(t *testing.T) string {
	t.Helper()
	if bin := os.Getenv("GAVEL_BINARY"); bin != "" {
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
		t.Skipf("GAVEL_BINARY not found: %s", bin)
	}
	root := projectRoot(t)
	out, err := exec.Command("bazel", "cquery", "//apps/cli/cmd/gavel", "--output=files").Output()
	if err == nil {
		bin := filepath.Join(root, strings.TrimSpace(string(out)))
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
	}
	t.Skip("gavel binary not found")
	return ""
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

type serverInstance struct {
	URL  string
	cmd  *exec.Cmd
	port int
}

func startRealServer(t *testing.T) *serverInstance {
	t.Helper()

	dsn := testkit.TestDSN(t)
	_ = testkit.TestDB(t)

	bin := serverBinary(t)
	port := freePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	cmd := exec.Command(bin, "serve")
	cmd.Env = []string{
		fmt.Sprintf("GAVEL_DATABASE_URL=%s", dsn),
		fmt.Sprintf("GAVEL_ADDR=%s", addr),
		"GAVEL_SECURE_COOKIES=false",
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Start(), "start gavel-server")

	t.Cleanup(func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			t.Fatal("server did not become ready within 30s")
		}
		resp, err := http.Get(url + "/api/v1/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	return &serverInstance{URL: url, cmd: cmd, port: port}
}

func issueToken(t *testing.T, serverURL string) string {
	t.Helper()
	client := &http.Client{Jar: nil}

	cookie := login(t, client, serverURL, adminEmail, adminPassword)

	changePwBody := fmt.Sprintf(`{"current_password":"%s","new_password":"%s"}`, adminPassword, newPassword)
	req, _ := http.NewRequest("POST", serverURL+"/api/v1/me/password", bytes.NewBufferString(changePwBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	cookie = login(t, client, serverURL, adminEmail, newPassword)

	tokenReq, _ := http.NewRequest("POST", serverURL+"/api/v1/me/tokens",
		bytes.NewBufferString(`{"name":"e2e","scopes":["ingest","project_sync"]}`))
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenReq.AddCookie(cookie)
	resp3, err := client.Do(tokenReq)
	require.NoError(t, err)
	defer resp3.Body.Close()
	require.Equal(t, 201, resp3.StatusCode, "token creation should succeed")

	var tokenResult struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&tokenResult))
	require.NotEmpty(t, tokenResult.Token)
	return tokenResult.Token
}

func login(t *testing.T, client *http.Client, serverURL, email, password string) *http.Cookie {
	t.Helper()
	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	resp, err := client.Post(serverURL+"/api/v1/sessions", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode, "login should succeed for %s", email)

	for _, c := range resp.Cookies() {
		if c.Name == "gavel_session" {
			return c
		}
	}
	t.Fatal("no session cookie returned")
	return nil
}

func runGavel(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, gavelBinary(t), args...)
	cmd.Dir = dir
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "BUILD_WORKSPACE_DIRECTORY=") {
			env = append(env, e)
		}
	}
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestServer_HealthEndpoint(t *testing.T) {
	srv := startRealServer(t)

	resp, err := http.Get(srv.URL + "/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestServer_AdminBootstrap(t *testing.T) {
	srv := startRealServer(t)
	token := issueToken(t, srv.URL)
	assert.True(t, strings.HasPrefix(token, "gav_"), "token should have gav_ prefix")
}

func TestServer_CLIJudgeSubmitsToServer(t *testing.T) {
	srv := startRealServer(t)
	token := issueToken(t, srv.URL)

	root := projectRoot(t)
	workspace := filepath.Join(root, "examples", "go-repo")

	stdout, stderr, exitCode := runGavel(t, workspace,
		"judge", "--project=api-gateway",
		"--server="+srv.URL, "--token="+token)

	assert.True(t, exitCode == 0 || exitCode == 1,
		"judge should produce a verdict; stdout: %s\nstderr: %s", stdout, stderr)
	assert.NotContains(t, stderr, "server submission failed")
}

func TestServer_TrendsAfterJudge(t *testing.T) {
	srv := startRealServer(t)
	token := issueToken(t, srv.URL)

	root := projectRoot(t)
	workspace := filepath.Join(root, "examples", "go-repo")

	_, _, exitCode := runGavel(t, workspace,
		"judge", "--project=api-gateway",
		"--server="+srv.URL, "--token="+token)
	require.True(t, exitCode == 0 || exitCode == 1)

	stdout, _, trendsExit := runGavel(t, workspace,
		"trends", "--project=api-gateway",
		"--server="+srv.URL, "--token="+token, "--json")
	require.Equal(t, 0, trendsExit, "trends should succeed")

	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &entries))
	require.NotEmpty(t, entries, "should have at least 1 history entry")
	assert.Contains(t, entries[0], "verdict_outcome")
	assert.Contains(t, entries[0], "total_findings")
}

func TestServer_BaselineTracking(t *testing.T) {
	srv := startRealServer(t)
	token := issueToken(t, srv.URL)

	root := projectRoot(t)
	workspace := filepath.Join(root, "examples", "go-repo")

	_, _, exit1 := runGavel(t, workspace,
		"judge", "--project=api-gateway", "--commit=aaa111",
		"--server="+srv.URL, "--token="+token)
	require.True(t, exit1 == 0 || exit1 == 1)

	_, _, exit2 := runGavel(t, workspace,
		"judge", "--project=api-gateway", "--commit=bbb222",
		"--server="+srv.URL, "--token="+token)
	require.True(t, exit2 == 0 || exit2 == 1)

	stdout, _, trendsExit := runGavel(t, workspace,
		"trends", "--project=api-gateway",
		"--server="+srv.URL, "--token="+token, "--json")
	require.Equal(t, 0, trendsExit)

	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &entries))
	assert.GreaterOrEqual(t, len(entries), 2, "should have 2 history entries after 2 judges")
}
