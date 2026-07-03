package report_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/report"
	"github.com/usegavel/gavel/core/userinterface/cli/report/checks"
	"github.com/usegavel/gavel/core/userinterface/cli/report/github"
)

type fakePublisher struct {
	received checks.CheckRun
	result   github.Result
	err      error
}

func (fake *fakePublisher) Publish(_ context.Context, checkRun checks.CheckRun) (github.Result, error) {
	fake.received = checkRun
	return fake.result, fake.err
}

func runReport(workspace string, publisher report.ChecksPublisher, arguments ...string) (string, error) {
	command := report.NewCommand(
		func() (string, error) { return workspace, nil },
		func(github.Config) (report.ChecksPublisher, error) { return publisher, nil },
	)
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs(arguments)
	err := command.Execute()
	return output.String(), err
}

func writeVerdict(t *testing.T, workspace, project, content string) {
	t.Helper()
	dir := filepath.Join(workspace, ".gavel", "results", project)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(content), 0o644))
}

func TestNewCommandUsesReportVerb(t *testing.T) {
	command := report.NewCommand(nil, nil)
	assert.Equal(t, "report", command.Use)
}

func TestNewCommandRegistersEverySpecFlag(t *testing.T) {
	command := report.NewCommand(nil, nil)
	for _, name := range []string{
		"to", "github-token", "repo", "commit", "check-name", "new-only", "project",
	} {
		assert.NotNilf(t, command.Flags().Lookup(name), "flag --%s must be registered", name)
	}
}

func TestReportDeliversCachedVerdict(t *testing.T) {
	workspace := t.TempDir()
	writeVerdict(t, workspace, "core", `{"name":"core","verdict":"fail","commit_sha":"abc",
		"findings":[{"severity":"error","file_path":"a.go","line":3,"message":"boom","status":"new"}]}`)
	publisher := &fakePublisher{result: github.Result{URL: "https://github.com/o/r/runs/1"}}

	output, err := runReport(workspace, publisher, "--repo=o/r", "--github-token=tok")
	require.NoError(t, err)
	assert.Equal(t, checks.ConclusionFailure, publisher.received.Conclusion)
	require.Len(t, publisher.received.Annotations, 1)
	assert.Equal(t, "a.go", publisher.received.Annotations[0].Path)
	assert.Contains(t, output, "https://github.com/o/r/runs/1")
}

func TestReportErrorsWithoutCache(t *testing.T) {
	_, err := runReport(t.TempDir(), &fakePublisher{}, "--repo=o/r", "--github-token=tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gavel judge")
}

func TestReportRejectsUnsupportedSink(t *testing.T) {
	_, err := runReport(t.TempDir(), &fakePublisher{}, "--to=gitlab")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gitlab")
}

func TestNewGitHubPublisherBuildsPublisher(t *testing.T) {
	publisher, err := report.NewGitHubPublisher(github.Config{Token: "tok", Repo: "octo/repo"})
	require.NoError(t, err)
	assert.NotNil(t, publisher)
}

func TestNewGitHubPublisherRejectsBadConfig(t *testing.T) {
	publisher, err := report.NewGitHubPublisher(github.Config{Token: "tok", Repo: "noslash"})
	require.Error(t, err)
	assert.Nil(t, publisher)
}

func TestReportFiltersByProject(t *testing.T) {
	workspace := t.TempDir()
	writeVerdict(t, workspace, "core", `{"name":"core","verdict":"pass"}`)
	writeVerdict(t, workspace, "web", `{"name":"web","verdict":"fail"}`)
	publisher := &fakePublisher{result: github.Result{URL: "https://example.test/1"}}

	_, err := runReport(workspace, publisher, "--project=core", "--repo=o/r", "--github-token=tok")
	require.NoError(t, err)
	assert.Equal(t, checks.ConclusionSuccess, publisher.received.Conclusion)
}
