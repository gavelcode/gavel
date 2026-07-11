package ui

import (
	"fmt"
	"strings"
	"time"
)

func Header(command string) string {
	title := Title.Render(fmt.Sprintf("\u2696  GAVEL  ///  %s", command))
	return "\n" + title + "\n"
}

func PhaseHeader(number, total int, name, description string) string {
	phase := Phase.Render(fmt.Sprintf("\u2696 PHASE %d/%d", number, total))
	label := Title.Render(name)
	desc := Dim.Render(" \u2014 " + description)
	return "\n" + phase + "  " + label + desc + "\n"
}

func PhaseItem(name, status string, ok bool) string {
	statusStyle := Success
	if !ok {
		statusStyle = Error
	}
	return fmt.Sprintf("  ├─ %s  ▌ %s\n", Dim.Render(name), statusStyle.Render(status))
}

func TreeItem(text string) string {
	return fmt.Sprintf("  ├─ %s\n", Dim.Render(text))
}

func TreeLastItem(text string) string {
	return fmt.Sprintf("  └─ %s\n", Dim.Render(text))
}

func TreeLastItemWithStatus(text, status string, ok bool) string {
	statusStyle := Success
	if !ok {
		statusStyle = Error
	}
	return fmt.Sprintf("  └─ %s — %s\n", Dim.Render(text), statusStyle.Render(status))
}

func Verdict(configPath string) string {
	return "\n" + Title.Render("\u2696  SO ORDERED") + Dim.Render(" \u2014 "+configPath) + "\n"
}

func JudgeVerdict(verdict, caseFilePath string, findings, violations int, coveragePercent float64, coverageSkipped bool, elapsed time.Duration) string {
	coverageInfo := ""
	if !coverageSkipped {
		coverageInfo = fmt.Sprintf(" | %.1f%% coverage", coveragePercent)
	}
	elapsedInfo := Dim.Render(fmt.Sprintf("  (%s)", formatDuration(elapsed)))
	if verdict == "pass" {
		return "\n" + Success.Render("\u2696  VERDICT: PASS") + Dim.Render(coverageInfo+" \u2014 "+caseFilePath) + elapsedInfo + "\n"
	}
	detail := fmt.Sprintf("%d findings", findings)
	if violations > 0 {
		detail += fmt.Sprintf(", %d violations", violations)
	}
	return "\n" + Error.Render(fmt.Sprintf("\u2696  VERDICT: FAIL \u2014 %s", detail)) + Dim.Render(coverageInfo+" \u2014 "+caseFilePath) + elapsedInfo + "\n"
}

func formatDuration(dur time.Duration) string {
	dur = dur.Round(time.Second)
	hrs := int(dur.Hours())
	mins := int(dur.Minutes()) % secondsPerMinute
	secs := int(dur.Seconds()) % secondsPerMinute
	if hrs > 0 {
		return fmt.Sprintf("%dh %dm %ds", hrs, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func PhaseHeaderWithElapsed(number, total int, name, description, elapsed string) string {
	phase := Phase.Render(fmt.Sprintf("⚖ PHASE %d/%d", number, total))
	label := Title.Render(name)
	desc := Dim.Render(" — " + description)
	time := Dim.Render(" — " + elapsed)
	return "\n" + phase + "  " + label + desc + time + "\n"
}

func RulingLine(subtype string, passed bool, detail string, last bool) string {
	status := Success.Render("PASS")
	if !passed {
		status = Error.Render("FAIL")
	}
	prefix := "├─"
	if last {
		prefix = "└─"
	}
	info := fmt.Sprintf("%s: %s", subtype, status)
	if detail != "" {
		info += Dim.Render(fmt.Sprintf(" — %s", detail))
	}
	return fmt.Sprintf("  %s %s\n", prefix, info)
}

func FirstRunHint() string {
	return fmt.Sprintf("  %s\n", Dim.Render("first run — baseline saved. Future runs will evaluate new findings only."))
}

func ServerFallbackWarning() string {
	return GoldBar.Render("  ⚠ server unreachable — using local pipeline") + "\n"
}

func BuildWarning() string {
	return GoldBar.Render("  ⚠ bazel build had failures — results may be incomplete; verdict is unreliable") + "\n"
}

func MissingTargetsWarning(project string, tools []string) string {
	msg := fmt.Sprintf("  ⚠ project %s enables %s, but no analyzable targets were found — those tools produced zero findings and are not covered by this verdict",
		project, strings.Join(tools, ", "))
	return GoldBar.Render(msg) + "\n"
}

type ProjectSummary struct {
	Name        string
	Verdict     string
	Findings    int
	NewFindings int
	Coverage    float64
	Violations  int
}

func SummaryTable(projects []ProjectSummary, elapsed time.Duration) string {
	if len(projects) <= 1 {
		return ""
	}

	nameW := len("PROJECT")
	for _, proj := range projects {
		if len(proj.Name) > nameW {
			nameW = len(proj.Name)
		}
	}

	header := fmt.Sprintf("\n%s\n\n  %-*s  %-7s  %10s  %10s  %s\n",
		Title.Render("⚖  SUMMARY"),
		nameW, "PROJECT", "VERDICT", "FINDINGS", "COVERAGE", "ARCHITECTURE")

	var rows string
	failed := 0
	for _, proj := range projects {
		verdict := Success.Render("PASS")
		if proj.Verdict != "pass" {
			verdict = Error.Render("FAIL")
			failed++
		}
		findingsCol := fmt.Sprintf("%d (%d new)", proj.Findings, proj.NewFindings)
		coverageCol := fmt.Sprintf("%.1f%%", proj.Coverage)
		archCol := "—"
		if proj.Violations > 0 {
			archCol = fmt.Sprintf("%d violations", proj.Violations)
		}
		rows += fmt.Sprintf("  %-*s  %s   %10s  %10s  %s\n",
			nameW, proj.Name, verdict, findingsCol, coverageCol, archCol)
	}

	status := Success.Render(fmt.Sprintf("  %d/%d PASSED", len(projects)-failed, len(projects)))
	if failed > 0 {
		status = Error.Render(fmt.Sprintf("  %d/%d FAILED", failed, len(projects)))
	}
	footer := "\n" + status + Dim.Render(fmt.Sprintf("  (%s)", formatDuration(elapsed))) + "\n"

	return header + rows + footer
}

type CoverageItem struct {
	Path         string
	CoveredLines int
	TotalLines   int
	Percent      float64
}

func CoverageBlock(items []CoverageItem) string {
	if len(items) == 0 {
		return ""
	}
	var out string
	out += Dim.Render("  Coverage by directory:") + "\n"
	for i, item := range items {
		prefix := "├─"
		if i == len(items)-1 {
			prefix = "└─"
		}
		pct := fmt.Sprintf("%.1f%%", item.Percent)
		detail := Dim.Render(fmt.Sprintf(" (%d/%d)", item.CoveredLines, item.TotalLines))
		out += fmt.Sprintf("  %s %s %s%s\n", prefix, item.Path, pct, detail)
	}
	return out
}

func ExistingConfig(configPath string) string {
	return "\n" + Dim.Render("\u2696  Config already exists:") + " " + Label.Render(configPath) + "\n" +
		Dim.Render("   Use --force to overwrite.") + "\n"
}
