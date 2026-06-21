---
title: VS Code integration
type: how-to
description: Viewing Gavel SARIF findings inline in VS Code via the SARIF Viewer extension.
---

# VS Code Integration

Gavel produces SARIF 2.1.0 reports that VS Code can display inline via the
SARIF Viewer extension. Findings appear as editor diagnostics with file
locations, severity, and rule IDs.

## Setup

1. Install the [SARIF Viewer](https://marketplace.visualstudio.com/items?itemName=MS-SarifVSCode.sarif-viewer)
   extension from the VS Code marketplace.

2. Run Gavel with the `--output-sarif` flag:

   ```bash
   gavel judge --output-sarif .gavel/report.sarif
   ```

3. Open the generated SARIF file in VS Code. The SARIF Viewer extension
   activates automatically for `.sarif` files.

4. Navigate findings in the **SARIF** panel (View > Open View > SARIF).
   Clicking a finding jumps to the source location.

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
  The SARIF Viewer picks up file changes automatically.
- The SARIF report includes findings from all configured tools (golangci-lint,
  PMD, SpotBugs, ESLint, etc.) merged into a single file.
