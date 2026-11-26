package collector_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

var _ collectevidence.BridgeCollector = (*collector.BazelBridgeCollector)(nil)

const validBridgeDoc = `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"todo"}},"results":[
{"ruleId":"todo:found","message":{"text":"TODO"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"app/a.sh"},"region":{"startLine":2}}}]}]}]}`

type fakeBridgeRunner struct {
	results map[string]*runner.BridgeResult
	errs    map[string]error
	configs []runner.BridgeConfig
}

func (f *fakeBridgeRunner) RunBridge(_ context.Context, config runner.BridgeConfig) (*runner.BridgeResult, error) {
	f.configs = append(f.configs, config)
	if err := f.errs[config.Aspect]; err != nil {
		return nil, err
	}
	return f.results[config.Aspect], nil
}

type recordingFindingsParser struct {
	commands []ingestfind.Command
}

func (p *recordingFindingsParser) Execute(_ context.Context, cmd ingestfind.Command) (ingestfind.Result, error) {
	p.commands = append(p.commands, cmd)
	return ingestfind.Result{Evidence: evidencedto.Evidence{
		Subtype:  "code_quality",
		Source:   cmd.Source(),
		Findings: []evidencedto.Finding{{Tool: cmd.Source(), RuleID: "todo:found", Provenance: cmd.Provenance().String()}},
	}}, nil
}

func spec(name string) collectevidence.BridgeSpec {
	return collectevidence.BridgeSpec{Name: name, Aspect: "//tools:" + name + ".bzl%" + name, SARIFSuffix: "." + name + ".sarif", BuildFlags: []string{"--f"}}
}

func files(docs ...string) *runner.BridgeResult {
	result := &runner.BridgeResult{}
	for index, doc := range docs {
		result.Files = append(result.Files, runner.BridgeFile{Name: fmt.Sprintf("pkg%d/x.sarif", index), Data: []byte(doc)})
	}
	return result
}

func collect(t *testing.T, r *fakeBridgeRunner, p *recordingFindingsParser, specs ...collectevidence.BridgeSpec) collectevidence.BridgeResult {
	t.Helper()
	c := collector.NewBazelBridgeCollector(r, p)
	result, err := c.CollectBridges(context.Background(), "/ws", []string{"//app/..."}, specs, []string{"ci"})
	require.NoError(t, err)
	return result
}

func failuresOf(result collectevidence.BridgeResult) []evidencedto.ToolFailure {
	return evidencedto.ExtractToolFailures(result.Evidences)
}

func TestBridgeCollectorRunsEachBridgeOverTheProjectTargets(t *testing.T) {
	bridgeRunner := &fakeBridgeRunner{results: map[string]*runner.BridgeResult{
		"//tools:todo.bzl%todo": files(validBridgeDoc), "//tools:other.bzl%other": files(validBridgeDoc),
	}}

	collect(t, bridgeRunner, &recordingFindingsParser{}, spec("todo"), spec("other"))

	require.Len(t, bridgeRunner.configs, 2)
	assert.Equal(t, runner.BridgeConfig{
		Workspace: "/ws", Targets: []string{"//app/..."}, Aspect: "//tools:todo.bzl%todo",
		SARIFSuffix: ".todo.sarif", BuildFlags: []string{"--f"}, BazelConfigs: []string{"ci"},
	}, bridgeRunner.configs[0])
}

func TestBridgeCollectorMarksFindingsAsBridgeAndNamesTheToolAfterTheBridge(t *testing.T) {
	bridgeRunner := &fakeBridgeRunner{results: map[string]*runner.BridgeResult{"//tools:todo.bzl%todo": files(validBridgeDoc)}}
	parser := &recordingFindingsParser{}

	result := collect(t, bridgeRunner, parser, spec("todo"))

	require.Len(t, parser.commands, 1)
	assert.Equal(t, "todo", parser.commands[0].Source(), "the bridge name, not the SARIF driver name, is the tool")
	assert.True(t, parser.commands[0].Provenance().IsBridge())
	require.Len(t, result.Evidences, 1)
	assert.Equal(t, "bridge", result.Evidences[0].Findings[0].Provenance)
	assert.Empty(t, failuresOf(result))
	require.Len(t, result.RawFiles, 1)
	assert.Equal(t, "bridge", result.RawFiles[0].Provenance)
	assert.Contains(t, string(result.RawFiles[0].Data), `"provenance":"bridge"`, "emitted SARIF carries the marker in the run property bag")
	assert.Contains(t, string(result.RawFiles[0].Data), `"bridge":"todo"`)
}

