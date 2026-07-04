package checks

import (
	"fmt"
	"strings"

	outputjson "github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
)

const MaxAnnotationsPerRequest = 50

const (
	ConclusionSuccess = "success"
	ConclusionFailure = "failure"
)

type Level string

const (
	LevelNotice  Level = "notice"
	LevelWarning Level = "warning"
	LevelFailure Level = "failure"
)

const (
	defaultCheckName = "gavel"
	verdictPass      = "pass"
	severityError    = "error"
	severityWarning  = "warning"
	statusExisting   = "existing"
	minGitHubLine    = 1
)

type Annotation struct {
	Path      string
	StartLine int
	EndLine   int
	Level     Level
	Title     string
	Message   string
}

type CheckRun struct {
	Name        string
	HeadSHA     string
	Conclusion  string
	Title       string
	Summary     string
	Annotations []Annotation
}

type Options struct {
	CheckName string
	HeadSHA   string
	NewOnly   bool
}

func Build(verdicts []outputjson.Verdict, opts Options) CheckRun {
	return CheckRun{
		Name:        checkName(opts.CheckName),
		HeadSHA:     headSHA(opts.HeadSHA, verdicts),
		Conclusion:  conclusion(verdicts),
		Title:       title(conclusion(verdicts)),
		Summary:     summary(verdicts),
		Annotations: annotations(verdicts, opts.NewOnly),
	}
}

func BatchAnnotations(annotations []Annotation, maxPerBatch int) [][]Annotation {
	if len(annotations) == 0 || maxPerBatch <= 0 {
		return nil
	}
	var batches [][]Annotation
	for start := 0; start < len(annotations); start += maxPerBatch {
		end := start + maxPerBatch
		if end > len(annotations) {
			end = len(annotations)
		}
		batches = append(batches, annotations[start:end])
	}
	return batches
}

func checkName(name string) string {
	if name == "" {
		return defaultCheckName
	}
	return name
}

func headSHA(override string, verdicts []outputjson.Verdict) string {
	if override != "" {
		return override
	}
	if len(verdicts) > 0 {
		return verdicts[0].CommitSHA
	}
	return ""
}

func conclusion(verdicts []outputjson.Verdict) string {
	for _, verdict := range verdicts {
		if verdict.Verdict != verdictPass {
			return ConclusionFailure
		}
	}
	return ConclusionSuccess
}

func title(conclusion string) string {
	if conclusion == ConclusionFailure {
		return "Gavel: quality gate failed"
	}
	return "Gavel: quality gate passed"
}

func annotations(verdicts []outputjson.Verdict, newOnly bool) []Annotation {
	var out []Annotation
	for _, verdict := range verdicts {
		for _, finding := range verdict.Findings {
			if newOnly && finding.Status == statusExisting {
				continue
			}
			out = append(out, Annotation{
				Path:      finding.FilePath,
				StartLine: clampLine(finding.Line),
				EndLine:   clampLine(finding.Line),
				Level:     level(finding.Severity),
				Title:     findingTitle(finding),
				Message:   finding.Message,
			})
		}
	}
	return out
}

func clampLine(line int) int {
	return max(line, minGitHubLine)
}

func level(severity string) Level {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case severityError:
		return LevelFailure
	case severityWarning:
		return LevelWarning
	default:
		return LevelNotice
	}
}

func findingTitle(finding outputjson.VerdictFinding) string {
	switch {
	case finding.Tool != "" && finding.RuleID != "":
		return finding.Tool + " (" + finding.RuleID + ")"
	case finding.Tool != "":
		return finding.Tool
	default:
		return finding.RuleID
	}
}

func summary(verdicts []outputjson.Verdict) string {
	var builder strings.Builder
	builder.WriteString("## Gavel verdict\n\n")
	builder.WriteString("| Project | Verdict | New | Fixed | Coverage |\n")
	builder.WriteString("|---------|---------|-----|-------|----------|\n")
	for _, verdict := range verdicts {
		newCount, fixedCount := 0, 0
		if verdict.Delta != nil {
			newCount = verdict.Delta.NewCount
			fixedCount = verdict.Delta.FixedCount
		}
		fmt.Fprintf(&builder, "| %s | %s | %d | %d | %s |\n",
			verdict.Name, verdictMark(verdict.Verdict), newCount, fixedCount, coverageCell(verdict.CoveragePercent))
	}
	writeViolations(&builder, verdicts)
	return builder.String()
}

func verdictMark(verdict string) string {
	if verdict == verdictPass {
		return "✅ pass"
	}
	return "❌ fail"
}

func coverageCell(percent *float64) string {
	if percent == nil {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", *percent)
}

func writeViolations(builder *strings.Builder, verdicts []outputjson.Verdict) {
	var violations []outputjson.VerdictViolation
	for _, verdict := range verdicts {
		violations = append(violations, verdict.Violations...)
	}
	if len(violations) == 0 {
		return
	}
	builder.WriteString("\n### Architecture violations\n\n")
	for _, violation := range violations {
		fmt.Fprintf(builder, "- `%s` → `%s` (%s): %s\n",
			violation.SourcePkg, violation.TargetPkg, violation.Rule, violation.Message)
	}
}
