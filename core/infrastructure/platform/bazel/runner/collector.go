package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

type SARIFReport struct {
	Data   []byte
	Source string
}

type FindingsCollector interface {
	Collect(ctx context.Context, workspace string, targets []string) ([]SARIFReport, error)
}

type AspectCollector struct {
	aspects []catalog.Aspect
	runner  CommandRunner
}

func NewAspectCollector(aspects []catalog.Aspect, cmd CommandRunner) *AspectCollector {
	return &AspectCollector{aspects: aspects, runner: cmd}
}

func (c *AspectCollector) Collect(ctx context.Context, workspace string, targets []string) ([]SARIFReport, error) {
	var reports []SARIFReport
	for _, asp := range c.aspects {
		sarifFiles, err := runAspect(ctx, c.runner, workspace, targets, asp)
		if err != nil {
			return nil, fmt.Errorf("aspect %s: %w", asp.Name, err)
		}
		for _, data := range sarifFiles {
			reports = append(reports, SARIFReport{Data: data, Source: ExtractToolName(data, asp.Name)})
		}
	}
	return reports, nil
}

type ReportCollector struct {
	runner CommandRunner
}

func NewReportCollector(cmd CommandRunner) *ReportCollector {
	return &ReportCollector{runner: cmd}
}

func (c *ReportCollector) Collect(ctx context.Context, workspace string, _ []string) ([]SARIFReport, error) {
	binDir, err := bazelBinDir(ctx, c.runner, workspace)
	if err != nil {
		return nil, err
	}
	return CollectReportsFromDir(binDir)
}

func BazelBinDir(ctx context.Context, workspace string) (string, error) {
	return bazelBinDir(ctx, &ExecRunner{}, workspace)
}

func bazelBinDir(ctx context.Context, cmd CommandRunner, workspace string) (string, error) {
	stdout, stderr, err := cmd.Run(ctx, workspace, "bazel", "info", "bazel-bin")
	if err != nil {
		return "", fmt.Errorf("bazel info bazel-bin: %w\n%s", err, string(stderr))
	}
	return strings.TrimSpace(string(stdout)), nil
}

func CollectReportsFromDir(dir string) ([]SARIFReport, error) {
	var reports []SARIFReport
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !isRulesLintReport(entry.Name()) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read report %s: %w", path, readErr)
		}
		source := ExtractToolName(data, entry.Name())
		reports = append(reports, SARIFReport{Data: data, Source: source})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan bazel-bin: %w", err)
	}
	return reports, nil
}

func isRulesLintReport(name string) bool {
	return strings.Contains(name, "AspectRulesLint") && strings.HasSuffix(name, "report")
}

type HybridCollector struct {
	primary    FindingsCollector
	supplement FindingsCollector
}

func NewHybridCollector(exclusiveAspects []catalog.Aspect, cmd CommandRunner) *HybridCollector {
	return &HybridCollector{
		primary:    NewReportCollector(cmd),
		supplement: NewAspectCollector(exclusiveAspects, cmd),
	}
}

func (c *HybridCollector) Collect(ctx context.Context, workspace string, targets []string) ([]SARIFReport, error) {
	reports, err := c.primary.Collect(ctx, workspace, targets)
	if err != nil {
		return nil, err
	}
	supplements, err := c.supplement.Collect(ctx, workspace, targets)
	if err != nil {
		return nil, err
	}
	return append(reports, supplements...), nil
}

func HasRulesLintReports(ctx context.Context, workspace string) bool {
	return hasRulesLintReportsWithRunner(ctx, &ExecRunner{}, workspace)
}

func hasRulesLintReportsWithRunner(ctx context.Context, cmd CommandRunner, workspace string) bool {
	binDir, err := bazelBinDir(ctx, cmd, workspace)
	if err != nil {
		return false
	}
	return HasRulesLintReportsInDir(binDir)
}

func HasRulesLintReportsInDir(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if isRulesLintReport(entry.Name()) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func ExtractToolName(data []byte, filename string) string {
	var doc struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Name string `json:"name"`
				} `json:"driver"`
			} `json:"tool"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &doc); err == nil && len(doc.Runs) > 0 {
		name := doc.Runs[0].Tool.Driver.Name
		if name != "" {
			return name
		}
	}
	return filename
}
