package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	coresarif "github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

// maxListedOffenders bounds how many offending files a contract failure names.
const maxListedOffenders = 3

type BridgeRunner interface {
	RunBridge(ctx context.Context, config runner.BridgeConfig) (*runner.BridgeResult, error)
}

// BazelBridgeCollector builds the project's bridges and judges what they
// produced. A bridge is the project's own aspect, so every way it can fail
// (build, missing output, contract) becomes a failed tool execution rather
// than silence or a hard error that would hide the other bridges.
type BazelBridgeCollector struct {
	runner BridgeRunner
	parser FindingsParser
}

func NewBazelBridgeCollector(r BridgeRunner, p FindingsParser) *BazelBridgeCollector {
	return &BazelBridgeCollector{runner: r, parser: p}
}

func (c *BazelBridgeCollector) CollectBridges(ctx context.Context, workspace string, targets []string, bridges []collectevidence.BridgeSpec, bazelConfigs []string) (collectevidence.BridgeResult, error) {
	var result collectevidence.BridgeResult
	for _, bridge := range bridges {
		built, err := c.runner.RunBridge(ctx, runner.BridgeConfig{
			Workspace:    workspace,
			Targets:      targets,
			Aspect:       bridge.Aspect,
			SARIFSuffix:  bridge.SARIFSuffix,
			BuildFlags:   bridge.BuildFlags,
			BazelConfigs: bazelConfigs,
		})
		if err != nil {
			result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, err.Error(), false))
			continue
		}
		c.judge(ctx, bridge, built.Files, &result)
	}
	return result, nil
}

func (c *BazelBridgeCollector) judge(ctx context.Context, bridge collectevidence.BridgeSpec, files []runner.BridgeFile, result *collectevidence.BridgeResult) {
	if reason := contractViolations(files); reason != "" {
		result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, reason, false))
		return
	}

	var findings []evidencedto.Evidence
	var raws []collectevidence.RawFile
	var failures []evidencedto.ToolFailure
	for _, file := range files {
		tagged, err := coresarif.TagRuns(file.Data, map[string]any{
			"provenance": evidence.ProvenanceBridge.String(),
			"bridge":     bridge.Name,
			"guaranteed": false,
		})
		if err != nil {
			result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, fmt.Sprintf("%s: %v", file.Name, err), false))
			return
		}
		fileFailures, err := coresarif.NewParser().ParseToolExecutions(file.Data)
		if err != nil {
			result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, fmt.Sprintf("%s: %v", file.Name, err), false))
			return
		}
		failures = append(failures, fileFailures...)

		cmd, err := ingestfind.NewCommand(tagged, "sarif", bridge.Name, "code_quality")
		if err != nil {
			result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, fmt.Sprintf("%s: %v", file.Name, err), false))
			return
		}
		parsed, err := c.parser.Execute(ctx, cmd.WithProvenance(evidence.ProvenanceBridge))
		if err != nil {
			result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, fmt.Sprintf("%s: could not be ingested: %v", file.Name, err), false))
			return
		}
		findings = append(findings, parsed.Evidence)
		raws = append(raws, collectevidence.RawFile{
			Format:     "sarif",
			Source:     bridge.Name,
			Data:       tagged,
			Provenance: evidence.ProvenanceBridge.String(),
		})
	}

	result.Evidences = append(result.Evidences, findings...)
	result.RawFiles = append(result.RawFiles, raws...)
	for _, failed := range failures {
		result.Evidences = append(result.Evidences, bridgeFailure(bridge.Name, failed.Reason, failed.Degraded))
	}
}

// contractViolations returns "" when every file conforms, else one message
// naming the first offending files. A single bad file fails the whole bridge:
// partial findings from a tool that breaks its contract would read as a clean
// or smaller result that nobody can trust.
func contractViolations(files []runner.BridgeFile) string {
	var offenders []string
	for _, file := range files {
		if problems := coresarif.ValidateBridgeSubmission(file.Data); len(problems) > 0 {
			offenders = append(offenders, fmt.Sprintf("%s: %s", file.Name, problems[0]))
		}
	}
	if len(offenders) == 0 {
		return ""
	}
	listed := offenders
	if len(listed) > maxListedOffenders {
		listed = listed[:maxListedOffenders]
	}
	reason := fmt.Sprintf("SARIF contract violated in %d of %d files (see gavel-tools docs/bridge.md): %s", len(offenders), len(files), strings.Join(listed, "; "))
	if extra := len(offenders) - len(listed); extra > 0 {
		reason += fmt.Sprintf("; and %d more", extra)
	}
	return reason
}

func bridgeFailure(name, reason string, degraded bool) evidencedto.Evidence {
	return evidencedto.Evidence{
		Subtype:     evidence.SubtypeToolExecution.String(),
		Source:      name,
		CollectedAt: time.Now().UTC(),
		ToolExecution: &evidencedto.ToolExecution{Failures: []evidencedto.ToolFailure{{
			Tool:       name,
			Reason:     reason,
			Degraded:   degraded,
			Provenance: evidence.ProvenanceBridge.String(),
		}}},
	}
}
