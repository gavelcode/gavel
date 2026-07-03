---
title: MCP server integration
type: how-to
description: Exposing Gavel config, projects, and findings to LLM agents over the Model Context Protocol.
tags: [ide, mcp, agents]
---

# MCP Server Integration

Gavel exposes a [Model Context Protocol](https://modelcontextprotocol.io)
server that gives LLM agents read access to your quality gate configuration,
findings, coverage, and architecture rules. The server runs on stdio and
delegates every operation to the `gavel` CLI binary.

## Prerequisites

- The `gavel` CLI installed and on your `PATH` (see [Install](../../README.md#install)).
- A workspace where you have run `gavel init`.

## Claude Code

Add to `.claude/settings.json` (project-level) or `~/.claude/settings.json`
(global):

```json
{
  "mcpServers": {
    "gavel": {
      "command": "gavel",
      "args": ["mcp"],
      "env": {
        "GAVEL_WORKSPACE": "${workspaceFolder}"
      }
    }
  }
}
```

## Other MCP clients

Any MCP-compatible client (Cursor, Windsurf, Zed, etc.) uses the same
command. Adapt the JSON to the client's settings format — the command and
args are identical.

## Exposed tools

| Tool | Description |
|------|-------------|
| `gavel_judge` | Run static analyzers and evaluate the quality gate |
| `gavel_findings` | Discover all findings across a project, with by-rule summary and rule/severity filters |
| `gavel_lint_file` | Lint findings for a specific file |
| `gavel_coverage` | Extract coverage percentages, with per-file or per-package (`pkg/...`) uncovered-line breakdown, plus `diff` mode (change since last green baseline) |
| `gavel_validate` | Check workspace structural setup |
| `gavel_trends` | Show analysis history (coverage and findings over time) |
| `gavel_arch` | Report architecture (layer) violations for a project |
| `gavel_init` | Initialize Gavel in a workspace |

## Exposed resources

| URI | Description |
|-----|-------------|
| `gavel://config` | Workspace configuration (`gavel.yaml`) |
| `gavel://projects` | List of projects with patterns and gates |
| `gavel://projects/{name}/quality-gate` | Quality gate rules for a project |
| `gavel://projects/{name}/baseline` | Baseline state (findings/violations) |
| `gavel://architecture` | Architecture policy layers and deny rules |

## Dogfooding (gavel's own repo)

In the Gavel monorepo itself, the CLI is a local target, not an external
dependency:

```json
{
  "mcpServers": {
    "gavel": {
      "command": "bazel",
      "args": ["run", "//apps/cli/cmd/gavel", "--", "mcp"]
    }
  }
}
```

## Troubleshooting

**`gavel` not found** — Ensure the CLI is installed and on your `PATH`
(`gavel --version` should print a version).

**`GAVEL_WORKSPACE` not set** — The MCP server uses this variable to locate
the workspace root. If not set, it defaults to the directory the client
launches it from.
