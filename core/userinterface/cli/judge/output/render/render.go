package render

import (
	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/tree"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

func Findings(projResult pipeline.Result) string {
	items := make([]ui.FindingItem, 0, len(projResult.Findings))
	for _, finding := range projResult.Findings {
		items = append(items, ui.FindingItem{
			FilePath: finding.FilePath,
			Line:     finding.Line,
			Severity: finding.Severity,
			Message:  finding.Message,
			Tool:     finding.Tool,
			RuleID:   finding.RuleID,
			IsNew:    projResult.Delta.NewFingerprints[finding.FingerprintID],
		})
	}
	vItems := make([]ui.ViolationItem, 0, len(projResult.Violations))
	for _, verdict := range projResult.Violations {
		projectID := verdict.Rule + ":" + verdict.SourcePkg + ":" + verdict.TargetPkg
		vItems = append(vItems, ui.ViolationItem{
			Rule:      verdict.Rule,
			SourcePkg: verdict.SourcePkg,
			TargetPkg: verdict.TargetPkg,
			Message:   verdict.Message,
			IsNew:     projResult.Delta.NewViolationIDs[projectID],
		})
	}
	return ui.FindingsBlock(items) + ui.ToolSummary(items) + ui.ViolationsBlock(vItems)
}

func CoverageSummary(projResult pipeline.Result) string {
	if len(projResult.CoverageByFile) == 0 {
		return ""
	}
	files := make([]tree.FileCoverage, 0, len(projResult.CoverageByFile))
	for _, fc := range projResult.CoverageByFile {
		files = append(files, tree.FileCoverage{
			FilePath:     fc.FilePath,
			CoveredLines: len(fc.Covered),
			TotalLines:   len(fc.Covered) + len(fc.Uncovered),
		})
	}
	root := tree.BuildCoverageTree(files)
	if root == nil || len(root.Children) == 0 {
		return ""
	}
	items := make([]ui.CoverageItem, 0, len(root.Children))
	for _, child := range root.Children {
		items = append(items, ui.CoverageItem{
			Path:         child.Path,
			CoveredLines: child.CoveredLines,
			TotalLines:   child.TotalLines,
			Percent:      child.Percent,
		})
	}
	return ui.CoverageBlock(items)
}
