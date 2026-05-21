package git_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/git"
)

func TestCommitSHAReturnsValidHex(t *testing.T) {
	dir := initGitRepo(t)
	sc := git.NewSourceContextInDir(dir)

	sha, err := sc.CommitSHA(context.Background())

	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{40}$`), sha)
}

func TestBranchReturnsNonEmpty(t *testing.T) {
	dir := initGitRepo(t)
	sc := git.NewSourceContextInDir(dir)

	branch, err := sc.Branch(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, branch)
}

func TestCommitSHAReturnsErrorOnNonGitDir(t *testing.T) {
	skipIfNoGit(t)
	sc := git.NewSourceContextInDir(t.TempDir())

	_, err := sc.CommitSHA(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve commit SHA")
}

func TestBranchReturnsErrorOnNonGitDir(t *testing.T) {
	skipIfNoGit(t)
	sc := git.NewSourceContextInDir(t.TempDir())

	_, err := sc.Branch(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve branch")
}

func TestChangedLinesIntegration(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "checkout", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("line1\n"), 0644))
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "initial")

	gitRun(t, dir, "checkout", "-b", "feature")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("line1\nline2\nline3\n"), 0644))
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "add lines")

	sc := git.NewSourceContext()
	changed, err := sc.ChangedLines(context.Background(), dir, "main")

	require.NoError(t, err)
	assert.Contains(t, changed, "file.txt")
	assert.Equal(t, []int{2, 3}, changed["file.txt"])
}

func TestChangedLinesErrorOnInvalidBaseRef(t *testing.T) {
	dir := initGitRepo(t)

	sc := git.NewSourceContext()
	_, err := sc.ChangedLines(context.Background(), dir, "nonexistent-ref")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git merge-base")
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	skipIfNoGit(t)
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "checkout", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("initial\n"), 0644))
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "initial")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
}
