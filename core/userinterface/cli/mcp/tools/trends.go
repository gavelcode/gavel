package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

const (
	shortSHALength     = 7
	hoursPerDay        = 24
	directionThreshold = 0.01
)

type TrendsInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	Project    string `json:"project"              jsonschema:"Project name (required)"`
	Limit      int    `json:"limit,omitempty"       jsonschema:"Number of recent runs to show (default 10)"`
	Branch     string `json:"branch,omitempty"      jsonschema:"Filter to a specific branch"`
}

var openWorld = true

func RegisterTrends(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
		Name:        "gavel_trends",
		Description: "Show quality trends for a project — coverage, findings, and verdict history over recent analysis runs. Requires a Gavel server (GAVEL_SERVER_URL).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TrendsInput) (*mcp.CallToolResult, any, error) {
		result, err := runTrends(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runTrends(ctx context.Context, cli *executor.CLI, input TrendsInput) (string, error) {
	args := []string{"trends", "--project", input.Project}
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, "--limit", strconv.Itoa(limit))
	if input.Branch != "" {
		args = append(args, "--branch", input.Branch)
	}

	output, _, err := cli.RunInJSON(ctx, input.Gavelspace, args...)
	if err != nil {
		return "", fmt.Errorf("execute gavel trends: %w", err)
	}

	return formatTrendsOutput(output, input.Project)
}

type trendEntry struct {
	CommitSHA        string   `json:"commit_sha"`
	Branch           string   `json:"branch"`
	CoveragePercent  *float64 `json:"coverage_percent,omitempty"`
	TotalFindings    int      `json:"total_findings"`
	NewFindings      int      `json:"new_findings"`
	ResolvedFindings int      `json:"resolved_findings"`
	VerdictOutcome   string   `json:"verdict_outcome"`
	CreatedAt        string   `json:"created_at"`
}

func formatTrendsOutput(data []byte, project string) (string, error) {
	var entries []trendEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return "", fmt.Errorf("parse trends output: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("No analysis history found for project %s. Run `gavel judge --server` to start collecting data.", project), nil
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# Trends — %s (last %d runs)\n\n", project, len(entries))
	builder.WriteString("Commit   | Coverage | Findings | New | Fixed | Verdict | When\n")
	builder.WriteString("---------|----------|----------|-----|-------|---------|-----\n")

	for _, entry := range entries {
		cov := "  —"
		if entry.CoveragePercent != nil {
			cov = fmt.Sprintf("%.1f%%", *entry.CoveragePercent)
		}
		sha := entry.CommitSHA
		if len(sha) > shortSHALength {
			sha = sha[:shortSHALength]
		}
		fmt.Fprintf(&builder, "%-8s | %8s | %8d | %3d | %5d | %-7s | %s\n",
			sha, cov, entry.TotalFindings, entry.NewFindings, entry.ResolvedFindings,
			entry.VerdictOutcome, trendTimeAgo(entry.CreatedAt))
	}

	builder.WriteString("\n")
	writeTrendSummary(&builder, entries)
	return builder.String(), nil
}

func writeTrendSummary(builder *strings.Builder, entries []trendEntry) {
	newest := entries[0]
	oldest := entries[len(entries)-1]

	if newest.CoveragePercent != nil && oldest.CoveragePercent != nil {
		delta := *newest.CoveragePercent - *oldest.CoveragePercent
		arrow := trendArrow(delta)
		fmt.Fprintf(builder, "Coverage: %.1f%% (%s %+.1f%% over %d runs)\n",
			*newest.CoveragePercent, arrow, delta, len(entries))
	}

	findingsDelta := float64(newest.TotalFindings - oldest.TotalFindings)
	arrow := trendArrow(-findingsDelta)
	fmt.Fprintf(builder, "Findings: %d (%s %+d over %d runs)\n",
		newest.TotalFindings, arrow, newest.TotalFindings-oldest.TotalFindings, len(entries))

	passCount := 0
	for _, entry := range entries {
		if entry.VerdictOutcome == "pass" {
			passCount++
		}
	}
	fmt.Fprintf(builder, "Pass rate: %d/%d\n", passCount, len(entries))
}

func trendArrow(delta float64) string {
	if math.Abs(delta) < directionThreshold {
		return "→"
	}
	if delta > 0 {
		return "▲"
	}
	return "▼"
}

func trendTimeAgo(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	dur := time.Since(t)
	switch {
	case dur < time.Minute:
		return "just now"
	case dur < time.Hour:
		return fmt.Sprintf("%dm ago", int(dur.Minutes()))
	case dur < hoursPerDay*time.Hour:
		return fmt.Sprintf("%dh ago", int(dur.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(dur.Hours()/hoursPerDay))
	}
}
