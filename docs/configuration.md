---
title: Configuration reference
type: reference
description: The gavel.yaml and architecture.yml schema — projects, quality-gate rules, and DDD layer definitions.
tags: [configuration, gavel.yaml, architecture, quality-gate]
---

# Configuration Reference

Gavel uses two configuration files, both in the `.gavel/` directory of a Bazel
workspace:

| File | Purpose | Required |
|------|---------|----------|
| `.gavel/gavel.yaml` | Project definitions, quality gate rules, server | Yes |
| `.gavel/architecture.yml` | Layer definitions and dependency rules | No |

Both are created by `gavel init`. The quality gate and architecture config can
be added or modified manually after initialization.

---

## gavel.yaml

Defines the workspace name, projects, quality gate rules, and server connection.

### Full schema

```yaml
name: <string>                              # workspace name (required)

projects:                                    # list of projects (required, at least one)
  - name: <string>                           # project name (required)
    pattern: <string>                        # Bazel target pattern (required, e.g. "//..." or "//payments/...")
    exclude:                                 # paths to drop from the whole gate (optional)
      - <pattern>                            # Bazel pattern within `pattern`, e.g. "//payments/gen/..."
    tooling:                                 # language → tools to run (required, at least one language)
      <language>: [<tool>, ...]              # e.g. go: [golangci-lint, archtest]

    quality_gate:                            # quality gate rules (optional)
      findings:                              # code quality findings rule (optional)
        max_error: <int>                     # maximum NEW errors allowed (optional)
        max_warning: <int>                   # maximum NEW warnings allowed (optional)
        max_note: <int>                      # maximum NEW notes allowed (optional)
        min_resolved: <int>                  # minimum findings resolved per run (optional)

      coverage:                              # coverage rule (optional)
        min: <float>                         # minimum coverage percentage, 0-100 (required)
        min_delta: <float>                   # minimum coverage improvement per run (optional)

      architecture_violations:               # architecture violations rule (optional)
        max: <int>                           # maximum NEW violations allowed (required)
        min_resolved: <int>                  # minimum violations resolved per run (optional)

server:                                      # server connection (optional)
  url: <string>                              # server URL (overridden by --server flag)
  token: <string>                            # API token (overridden by --token flag)
```

### Fields

**`name`** (string, required)
Workspace name. Used for display and identification.

**`projects`** (list, required)
One or more projects to analyze. Each project maps to a set of Bazel targets
and a set of languages.

**`projects[].pattern`** (string, required)
Bazel target pattern defining the scope of analysis. Use `//...` for the entire
workspace or `//payments/...` for a subtree. Must be a valid Bazel target
pattern.

**`projects[].exclude`** (list of strings, optional)
Bazel target patterns to drop from the project's scope. Each must be a valid
pattern that resolves **within** `pattern` (an out-of-scope exclude is a config
error). Excluded targets count for nothing across the **whole gate** — not
findings, not coverage, not architecture — so it is the right home for generated
or vendored code (e.g. `//payments/gen/...`). This is a coarse, file/dir-level
scope decision, never a per-line coverage exclusion; see
[coverage exclusion policy](design/coverage-exclusion-policy.md).

**`projects[].tooling`** (map, required)
Maps each language to the list of tools to run for it — e.g.
`go: [golangci-lint, archtest]`. The available languages and their tools are the
gavel-tools catalog (see [Language support](languages/index.md), the single
source of truth); a project runs exactly what it lists. Selection is
**explicit**: a language with no tools is a config error, not a silent "run
everything" default, and listing a tool the catalog does not publish for that
language is likewise an error.

**`projects[].quality_gate`** (object, optional)
Quality gate rules. When omitted, no quality gate is applied and the verdict is
always pass.

### Quality gate rules

All rules are evaluated as **blocking** — the verdict fails if any rule is
violated. The overall verdict passes only when every rule passes.

Thresholds for findings and architecture violations evaluate only **new**
items compared to the baseline. Existing items (already in the baseline)
are excluded before evaluation. This means `max_error: 0` blocks on zero
new errors, not zero total errors.

Use `--absolute` to evaluate against all findings regardless of baseline
(useful for release gates or audits).

**`findings`** — limits on new code quality findings by severity. When no
thresholds are specified, defaults to zero tolerance (no new findings
allowed). Optional `min_resolved` requires a minimum number of findings
to be resolved per run.

**`coverage`** — minimum test coverage percentage. The `min` field is
required when this rule is present. Optional `min_delta` requires coverage
to improve by at least the given amount per run (e.g., `min_delta: 0.0`
means "do not regress").

**`architecture_violations`** — maximum allowed new architecture rule
violations. The `max` field is required when this rule is present.
Requires `architecture.yml` to be configured. Optional `min_resolved`
requires a minimum number of violations to be resolved per run.

### Supported languages and tools

See [Language support](languages/index.md) for the tools and pinned versions
per language (the single source of truth for that table).

### Examples

Minimal (single project, no quality gate):

```yaml
name: my-project
projects:
  - name: my-project
    pattern: "//..."
    tooling:
      go: [golangci-lint, archtest]
```

Multi-project monorepo with quality gate:

