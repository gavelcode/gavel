---
title: IntelliJ IDEA integration
type: how-to
description: Viewing Gavel SARIF findings inline in IntelliJ via the Qodana SARIF plugin.
tags: [ide, intellij, sarif]
---

# IntelliJ IDEA Integration

Gavel produces SARIF 2.1.0 reports that IntelliJ can display inline via the
Qodana SARIF plugin. Findings appear as editor inspections with file
locations, severity, and rule IDs.

## Setup

1. Install the [Qodana](https://plugins.jetbrains.com/plugin/15498-qodana)
   plugin from the JetBrains Marketplace (Settings > Plugins > Marketplace).

2. Run Gavel with the `--output-sarif` flag:

   ```bash
   gavel judge --output-sarif .gavel/report.sarif
   ```

3. Open the SARIF report: **Tools > Qodana > Open SARIF Report**, then
   select `.gavel/report.sarif`.

4. Findings appear in the **Problems** tool window. Clicking a finding
   navigates to the source location in the editor.

## Alternative: SARIF Viewer plugin

If you prefer a lightweight viewer without Qodana:

1. Install the [SARIF Viewer](https://plugins.jetbrains.com/plugin/18868-sarif-viewer)
   plugin from the JetBrains Marketplace.

2. Right-click `.gavel/report.sarif` in the Project view and select
   **Open as SARIF Report**.

## Workflow options

### Full analysis

```bash
gavel judge --output-sarif .gavel/report.sarif
```

Runs linting, coverage, and architecture checks. Produces a merged SARIF
report with all findings from all tools.

### Quick mode (findings only)

```bash
gavel judge --quick --output-sarif .gavel/report.sarif
```

Skips coverage and architecture checks. Faster feedback loop during
development.

### Single project

```bash
gavel judge --project=backend --output-sarif .gavel/report.sarif
```

## Tips

- Add `.gavel/report.sarif` to `.gitignore` to avoid committing generated
  reports.
- Re-run `gavel judge --output-sarif` to refresh findings after code changes.
- The SARIF report includes findings from all configured tools (golangci-lint,
  PMD, SpotBugs, ESLint, etc.) merged into a single file.
