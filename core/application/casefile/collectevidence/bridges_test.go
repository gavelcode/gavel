package collectevidence_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

type fakeBridgeCollector struct {
	result collectevidence.BridgeResult
	err    error

	called    bool
	workspace string
	targets   []string
	bridges   []collectevidence.BridgeSpec
	configs   []string
}

func (f *fakeBridgeCollector) CollectBridges(_ context.Context, workspace string, targets []string, bridges []collectevidence.BridgeSpec, bazelConfigs []string) (collectevidence.BridgeResult, error) {
	f.called = true
	f.workspace, f.targets, f.bridges, f.configs = workspace, targets, bridges, bazelConfigs
	return f.result, f.err
}

type countingToolExecParser struct{ calls int }

func (p *countingToolExecParser) ParseToolExecutions(_ []byte) ([]evidencedto.ToolFailure, error) {
	p.calls++
	return nil, nil
}

func bridgeSpec() collectevidence.BridgeSpec {
	return collectevidence.BridgeSpec{Name: "todo", Aspect: "//tools:todo.bzl%todo", SARIFSuffix: ".todo.sarif"}
}

func bridgeHandler(bridges collectevidence.BridgeCollector, opts ...collectevidence.HandlerOption) *collectevidence.Handler {
	return newHandlerWith(&fakeFindingsCollector{}, append([]collectevidence.HandlerOption{collectevidence.WithBridgeCollector(bridges)}, opts...)...)
}

func TestExecute_RunsBridgesOverTheProjectTargetsAndCountsTheirFindings(t *testing.T) {
	bridges := &fakeBridgeCollector{result: collectevidence.BridgeResult{
		Evidences: []evidencedto.Evidence{{Subtype: "code_quality", Source: "todo", Findings: []evidencedto.Finding{
			{Tool: "todo", RuleID: "todo:found", FingerprintID: "bridge/todo/fp", Provenance: "bridge"},
		}}},
		RawFiles: []collectevidence.RawFile{{Format: "sarif", Source: "todo", Data: []byte("{}"), Provenance: "bridge"}},
	}}
	cmd, err := collectevidence.NewCommand("/ws", "//app/...", "app", "main", nil, true, false, nil,
		collectevidence.WithBridges([]collectevidence.BridgeSpec{bridgeSpec()}),
		collectevidence.WithScopedTargets([]string{"//app:a"}),
		collectevidence.WithBazelConfigs([]string{"ci"}))
	require.NoError(t, err)

	result, err := bridgeHandler(bridges).Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, "/ws", bridges.workspace)
	assert.Equal(t, []string{"//app:a"}, bridges.targets)
	assert.Equal(t, []collectevidence.BridgeSpec{bridgeSpec()}, bridges.bridges)
	assert.Equal(t, []string{"ci"}, bridges.configs)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Equal(t, []string{"bridge/todo/fp"}, result.Fingerprints)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, "bridge", result.Findings[0].Provenance)
	require.Len(t, result.RawSARIF, 1)
	assert.Equal(t, "bridge", result.RawSARIF[0].Provenance)
	assert.Len(t, result.SARIFDocs, 1)
}

func TestExecute_KeepsBridgeFailuresAsToolExecutionEvidence(t *testing.T) {
	failure := evidencedto.Evidence{Subtype: "tool_execution", Source: "todo", ToolExecution: &evidencedto.ToolExecution{
		Failures: []evidencedto.ToolFailure{{Tool: "todo", Reason: "boom", Provenance: "bridge"}},
	}}
	bridges := &fakeBridgeCollector{result: collectevidence.BridgeResult{Evidences: []evidencedto.Evidence{failure}}}
	cmd, err := collectevidence.NewCommand("/ws", "//app/...", "app", "main", nil, true, false, nil,
		collectevidence.WithBridges([]collectevidence.BridgeSpec{bridgeSpec()}))
	require.NoError(t, err)

	result, err := bridgeHandler(bridges).Execute(context.Background(), cmd)

	require.NoError(t, err)
	failures := evidencedto.ExtractToolFailures(result.Evidences)
	require.Len(t, failures, 1)
	assert.Equal(t, "bridge", failures[0].Provenance)
}

func TestExecute_DoesNotReparseBridgeSARIFForToolExecutions(t *testing.T) {
	parser := &countingToolExecParser{}
	bridges := &fakeBridgeCollector{result: collectevidence.BridgeResult{
		RawFiles: []collectevidence.RawFile{{Format: "sarif", Source: "todo", Data: []byte("{}"), Provenance: "bridge"}},
	}}
	cmd, err := collectevidence.NewCommand("/ws", "//app/...", "app", "main", nil, true, false, nil,
		collectevidence.WithBridges([]collectevidence.BridgeSpec{bridgeSpec()}))
	require.NoError(t, err)

	_, err = bridgeHandler(bridges, collectevidence.WithToolExecutionParser(parser)).Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Zero(t, parser.calls, "the bridge collector already turned its SARIF failures into bridge-marked tool executions")
}

func TestExecute_SkipsBridgeCollectorWhenProjectDeclaresNone(t *testing.T) {
	bridges := &fakeBridgeCollector{}
	cmd, err := collectevidence.NewCommand("/ws", "//app/...", "app", "main", nil, true, false, nil)
	require.NoError(t, err)

	_, err = bridgeHandler(bridges).Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.False(t, bridges.called)
}

func TestCommandBridgesAreDefensivelyCopied(t *testing.T) {
	const tampered = "tampered"
	specs := []collectevidence.BridgeSpec{bridgeSpec()}
	cmd, err := collectevidence.NewCommand("/ws", "//app/...", "app", "main", nil, true, false, nil, collectevidence.WithBridges(specs))
	require.NoError(t, err)

	specs[0].Name = tampered
	cmd.Bridges()[0].Name = tampered

	assert.Equal(t, "todo", cmd.Bridges()[0].Name)
}