```yaml
name: my-monorepo
projects:
  - name: backend
    pattern: "//backend/..."
    tooling:
      java: [error-prone, spotbugs, archtest]
      go: [golangci-lint, archtest]
    quality_gate:
      findings:
        max_error: 0
        max_warning: 10
      coverage:
        min: 80
      architecture_violations:
        max: 0

  - name: frontend
    pattern: "//frontend/..."
    tooling:
      typescript: [eslint, archtest]
    quality_gate:
      findings:
        max_error: 0
      coverage:
        min: 60
```

Legacy codebase adoption (require progress, don't block on existing debt):

```yaml
name: legacy-monorepo
projects:
  - name: backend
    pattern: "//backend/..."
    tooling:
      java: [error-prone, spotbugs, archtest]
    quality_gate:
      findings:
        max_error: 0
        min_resolved: 1
      coverage:
        min: 40
        min_delta: 0.0
      architecture_violations:
        max: 0
```

---

## architecture.yml

Defines architectural layers and dependency rules enforced by archtest aspects.
When present, `gavel judge` loads this file and runs architecture checks for
each project that has archtest aspects available.

Located at `.gavel/architecture.yml`.

### Full schema

```yaml
version: <int>                               # config version (optional)
module: <string>                             # module name (optional)

layers:                                      # layer definitions (required, at least one)
  <layer_name>:                              # layer name (string key)
    - <pattern>                              # package patterns (at least one per layer)

rules:                                       # deny rules (optional)
  - name: <string>                           # rule name (required)
    source: <string>                         # source layer (required, must exist in layers)
    deny:                                    # denied target layers (required, at least one)
      - <string>                             # layer name (must exist, cannot deny self)

detect_cycles: <bool>                        # detect circular dependencies (optional)

generic:                                     # generic options (optional)
  no_circular_deps: <bool>                   # alias for detect_cycles
```

### Fields

**`layers`** (map, required)
Maps layer names to lists of package patterns. Patterns use `...` suffix for
recursive matching (e.g., `internal/domain/...` matches all packages under
`internal/domain/`).

**`rules`** (list, optional)
Deny rules that specify which layers a source layer may not import from. Each
rule validates that:
- `source` exists in `layers`
- Every entry in `deny` exists in `layers`
- `deny` does not include the `source` layer itself

**`detect_cycles`** (bool, optional)
When true, checks for circular dependencies between layers.
`generic.no_circular_deps` is an alias for this field.

### How architecture checking works

Gavel generates architecture tests (archtest aspects) that Bazel executes
as part of `gavel judge`. Each test analyzes import paths in your source
code and verifies that packages in one layer do not import packages from
forbidden layers. The analysis is static — it checks import declarations,
not runtime behavior.

When a package in `domain` imports a package in `infrastructure`, Gavel
reports a violation identified as `rule:sourcePkg:targetPkg` (e.g.,
`domain-imports-nothing:domain/order:infrastructure/persistence`). These
violations are evaluated by the quality gate as delta — only **new**
violations compared to the baseline affect the verdict.

`gavel init` generates default layer patterns based on the project's
language:

| Language | Layer pattern style |
|----------|-------------------|
| Go | `internal/domain/...`, `internal/application/...` |
| Java | `src/main/java/**/domain/...` |
| Python, TypeScript, Rust | `src/domain/...` |

### Writing custom rules

Each deny rule says "packages in the source layer may not import packages
from the denied layers." Compose multiple rules to express your dependency
direction:

```yaml
rules:
  # domain depends on nothing — pure business logic
  - name: domain-imports-nothing
    source: domain
    deny: [application, infrastructure, userinterface]

  # application depends on domain only — use cases
  - name: application-imports-domain-only
    source: application
    deny: [infrastructure, userinterface]
```

`detect_cycles: true` additionally checks for circular dependencies
between layers, independent of the deny rules.

### Example

DDD layer isolation (Vernon strict):

```yaml
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
  infrastructure: ["internal/infrastructure/..."]
  userinterface: ["internal/userinterface/..."]

rules:
  - name: domain-imports-nothing
    source: domain
    deny: [application, infrastructure, userinterface]

  - name: application-imports-domain-only
    source: application
    deny: [infrastructure, userinterface]

  - name: infrastructure-no-application
    source: infrastructure
    deny: [application, userinterface]

  - name: userinterface-application-only
    source: userinterface
    deny: [domain, infrastructure]

detect_cycles: true
```

---

## File locations

The `.gavel/` directory at the workspace root contains both user-authored
configuration and generated artifacts:

| File | Type | Version control |
|------|------|-----------------|
| `.gavel/gavel.yaml` | User config | Yes |
| `.gavel/architecture.yml` | User config | Yes |
| `.gavel/baseline/<project>/` | Generated by `gavel judge` | Yes (committed) |
| `.gavel/gavel.bazelrc` | Generated by `gavel init` | No (gitignore) |
| `.gavel/gavel.MODULE.bazel` | Generated by `gavel init` | No (gitignore) |
| `.gavel/results/` | Generated by `gavel judge` | No (gitignore) |

`gavel judge` looks for configuration in this order:

1. Explicit `--config` flag (if provided)
2. `.gavel/gavel.yaml` (canonical location)
3. `gavel.yaml` (workspace root, fallback)

Architecture config is always at `.gavel/architecture.yml` and loaded
automatically when present.
