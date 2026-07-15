package trends

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
)

const (
	shortSHALength     = 7
	hoursPerDay        = 24
	directionThreshold = 0.01
)

func NewCommand() *cobra.Command {
	opts := Options{}
	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Show quality trends for a project (requires server)",
		Long: `Fetch historical analysis data from the Gavel server and display
coverage, findings, and verdict trends over recent runs.

Requires GAVEL_SERVER_URL and GAVEL_TOKEN to be configured.`,
		Example: `  gavel trends --project=core
  gavel trends --project=core --limit=20 --branch=main`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, opts)
		},
	}
	RegisterFlags(cmd, &opts)
	return cmd
}

func run(cmd *cobra.Command, opts Options) error {
	if opts.Project == "" {
		return fmt.Errorf("--project is required")
	}
	if opts.ServerURL == "" {
		return fmt.Errorf("trends require a Gavel server; set GAVEL_SERVER_URL or use --server")
	}

	client, err := apiclient.New(opts.ServerURL, opts.ServerToken)
	if err != nil {
		return err
	}

	entries, err := client.ListProjectCaseFiles(cmd.Context(), opts.Project, opts.Limit)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("could not reach Gavel server at %s (is it running?): %w", opts.ServerURL, err)
		}
		return fmt.Errorf("fetch trends: %w", err)
	}

	if opts.Branch != "" {
		entries = filterByBranch(entries, opts.Branch)
	}

	writer := cmd.OutOrStdout()
	if opts.JSONOutput {
		return writeJSON(writer, entries)
	}
	return writeTable(writer, entries, opts.Project, opts.Branch)
}

func filterByBranch(entries []apiclient.TrendEntry, branch string) []apiclient.TrendEntry {
	var filtered []apiclient.TrendEntry
	for _, entry := range entries {
		if entry.Branch == branch {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func writeJSON(writer io.Writer, entries []apiclient.TrendEntry) error {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func writeTable(writer io.Writer, entries []apiclient.TrendEntry, project, branch string) error {
	if len(entries) == 0 {
		if _, err := fmt.Fprintf(writer, "No analysis history found for project %s.", project); err != nil {
			return err
		}
		if branch != "" {
			if _, err := fmt.Fprintf(writer, " (branch: %s)", branch); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer, " Run `gavel judge --server` to start collecting data."); err != nil {
			return err
		}
		return nil
	}

	header := fmt.Sprintf("# Trends — %s (last %d runs", project, len(entries))
	if branch != "" {
		header += fmt.Sprintf(", branch: %s", branch)
	}
	header += ")\n\n"
	if _, err := fmt.Fprint(writer, header); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer, "Commit   | Coverage | Findings | New | Fixed | Verdict | When"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "---------|----------|----------|-----|-------|---------|-----"); err != nil {
		return err
	}
	for _, entry := range entries {
		cov := "  —"
		if entry.CoveragePercent != nil {
			cov = fmt.Sprintf("%.1f%%", *entry.CoveragePercent)
		}
		sha := entry.CommitSHA
		if len(sha) > shortSHALength {
			sha = sha[:shortSHALength]
		}
		if _, err := fmt.Fprintf(writer, "%-8s | %8s | %8d | %3d | %5d | %-7s | %s\n",
			sha, cov, entry.TotalFindings, entry.NewFindings, entry.ResolvedFindings,
			entry.VerdictOutcome, timeAgo(entry.CreatedAt)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return writeSummary(writer, entries)
}

func writeSummary(writer io.Writer, entries []apiclient.TrendEntry) error {
	newest := entries[0]
	oldest := entries[len(entries)-1]

	if newest.CoveragePercent != nil && oldest.CoveragePercent != nil {
		delta := *newest.CoveragePercent - *oldest.CoveragePercent
		arrow := directionArrow(delta)
		if _, err := fmt.Fprintf(writer, "Coverage: %.1f%% (%s %+.1f%% over %d runs)\n",
			*newest.CoveragePercent, arrow, delta, len(entries)); err != nil {
			return err
		}
	}

	findingsDelta := float64(newest.TotalFindings - oldest.TotalFindings)
	arrow := directionArrow(-findingsDelta)
	if _, err := fmt.Fprintf(writer, "Findings: %d (%s %+d over %d runs)\n",
		newest.TotalFindings, arrow, newest.TotalFindings-oldest.TotalFindings, len(entries)); err != nil {
		return err
	}

	passCount := 0
	for _, entry := range entries {
		if entry.VerdictOutcome == "pass" {
			passCount++
		}
	}
	if _, err := fmt.Fprintf(writer, "Pass rate: %d/%d\n", passCount, len(entries)); err != nil {
		return err
	}
	return nil
}

func directionArrow(delta float64) string {
	if math.Abs(delta) < directionThreshold {
		return "→"
	}
	if delta > 0 {
		return "▲"
	}
	return "▼"
}

func timeAgo(t time.Time) string {
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
