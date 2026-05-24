package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type FindingItem struct {
	FilePath string
	Line     int
	Severity string
	Message  string
	Tool     string
	RuleID   string
	IsNew    bool
}

type ViolationItem struct {
	Rule      string
	SourcePkg string
	TargetPkg string
	Message   string
	IsNew     bool
}

func FindingsBlock(findings []FindingItem) string {
	if len(findings) == 0 {
		return ""
	}

	grouped := groupByFile(findings)
	paths := sortedKeys(grouped)

	var builder strings.Builder
	builder.WriteString("\n")
	for _, path := range paths {
		fmt.Fprintf(&builder, "  %s\n", Label.Render(path))
		items := grouped[path]
		slices.SortFunc(items, func(a, builder FindingItem) int {
			return cmp.Compare(a.Line, builder.Line)
		})
		for _, finding := range items {
			builder.WriteString(formatFinding(finding))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func ToolSummary(findings []FindingItem) string {
	if len(findings) == 0 {
		return ""
	}

	counts := make(map[string]int)
	for _, finding := range findings {
		counts[finding.Tool]++
	}

	tools := make([]string, 0, len(counts))
	for t := range counts {
		tools = append(tools, t)
	}
	slices.Sort(tools)

	parts := make([]string, len(tools))
	for i, t := range tools {
		parts[i] = fmt.Sprintf("%s: %d", t, counts[t])
	}
	return fmt.Sprintf("  %s\n", Dim.Render(strings.Join(parts, " · ")))
}

func ViolationsBlock(violations []ViolationItem) string {
	if len(violations) == 0 {
		return ""
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "  %s\n", Label.Render("arch/violations"))
	for _, violation := range violations {
		arrow := fmt.Sprintf("%s → %s", violation.SourcePkg, violation.TargetPkg)
		newBadge := ""
		if violation.IsNew {
			newBadge = "  " + Success.Render("NEW")
		}
		fmt.Fprintf(&builder, "    %-20s %s  %s%s\n",
			Dim.Render(violation.Rule),
			arrow,
			Dim.Render(violation.Message),
			newBadge,
		)
	}
	return builder.String()
}

func formatFinding(finding FindingItem) string {
	sevStyle := severityStyle(finding.Severity)
	ruleRef := Dim.Render(fmt.Sprintf("%s:%s", finding.Tool, finding.RuleID))
	newBadge := ""
	if finding.IsNew {
		newBadge = "  " + Success.Render("NEW")
	}
	return fmt.Sprintf("    %6d  %-7s  %s  %s%s\n",
		finding.Line,
		sevStyle.Render(finding.Severity),
		finding.Message,
		ruleRef,
		newBadge,
	)
}

func DeltaSummary(findingsCount int, findingsDelta int, coveragePercent, coverageDelta float64, newCount, fixedCount, existingCount int, hasPrevious bool, archNew, archFixed, archExisting int, hasArchPrevious bool) string {
	if !hasPrevious && !hasArchPrevious {
		return ""
	}

	var builder strings.Builder
	if hasPrevious {
		fmt.Fprintf(&builder, "  %s\n", deltaLine("findings", findingsCount, findingsDelta))

		if coveragePercent > 0 || coverageDelta != 0 {
			covDir := "↑"
			if coverageDelta < 0 {
				covDir = "↓"
			}
			if coverageDelta == 0 {
				covDir = "="
			}
			absCovDelta := coverageDelta
			if absCovDelta < 0 {
				absCovDelta = -absCovDelta
			}
			fmt.Fprintf(&builder, "  %s\n",
				Dim.Render(fmt.Sprintf("coverage: %.1f%% (%s%.1f%%)", coveragePercent, covDir, absCovDelta)),
			)
		}

		fmt.Fprintf(&builder, "  %s\n",
			Dim.Render(fmt.Sprintf("%d new · %d fixed · %d existing", newCount, fixedCount, existingCount)),
		)
	}

	if hasArchPrevious {
		totalArch := archNew + archExisting
		archDelta := archNew - archFixed
		fmt.Fprintf(&builder, "  %s\n", deltaLine("violations", totalArch, archDelta))
		fmt.Fprintf(&builder, "  %s\n",
			Dim.Render(fmt.Sprintf("%d new · %d fixed · %d existing (arch)", archNew, archFixed, archExisting)),
		)
	}
	return builder.String()
}

func deltaLine(label string, count, delta int) string {
	if delta == 0 {
		return Dim.Render(fmt.Sprintf("%d %s (no change)", count, label))
	}
	dir := "↓"
	style := Success
	if delta > 0 {
		dir = "↑"
		style = Error
	}
	return fmt.Sprintf("%d %s %s",
		count, label,
		style.Render(fmt.Sprintf("(%s%d since last run)", dir, abs(delta))),
	)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func severityStyle(severity string) lipgloss.Style {
	switch severity {
	case "error":
		return Error
	case "warning":
		return GoldBar
	default:
		return Dim
	}
}

func groupByFile(findings []FindingItem) map[string][]FindingItem {
	grouped := make(map[string][]FindingItem)
	for _, finding := range findings {
		grouped[finding.FilePath] = append(grouped[finding.FilePath], finding)
	}
	return grouped
}

func sortedKeys(m map[string][]FindingItem) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
