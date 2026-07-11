package json

import (
	encjson "encoding/json"
	"io"
	"math"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/tree"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

type rulingDTO struct {
	Subtype string `json:"subtype"`
	Passed  bool   `json:"passed"`
	Detail  string `json:"detail"`
}

type findingDTO struct {
	Tool          string `json:"tool"`
	RuleID        string `json:"rule_id"`
	Severity      string `json:"severity"`
	FilePath      string `json:"file_path"`
	Line          int    `json:"line"`
	Message       string `json:"message"`
	FingerprintID string `json:"fingerprint"`
	Status        string `json:"status,omitempty"`
}

type violationDTO struct {
	Rule      string `json:"rule"`
	SourcePkg string `json:"source_pkg"`
	TargetPkg string `json:"target_pkg"`
	Message   string `json:"message"`
	Status    string `json:"status,omitempty"`
}

type deltaDTO struct {
	HasPrevious             bool    `json:"has_previous"`
	NewCount                int     `json:"new_count"`
	FixedCount              int     `json:"fixed_count"`
	ExistingCount           int     `json:"existing_count"`
	FindingsDelta           int     `json:"findings_delta"`
	CoverageDelta           float64 `json:"coverage_delta"`
	ViolationsDelta         int     `json:"violations_delta"`
	NewViolationsCount      int     `json:"new_violations_count,omitempty"`
	FixedViolationsCount    int     `json:"fixed_violations_count,omitempty"`
	ExistingViolationsCount int     `json:"existing_violations_count,omitempty"`
}

type fileCoverageDTO struct {
	FilePath        string   `json:"file_path"`
	CoveredLines    int      `json:"covered_lines"`
	TotalLines      int      `json:"total_lines"`
	Percent         float64  `json:"percent"`
	Covered         []int    `json:"covered"`
	Uncovered       []int    `json:"uncovered"`
	PreviousPercent *float64 `json:"previous_percent,omitempty"`
	CoverageDelta   *float64 `json:"coverage_delta,omitempty"`
	IsNew           bool     `json:"is_new,omitempty"`
}

type projectDTO struct {
	Name                   string             `json:"name"`
	Verdict                string             `json:"verdict"`
	CommitSHA              string             `json:"commit_sha,omitempty"`
	Branch                 string             `json:"branch,omitempty"`
	StartedAt              string             `json:"started_at,omitempty"`
	FindingsCount          int                `json:"findings_count"`
	ViolationsCount        int                `json:"violations_count"`
	CoveragePercent        *float64           `json:"coverage_percent,omitempty"`
	NewCodeCoveragePercent float64            `json:"new_code_coverage_percent"`
	CoverageByFile         []fileCoverageDTO  `json:"coverage_by_file,omitempty"`
	Rulings                []rulingDTO        `json:"rulings"`
	Findings               []findingDTO       `json:"findings"`
	CoverageTree           *tree.CoverageNode `json:"coverage_tree,omitempty"`
	FindingsTree           *tree.FindingsNode `json:"findings_tree,omitempty"`
	Violations             []violationDTO     `json:"violations,omitempty"`
	Delta                  *deltaDTO          `json:"delta,omitempty"`
	BuildWarning           string             `json:"build_warning,omitempty"`
	UnanalyzedTools        []string           `json:"unanalyzed_tools,omitempty"`
}

type responseDTO struct {
	Projects []projectDTO `json:"projects"`
}

func Write(writer io.Writer, results []pipeline.Result) error {
	projects := make([]projectDTO, 0, len(results))
	for _, result := range results {
		projects = append(projects, toProjectDTO(result))
	}

	enc := encjson.NewEncoder(writer)
	enc.SetIndent("", "  ")
	return enc.Encode(responseDTO{Projects: projects})
}

func toProjectDTO(result pipeline.Result) projectDTO {
	var covPct *float64
	if !result.CoverageSkipped {
		covPct = &result.CoveragePercent
	}
	var startedAt string
	if !result.StartedAt.IsZero() {
		startedAt = result.StartedAt.UTC().Format(time.RFC3339)
	}
	return projectDTO{
		Name:                   result.Name,
		Verdict:                result.Verdict,
		CommitSHA:              result.CommitSHA,
		Branch:                 result.Branch,
		StartedAt:              startedAt,
		FindingsCount:          result.FindingsCount,
		ViolationsCount:        result.ViolationsCount,
		CoveragePercent:        covPct,
		NewCodeCoveragePercent: result.NewCodeCoveragePercent,
		CoverageByFile:         toFileCoverageDTOs(result),
		CoverageTree:           buildCoverageTree(result),
		FindingsTree:           buildFindingsTree(result),
		Rulings:                toRulingDTOs(result.Rulings),
		Findings:               toFindingDTOs(result),
		Violations:             toViolationDTOs(result),
		Delta:                  toDeltaDTO(result),
		BuildWarning:           result.BuildWarning,
		UnanalyzedTools:        result.UnanalyzedTools,
	}
}

const (
	percentScale   = 100
	decimalQuantum = 10
)

func roundPercent(p float64) float64 {
	return math.Round(p*decimalQuantum) / decimalQuantum
}

func toFileCoverageDTOs(result pipeline.Result) []fileCoverageDTO {
	if result.CoverageSkipped {
		return nil
	}
	previous := previousPercents(result.PreviousCoverageByFile)
	out := make([]fileCoverageDTO, 0, len(result.CoverageByFile))
	for _, fileCov := range result.CoverageByFile {
		covered := len(fileCov.Covered)
		total := covered + len(fileCov.Uncovered)
		var percent float64
		if total > 0 {
			percent = roundPercent(float64(covered) / float64(total) * percentScale)
		}
		dto := fileCoverageDTO{
			FilePath:     fileCov.FilePath,
			CoveredLines: covered,
			TotalLines:   total,
			Percent:      percent,
			Covered:      fileCov.Covered,
			Uncovered:    fileCov.Uncovered,
		}
		applyCoverageDiff(&dto, result.Delta.HasPrevious, previous, percent)
		out = append(out, dto)
	}
	return out
}

func applyCoverageDiff(dto *fileCoverageDTO, hasPrevious bool, previous map[string]float64, percent float64) {
	if !hasPrevious {
		return
	}
	prev, ok := previous[dto.FilePath]
	if !ok {
		dto.IsNew = true
		return
	}
	delta := roundPercent(percent - prev)
	dto.PreviousPercent = &prev
	dto.CoverageDelta = &delta
}

func previousPercents(entries []evidencedto.FileCoverage) map[string]float64 {
	out := make(map[string]float64, len(entries))
	for _, entry := range entries {
		covered := len(entry.Covered)
		total := covered + len(entry.Uncovered)
		var percent float64
		if total > 0 {
			percent = roundPercent(float64(covered) / float64(total) * percentScale)
		}
		out[entry.FilePath] = percent
	}
	return out
}

func toRulingDTOs(rulings []corejudge.RulingView) []rulingDTO {
	out := make([]rulingDTO, 0, len(rulings))
	for _, rl := range rulings {
		out = append(out, rulingDTO{
			Subtype: rl.Subtype,
			Passed:  rl.Passed,
			Detail:  rl.Detail,
		})
	}
	return out
}

func toFindingDTOs(result pipeline.Result) []findingDTO {
	out := make([]findingDTO, 0, len(result.Findings))
	for _, finding := range result.Findings {
		status := ""
		if result.Delta.HasPrevious {
			if result.Delta.NewFingerprints[finding.FingerprintID] {
				status = "new"
			} else {
				status = "existing"
			}
		}
		out = append(out, findingDTO{
			Tool:          finding.Tool,
			RuleID:        finding.RuleID,
			Severity:      finding.Severity,
			FilePath:      finding.FilePath,
			Line:          finding.Line,
			Message:       finding.Message,
			FingerprintID: finding.FingerprintID,
			Status:        status,
		})
	}
	return out
}

func toViolationDTOs(result pipeline.Result) []violationDTO {
	out := make([]violationDTO, 0, len(result.Violations))
	for _, violation := range result.Violations {
		status := ""
		if result.Delta.HasArchPrevious {
			id := violation.Rule + ":" + violation.SourcePkg + ":" + violation.TargetPkg
			if result.Delta.NewViolationIDs[id] {
				status = "new"
			} else {
				status = "existing"
			}
		}
		out = append(out, violationDTO{
			Rule:      violation.Rule,
			SourcePkg: violation.SourcePkg,
			TargetPkg: violation.TargetPkg,
			Message:   violation.Message,
			Status:    status,
		})
	}
	return out
}

func buildCoverageTree(result pipeline.Result) *tree.CoverageNode {
	if result.CoverageSkipped {
		return nil
	}
	files := make([]tree.FileCoverage, 0, len(result.CoverageByFile))
	for _, fileCov := range result.CoverageByFile {
		files = append(files, tree.FileCoverage{
			FilePath:     fileCov.FilePath,
			CoveredLines: len(fileCov.Covered),
			TotalLines:   len(fileCov.Covered) + len(fileCov.Uncovered),
		})
	}
	return tree.BuildCoverageTree(files)
}

func buildFindingsTree(result pipeline.Result) *tree.FindingsNode {
	if len(result.Findings) == 0 {
		return nil
	}
	inputs := make([]tree.FindingInput, 0, len(result.Findings))
	for _, finding := range result.Findings {
		inputs = append(inputs, tree.FindingInput{
			FilePath: finding.FilePath,
			Severity: finding.Severity,
		})
	}
	return tree.BuildFindingsTree(inputs)
}

func toDeltaDTO(result pipeline.Result) *deltaDTO {
	if !result.Delta.HasPrevious && !result.Delta.HasArchPrevious {
		return nil
	}
	return &deltaDTO{
		HasPrevious:             result.Delta.HasPrevious,
		NewCount:                result.Delta.NewCount,
		FixedCount:              result.Delta.FixedCount,
		ExistingCount:           result.Delta.ExistingCount,
		FindingsDelta:           result.Delta.FindingsDelta,
		CoverageDelta:           result.Delta.CoverageDelta,
		ViolationsDelta:         result.Delta.ViolationsDelta,
		NewViolationsCount:      result.Delta.NewViolationsCount,
		FixedViolationsCount:    result.Delta.FixedViolationsCount,
		ExistingViolationsCount: result.Delta.ExistingViolationsCount,
	}
}