func TestBridgeCollectorReportsBuildFailureAsFailedBridgeExecution(t *testing.T) {
	r := &fakeBridgeRunner{errs: map[string]error{"//tools:todo.bzl%todo": errors.New("bazel build failed: boom")}}

	result := collect(t, r, &recordingFindingsParser{}, spec("todo"))

	failures := failuresOf(result)
	require.Len(t, failures, 1)
	assert.Equal(t, "todo", failures[0].Tool)
	assert.Equal(t, "bridge", failures[0].Provenance)
	assert.False(t, failures[0].Degraded)
	assert.Contains(t, failures[0].Reason, "boom")
	assert.Empty(t, result.RawFiles)
}

func TestBridgeCollectorFailsTheWholeBridgeOnContractViolationAndListsFirstOffendingFiles(t *testing.T) {
	bad := `{"runs":[{"tool":{"driver":{"name":"todo"}},"results":[{"ruleId":"","message":{"text":"m"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"a.sh"},"region":{"startLine":1}}}]}]}]}`
	r := &fakeBridgeRunner{results: map[string]*runner.BridgeResult{"//tools:todo.bzl%todo": files(validBridgeDoc, bad, "{nope", bad, bad)}}
	p := &recordingFindingsParser{}

	result := collect(t, r, p, spec("todo"))

	failures := failuresOf(result)
	require.Len(t, failures, 1)
	assert.Contains(t, failures[0].Reason, "SARIF contract violated in 4 of 5 files")
	assert.Contains(t, failures[0].Reason, "pkg1/x.sarif: runs[0]: results[0]: ruleId must be non-empty")
	assert.Contains(t, failures[0].Reason, "pkg2/x.sarif: not valid SARIF JSON")
	assert.Contains(t, failures[0].Reason, "pkg3/x.sarif")
	assert.NotContains(t, failures[0].Reason, "pkg4/x.sarif", "only the first offending files are listed")
	assert.Contains(t, failures[0].Reason, "and 1 more")
	assert.Empty(t, p.commands, "no finding of a contract-breaking bridge is ingested, even from its valid files")
	assert.Empty(t, result.RawFiles)
}

func TestBridgeCollectorSurfacesUnsuccessfulInvocationAsFailureWithItsReason(t *testing.T) {
	doc := `{"runs":[{"tool":{"driver":{"name":"todo"}},"results":[],"invocations":[{"executionSuccessful":false,"toolExecutionNotifications":[{"message":{"text":"tool crashed on a.sh"}}]}]}]}`
	r := &fakeBridgeRunner{results: map[string]*runner.BridgeResult{"//tools:todo.bzl%todo": files(doc)}}

	result := collect(t, r, &recordingFindingsParser{}, spec("todo"))

	failures := failuresOf(result)
	require.Len(t, failures, 1)
	assert.Equal(t, "todo", failures[0].Tool, "named after the bridge, not the SARIF driver")
	assert.Equal(t, "bridge", failures[0].Provenance)
	assert.Contains(t, failures[0].Reason, "tool crashed on a.sh")
}

func TestBridgeCollectorKeepsBridgesIndependent(t *testing.T) {
	r := &fakeBridgeRunner{
		results: map[string]*runner.BridgeResult{"//tools:other.bzl%other": files(validBridgeDoc)},
		errs:    map[string]error{"//tools:todo.bzl%todo": errors.New("boom")},
	}

	result := collect(t, r, &recordingFindingsParser{}, spec("todo"), spec("other"))

	assert.Len(t, failuresOf(result), 1)
	assert.Len(t, evidencedto.ExtractFindings(result.Evidences), 1)
}

func TestBridgeCollectorFailureEvidenceIsAToolExecutionThatConvertsToDomain(t *testing.T) {
	r := &fakeBridgeRunner{errs: map[string]error{"//tools:todo.bzl%todo": errors.New("boom")}}

	result := collect(t, r, &recordingFindingsParser{}, spec("todo"))

	require.Len(t, result.Evidences, 1)
	assert.Equal(t, evidence.SubtypeToolExecution.String(), result.Evidences[0].Subtype)
	_, err := evidencedto.EvidenceToDomain(result.Evidences[0])
	require.NoError(t, err)
}

func TestBridgeCollectorWithoutBridgesDoesNothing(t *testing.T) {
	r := &fakeBridgeRunner{}

	result := collect(t, r, &recordingFindingsParser{})

	assert.Empty(t, result.Evidences)
	assert.Empty(t, r.configs)
}
