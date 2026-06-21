---
title: Domain model
type: reference
description: The canonical domain model — aggregates, value objects, invariants, identity types, and events, per Vernon IDDD strict.
---

# Gavel Core — Domain Model

This document follows Vernon's *Implementing Domain-Driven Design* (IDDD) strict
interpretation. Key Vernon rules applied throughout:

- **Reference Other Aggregates by Identity** (IDDD ch. 10) — Aggregates store
  typed identifiers of other aggregates, never holding them by reference.
- **Identity types as Value Objects** (IDDD ch. 5) — IDs are never raw strings.
  `ProjectID`, `CaseFileID`, etc. are VOs with constructors that validate.
- **Small Aggregates, Eventual Consistency** (IDDD ch. 10) — One aggregate per
  transaction. Cross-aggregate updates happen sequentially in application
  services, accepting brief inconsistency windows.
- **Domain Events with stable identity** (IDDD ch. 8) — Every event implements
  the `DomainEvent` interface (name + timestamp). Application drains, never
  the aggregate itself.

## Contents

- [Aggregates](#aggregates) — [Gavelspace](#gavelspace) · [Project](#project) · [CaseFile](#casefile)
  - Each aggregate section ends with an **Editing** checklist (invariants,
    events, critical tests, file map) — start there when modifying.
- [Entities](#entities) — [Evidence](#evidence) · [EvidenceContent](#evidencecontent-interface)
- [Value Objects](#value-objects) — [Identity types](#identity-types) · [Finding](#finding) ·
  [TrackingResult](#trackingresult) · [Ruling](#ruling) · [Verdict](#verdict) ·
  [Severity](#severity) · [QualityGate](#qualitygate) ·
  [Evaluation strategies](#evaluation-strategies-domain-knowledge)
- [Domain Events](#domain-events) — event catalog and dispatch contract
- [Relationships](#relationships) — ASCII map of who owns whom
- [Tracking](#tracking) — fingerprint classification across CaseFiles
- [Use case flow (judge)](#use-case-flow-judge) — end-to-end submit lifecycle
- [Package structure](#package-structure) — folder layout, dependency graph,
  and ownership rationale
- [Construction patterns](#construction-patterns) — `NewXxx` vs `ReconstituteXxx`
- [Known limitations](#known-limitations)

## Aggregates

### Gavelspace

The monorepo-level container. Owns the collection of Projects.

**Identity:** `GavelspaceID` value object (unique name, non-empty).

**Invariants:**
- `name` not empty
- No two projects with the same target pattern

**Children:**
- `projects []ProjectRef` — references to Projects owned by this Gavelspace.
  `ProjectRef` is a VO carrying `ProjectID` + `targetPattern` (no direct
  aggregate reference — Vernon "reference by identity").

**Behavior:**
- `AddProject(ref ProjectRef)` — validates target pattern uniqueness, emits `ProjectAdded`
- `RemoveProject(projectID)` — emits `ProjectRemoved`

No quality gate defaults, no configuration inheritance. Each project is
self-contained.

#### Editing Gavelspace

- **Invariants to preserve**: `name` non-empty; target pattern uniqueness
  across all registered projects.
- **Events to keep emitting**: `ProjectAdded` on AddProject, `ProjectRemoved`
  on RemoveProject. Removing emission breaks the application drain contract.
- **Tests to re-run / extend**: `core/domain/gavelspace/model/*_test.go`,
  in particular target-pattern collision and event emission. Application:
  `core/application/gavelspace/{registerproject,removeproject}/handler_test.go`.
- **File map**: `core/domain/gavelspace/model/{gavelspace.go, gavelspace_name.go, events.go, errors.go}`
  + `core/domain/gavelspace/service/repository.go`.

---

### Project

A directory or Bazel target scope within a Gavelspace that Gavel analyzes.

**Identity:** `ProjectID` value object (UUID generated in domain).

**Attributes:**
- `key` — unique slug identifier (lowercase, hyphens, 1–64 chars, regex `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
- `name` — human-readable name
- `targetPattern` — Bazel recursive target (e.g. `//core/...`)
- `languages` — list of Language in the project
- `defaultBranch` — branch used as baseline for tracking (default: `main`)
- `qualityGate` — quality gate configuration (VO)
- `architecturePolicy` — optional ArchitecturePolicy (layers + deny rules)

**References:** Gavelspace by `GavelspaceID`. Historical CaseFiles reference
Project by `ProjectID` (not contained in memory).

**Behavior:**
- `UpdateQualityGate(qualityGate)` — replaces quality gate config, emits `QualityGateUpdated`
- `UpdateLanguages(languages)` — replaces language list, emits `LanguagesUpdated`

#### Editing Project

- **Invariants to preserve**: `ProjectID` non-zero; `key` 1–64 chars matching
  `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`; `name` non-empty;
  `targetPattern` valid Bazel pattern (`//pkg/...`, `//pkg:target`, `//pkg`);
  `defaultBranch` non-empty; `qualityGate` is a valid VO; each `Language`
  validated by its constructor.
- **Events to keep emitting**: `QualityGateUpdated` on UpdateQualityGate,
  `LanguagesUpdated` on UpdateLanguages. Wholesale replacement, not partial
  mutation.
- **Tests to re-run / extend**: `core/domain/project/model/*_test.go` for
  invariants and event emission. Sub-model: `core/domain/project/model/qualitygate/*_test.go`
  for strategy / rule validation. Application:
  `core/application/project/{updatequalitygate,updatelanguages}/handler_test.go`.
- **File map**: `core/domain/project/model/{project.go, project_id.go, events.go, errors.go}`
  + `core/domain/project/model/qualitygate/*.go`
  + `core/domain/project/service/repository.go`.

---

### CaseFile

An analysis session for a specific commit. Collects evidence progressively
and emits a verdict when judged.

**Identity:** `CaseFileID` value object (UUID generated in domain).

**Invariants:**
- `projectID` not zero
- `commitSHA` not empty
- `branch` not empty
- `startedAt` not zero
- Cannot judge twice (`ErrAlreadyJudged` — judged is a terminal state)
- Cannot add evidence after judging

**Lifecycle:**
```
open → collecting → judged
```

- `open`: created with commit + branch + timestamp
- `collecting`: at least one evidence has been added
- `judged`: `Judge()` has been called, verdict emitted

**Behavior:**
- `AddEvidence(evidence)` — adds an Evidence entity, emits `EvidenceCollected` event
- `Judge(qualityGate, trackingResult?)` — evaluates evidence against the quality gate, produces Verdict with rulings, emits `VerdictRendered`. TrackingResult is optional: without it (CLI), evaluates all findings; with it (server), evaluates only new findings from the PR delta. `Judge()` handles the filtering — strategies always evaluate what they receive.

Multiple Evidences of the same subtype from different tools are allowed.
`Judge()` aggregates findings by subtype before evaluating each
QualityGateRule against the total.

**References:** Project by `ProjectID` (stored in the aggregate from
construction; cannot be modified after).

#### Editing CaseFile

- **Invariants to preserve**: `ProjectID` non-zero; `commitSHA` non-empty;
  `branch` non-empty; `startedAt` non-zero; **no AddEvidence after Judge**;
  **Judge is terminal** (calling twice returns `ErrAlreadyJudged` —
  judged is the final state of the aggregate lifecycle).
- **Events to keep emitting**: `EvidenceCollected` per AddEvidence,
  `VerdictRendered` once on Judge, `QualityGateFailed` only when outcome
  is `fail`.
- **Tests to re-run / extend**: `core/domain/casefile/model/*_test.go`
  for lifecycle (cannot-add-evidence-after-judge, terminal judge).
  Sub-model: `core/domain/casefile/model/evidence/*_test.go` for Evidence /
  Finding / Content invariants. Application: every handler under
  `core/application/casefile/`.
- **File map**: `core/domain/casefile/model/{casefile.go, casefile_id.go, tracking.go, ruling.go, verdict.go, events.go, errors.go}`
  + `core/domain/casefile/model/evidence/*.go`
  + `core/domain/casefile/service/repository.go`.

---

## Entities

### Evidence

A piece of evidence produced by a specific tool within a CaseFile.

**Identity:** `EvidenceID` value object (UUID generated in domain). Unique
within its CaseFile boundary.

**Invariants:**
- `subtype` must be a known EvidenceSubtype
- `source` (tool name) not empty
- `content` must match the subtype (see EvidenceContent below)
- belongs to exactly one CaseFile

**Attributes:**
- `type` — evidence category, **derived automatically** from subtype
  (not an input — `subtype.Type()` returns it). Kept for readability and
  UI grouping but never set directly.
- `subtype` — specific analysis kind (see EvidenceSubtype below)
- `source` — tool that produced it (e.g. "spotbugs", "lcov", "trivy")
- `content` — typed payload (see EvidenceContent below)
- `collectedAt` — timestamp

Evidence is pure data — what the tool produced. It carries no evaluation
result. The same Evidence can be evaluated by different quality gates (CLI
local vs server authoritative) producing different Verdicts.

### EvidenceContent (interface)

Typed payload for Evidence. Each subtype category has its own content type.
The Evidence constructor validates that subtype and content are compatible.

| Content type | Used by subtypes | What it carries |
|---|---|---|
| `FindingsContent` | code_quality, complexity, sast, secrets, malware, dast, sca | `findings []Finding` |
| `CoverageContent` | coverage | `totalLines`, `coveredLines`, `byLanguage []LanguageCoverage` |
| `NewCodeCoverageContent` | new_code_coverage | `coveredLines`, `coverableLines` (new/changed code only) |
| `LicenseContent` | license | `dependencies []DependencyLicense` |
| `ArchitectureContent` | architecture | `violations []ArchitectureViolation` |

Each content type is a value object with its own invariants and a `Merge()`
method for aggregating multiple Evidences of the same subtype during
`Judge()` evaluation.

---

## Value Objects

### Identity types

Vernon (IDDD ch. 5, *"Implementing Unique Identity"*): IDs are value objects,
never raw strings. This documents intent, makes the model express domain
language, and gives compile-time safety (impossible to pass a `CaseFileID`
where a `ProjectID` is expected).

| Type | Package | Format | Construction |
|------|---------|--------|--------------|
| `ProjectID` | `project/model` | UUID | `NewProjectID(string)` (rehydrate) / `GenerateProjectID()` (fresh) |
| `CaseFileID` | `casefile/model` | UUID | `NewCaseFileID(string)` / `GenerateCaseFileID()` |
| `EvidenceID` | `casefile/model/evidence` | UUID | `NewEvidenceID(string)` / `GenerateEvidenceID()` |
| `GavelspaceID` | `gavelspace/model` | non-empty string | `NewGavelspaceID(string)` |

All identity VOs expose: `String() string`, `Equal(other) bool`, `IsZero() bool`.

Convention:
- `New*` constructors validate non-emptiness. Used when hydrating from a wire
  payload or persistence.
- `Generate*` produces a fresh UUID. Used by aggregate constructors.
- Domain rejects zero-value IDs (`IsZero()`) at aggregate boundaries.

### Finding

An individual issue located in source code, produced by a tool.

**Invariants:**
- `tool` not empty
- `ruleID` not empty
- `severity` valid
- `filePath` not empty
- `line` >= 0
- `fingerprint` not empty

**Attributes:**
- `tool` — analyzer that found it (e.g. "spotbugs", "pmd")
- `ruleID` — rule identifier (e.g. "NP_NULL_ON_SOME_PATH")
- `severity` — error / warning / note
- `filePath` — file where the issue was found
- `line` — line number
- `message` — human-readable description
- `fingerprint` — stable identity for tracking across CaseFiles

Findings are localizable in source code (file + line). They can be rendered
inline in an editor. This is what distinguishes them from metrics.

Finding is a context-free value object — it carries no status (new/existing/
resolved). The same Finding is "new" in one comparison and "existing" in
another. Status lives in TrackingResult.

### Fingerprint

Stable identity of a finding across CaseFiles. Enables tracking.

**Attributes:**
- `value` — opaque string (typically `SHA-256(ruleID:filePath:startLine)`,
  overridable by the tool via SARIF `partialFingerprints`)

Two findings with the same fingerprint are considered the same finding across
different CaseFiles, regardless of message or severity changes.

### TrackingResult

Classification of findings by comparing current vs previous. Produced by
`ClassifyFindings()`, consumed by `Judge()` and the quality gate.

**Attributes:**
- `newFindings` — findings present now but absent in comparison baseline
- `existingFindings` — findings present in both current and baseline
- `resolvedCount` — count of fingerprints in baseline but absent now
  (no Finding objects — they no longer exist in current evidence)

TrackingResult is produced per comparison dimension. The server produces two:

| Comparison | Baseline | Purpose |
|---|---|---|
| Branch evolution | Previous CaseFile on same branch | Dashboard trends, alerts |
| PR delta | Latest CaseFile on target branch (Project.defaultBranch) | PR quality gate, annotations |

First analysis on a new branch: both comparisons fall back to the latest
CaseFile on the default branch.

### ClassifyFindings (domain function)

```
ClassifyFindings(current []Finding, previousFingerprints []Fingerprint) → TrackingResult
```

Pure domain function — no I/O. Compares fingerprints to classify findings.
The server's application layer orchestrates data fetching (previous
fingerprints from repository) and calls this function. The CLI does not
perform tracking — it collects and sends.

### RuleOutcome

Result of evaluating evidence against an EvaluationStrategy. Produced by
`Evaluate()`, consumed by `Judge()` to compose a Ruling.

**Attributes:**
- `passed` — whether the evidence met the threshold
- `detail` — human-readable explanation (e.g. "7 warnings (max 5)")

RuleOutcome is the raw evaluation result. `Judge()` enriches it with subtype
to produce a Ruling.

### Ruling

The judge's decision on a specific subtype within a CaseFile. One Ruling per
QualityGateRule evaluated. `Judge()` produces Rulings, then composes the
final Verdict from them.

**Attributes:**
- `subtype` — which evidence subtype was evaluated
- `passed` — whether the threshold was met
- `detail` — human-readable explanation (e.g. "7 warnings (max 5)")

### Verdict

Result of judging a CaseFile. Composition of rulings.

**Attributes:**
- `outcome` — pass / pass_with_warnings / fail
- `rulings` — list of Ruling (one per subtype evaluated)
- `evaluatedAt` — timestamp

**Outcome logic:**
- Any blocking ruling failed → **fail**
- No blocking failed, any warning ruling failed → **pass_with_warnings**
- Nothing failed → **pass**

### EvidenceType

Category that answers a fundamental question about the code. Used for
grouping in UI and dashboards. Not used for quality gate evaluation.

| Type | Question it answers |
|------|-------------------|
| `source_code` | How is the code written? |
| `security` | Is it secure? |
| `supply_chain` | What does it depend on? |
| `coverage` | Is it tested? |
| `architecture` | Does it follow the intended structure? |

### EvidenceSubtype

Specific kind of analysis within a type. This is what the quality gate
evaluates against — one rule per subtype.

| Type | Subtype | What it measures |
|------|---------|-----------------|
| `source_code` | `code_quality` | Code smells, best practices |
| `source_code` | `complexity` | Code complexity metrics |
| `security` | `sast` | Security vulnerabilities (static) |
| `security` | `secrets` | Hardcoded secrets/credentials |
| `security` | `malware` | Malware detection |
| `security` | `dast` | Security vulnerabilities (dynamic) |
| `supply_chain` | `sca` | Dependency vulnerabilities (CVEs) |
| `supply_chain` | `license` | Dependency license compliance |
| `coverage` | `coverage` | Test coverage percentage |
| `coverage` | `new_code_coverage` | Coverage on new/changed code only |
| `architecture` | `architecture` | Layer/dependency rule violations |

IaC findings are not a separate subtype. They are produced by tools and
classified as `source_code` (best practices) or `security` (vulnerabilities)
depending on the nature of the finding.

### Language

Identifies a programming language. Open set — not an enum, because custom
tools can report languages that Gavel does not have in its catalog.

**Invariants:**
- `name` not empty

**Attributes:**
- `name` — language identifier, normalized to lowercase (e.g. "java", "go",
  "typescript"). "Java" and "java" are the same Language.

Equality by `name`. Used in `Project.languages` and `LanguageCoverage.language`.

### LanguageCoverage

Coverage breakdown for a single language within CoverageContent.

**Invariants:**
- `language` valid Language
- `totalLines` >= 0
- `coveredLines` >= 0
- `coveredLines` <= `totalLines`

**Attributes:**
- `language` — Language value object
- `totalLines` — total lines for this language
- `coveredLines` — covered lines for this language

Percentage is derived (`coveredLines / totalLines`), not stored.

### DependencyLicense

A dependency and its license within LicenseContent. Evaluated by
ForbiddenList strategy.

**Invariants:**
- `name` not empty
- `version` not empty
- `license` not empty

**Attributes:**
- `name` — dependency identifier (e.g. "com.google.guava:guava")
- `version` — dependency version (e.g. "31.1-jre")
- `license` — SPDX license identifier or expression (e.g. "Apache-2.0",
  "MIT OR GPL-2.0"). ForbiddenList parses expressions to check against
  its forbidden set.

### ArchitectureViolation

A single violation of an architecture deny rule within ArchitectureContent.

**Invariants:**
- `rule` not empty
- `sourcePkg` not empty
- `message` not empty

**Attributes:**
- `rule` — name of the violated deny rule
- `sourcePkg` — package that imports the forbidden dependency
- `targetPkg` — forbidden dependency package (may be empty if rule is
  structural rather than import-based)
- `message` — human-readable violation description

### ProjectRef

A lightweight reference to a Project within a Gavelspace. Value object —
no independent identity or lifecycle.

**Invariants:**
- `id` valid ProjectID (non-zero)
- `targetPattern` non-empty

**Attributes:**
- `id` — `ProjectID` of the referenced project
- `targetPattern` — Bazel target pattern (used for uniqueness check in
  Gavelspace)

### Severity

Closed set: `error`, `warning`, `note`.

### Enforcement

Closed set: `blocking`, `warning`, `advisory`.

Determines how a QualityGateRule affects the final verdict:
- `blocking` — breaching this rule fails the verdict (red)
- `warning` — breaching this rule degrades the verdict to pass_with_warnings (yellow)
- `advisory` — reported only, no effect on verdict (neutral)

### QualityGate

Configurable per project. Defines how each evidence subtype is evaluated.

**Structure:**
- Collection of `QualityGateRule` (one per subtype enabled in the project)
- Verdict passes only when every rule passes

### QualityGateRule

Defines how to evaluate one evidence subtype.

**Attributes:**
- `subtype` — which evidence subtype this rule applies to
- `strategy` — evaluation strategy (carries its own threshold)
- `minResolved` — optional minimum items resolved per run (findings, violations)
- `minDelta` — optional minimum coverage improvement per run

### Evaluation strategies (domain knowledge)

The strategy defines HOW a subtype is evaluated. This is invariant — domain
knowledge, not configuration. The set of strategies is closed.
The threshold defines WHAT is acceptable. This is configurable per project.

Each strategy is a self-contained value object that carries its threshold and
knows how to evaluate. Implemented as the Strategy pattern: an
`EvaluationStrategy` interface with concrete implementations.

```
EvaluationStrategy interface {
    Evaluate(content EvidenceContent) RuleOutcome
}
```

Type safety at runtime is guaranteed by the `QualityGateRule` constructor,
which validates that the strategy is compatible with the subtype. By the
time `Judge()` calls `Evaluate()`, the content type always matches what the
strategy expects.

| Strategy | Applies to | Threshold | Example |
|----------|-----------|-----------|---------|
| CountBySeverity | sast, code_quality, dast, complexity, sca | max per severity | `{maxError: 0, maxWarning: 5, maxNote: 20}` |
| ZeroTolerance | secrets, malware | none (count must be 0) | `{}` |
| MinPercentage | coverage | minimum % | `{min: 80.0}` |
| MinNewCodeCoverage | new_code_coverage | minimum % on new code | `{min: 80.0}` |
| MaxViolations | architecture | max violation count | `{max: 0}` |
| ForbiddenList | license | forbidden items | `{forbidden: ["GPL-3.0", "AGPL"]}` |

Each strategy validates its own invariants in its constructor (e.g.
`CountBySeverity` rejects negative max values, `MinPercentage` rejects
values outside 0–100).

### Tool contract

A tool is any external analyzer that produces evidence. Gavel does not
implement analyzers — it orchestrates them and collects results.

The domain defines the contract: given an evidence subtype, a tool must
produce results that the quality gate can evaluate. For finding-based
subtypes, this means producing Finding value objects. For metric-based
subtypes (coverage), this means producing the relevant metrics.

Custom tools are supported: implement the contract, and Gavel treats them
the same as built-in tools.

---

## Domain Events

All events implement the `DomainEvent` interface defined in `core/domain/shared/event/`:

```go
type DomainEvent interface {
    EventName() string      // stable identifier, e.g. "casefile.verdict_rendered"
    OccurredAt() time.Time  // when the event was emitted (UTC)
}
```

Vernon (IDDD ch. 8): events are immutable past facts. The aggregate records
them in its private `events []event.DomainEvent` slice. The application service
drains them via `Events()` + `ClearEvents()` after the unit of work completes,
then dispatches to subscribers (logging, webhook delivery, read model updates).

### Event catalog

| Event | Emitted by | EventName | Payload |
|-------|-----------|-----------|---------|
| `ProjectAdded` | Gavelspace | `gavelspace.project_added` | GavelspaceID, ProjectID, TargetPattern |
| `ProjectRemoved` | Gavelspace | `gavelspace.project_removed` | GavelspaceID, ProjectID |
| `QualityGateUpdated` | Project | `project.quality_gate_updated` | ProjectID |
| `LanguagesUpdated` | Project | `project.languages_updated` | ProjectID |
| `ArchitecturePolicyUpdated` | Project | `project.architecture_policy_updated` | ProjectID |
| `CaseFileOpened` | CaseFile | `casefile.opened` | CaseFileID, ProjectID, CommitSHA, Branch |
| `EvidenceCollected` | CaseFile | `casefile.evidence_collected` | CaseFileID, ProjectID, Subtype, Source |
| `VerdictRendered` | CaseFile | `casefile.verdict_rendered` | CaseFileID, ProjectID, Outcome |
| `QualityGateFailed` | CaseFile | `casefile.quality_gate_failed` | CaseFileID, ProjectID, FailingSubtypes |

Every event has `OccurredAt time.Time` (UTC, set by the aggregate at emission)
exposed via the `OccurredAt()` method.

---

## Relationships

```
Gavelspace (1) ──── manages ────► (N) Project
Project (1) ──── configures ────► (1) QualityGate
Project (1) ──── has history ──► (N) CaseFile (by reference)
CaseFile (1) ──── contains ────► (N) Evidence
Evidence (1) ──── has ────► (1) EvidenceContent
FindingsContent (1) ── contains ──► (N) Finding
QualityGate (1) ──── has ────► (N) QualityGateRule (one per subtype)
QualityGateRule (1) ──── uses ──► (1) EvaluationStrategy (carries threshold)
CaseFile.Judge(qualityGate, trackingResult) ──── produces ────► Verdict (with Rulings)
ClassifyFindings(current, previous) ──── produces ────► TrackingResult
```

---

## Tracking

Tracking determines whether a finding is new, existing, or resolved by
comparing fingerprints across CaseFiles. This is orthogonal to quality gate
evaluation (SonarQube architecture insight: tracking answers "is this new?"
while baseline answers "what does the gate evaluate?").

### Dual comparison, zero configuration

The server performs two comparisons simultaneously:

| Comparison | Baseline | Question | Used for |
|---|---|---|---|
| Branch evolution | Previous CaseFile on same branch | Is this branch improving? | Dashboard, trends |
| PR delta | Latest CaseFile on Project.defaultBranch | What does this PR introduce? | Quality gate, PR annotations |

First analysis on a new branch: both fall back to the latest CaseFile on
the default branch (main).

### Finding classification matrix

| vs same branch | vs target branch | Meaning |
|---|---|---|
| new | new | You just introduced it — not in main, not in your branch before |
| new | existing | Inherited from main — appeared in your branch but was already in main |
| existing | existing | Known debt — was everywhere already |
| existing | new | Introduced in a previous commit on your branch — still not in main, pending fix |

### Architecture

- **Algorithm:** Pure domain function `ClassifyFindings()` in `casefile/`. No I/O.
- **Orchestration:** Server application layer fetches previous fingerprints
  from repository, calls `ClassifyFindings()`.
- **CLI:** Does not track. Collects evidence and sends to server.
- **Result:** `TrackingResult` value object. Finding stays status-free.

### Baseline evolution

When the target branch (main) receives new analyses, the PR delta comparison
reflects the current state — "what will be new if you merge NOW". This is
the industry standard (SonarQube Reference Branch behavior).

---

## Use case flow (judge)

```
1. CLI collects evidence for each configured tool
2. CLI sends evidence to server (CaseFile data + Evidence list)
3. Server creates CaseFile (commit, branch, now) for a Project
4. For each Evidence received:
   a. Server calls caseFile.AddEvidence(evidence)
   b. CaseFile emits EvidenceCollected event
5. Server fetches previous fingerprints (two baselines):
   a. Previous CaseFile on same branch
   b. Latest CaseFile on Project.defaultBranch
6. Server calls ClassifyFindings() for each baseline → two TrackingResults
7. Server loads the project's QualityGate
8. Server calls caseFile.Judge(qualityGate, prDeltaTrackingResult)
   a. Judge() groups Evidence by subtype (aggregating across tools)
   b. Judge() filters content using TrackingResult (only new findings)
   c. For each QualityGateRule: evaluates strategy against filtered content
   d. Produces one Ruling per rule (subtype + passed + detail)
   e. Composes rulings → final verdict (fail/pass)
   f. CaseFile emits VerdictRendered (and QualityGateFailed if fail)
9. Server stores branch evolution TrackingResult for dashboard/trends
10. Server returns Verdict to CLI
11. CLI may also evaluate locally with its own quality gate (from gavel.yaml)
    for immediate terminal feedback — calls Judge(qualityGate) without
    TrackingResult, evaluating all findings
```

---

## Package structure

Follows Vernon's approach: DDD tactical patterns with Hexagonal Architecture
for dependency inversion. Each aggregate package nests:

- `model/` — aggregate root, value objects, events, errors. Holds the whole
  domain model for the aggregate, including sub-models nested as
  `model/<concept>/` when they grow large (e.g. `model/evidence/` under
  `casefile/`, `model/qualitygate/` under `project/`).
- `service/` — domain service interfaces (ports) — present only when the
  aggregate needs persistence or external communication.

Each file contains one type. The file named after the package is the
aggregate root (by Go convention).

```
core/domain/
├── event/                       ← shared DomainEvent interface (Vernon ch. 8)
│   └── event.go                 ← interface { EventName() string; OccurredAt() time.Time }
├── failure/                     ← shared Kind enum + New (kinded sentinel constructor)
│   └── failure.go
├── casefile/                    ← analysis session (aggregate)
│   ├── model/
│   │   ├── casefile.go          ← aggregate root (CaseFile)
│   │   ├── casefile_id.go       ← value object (CaseFileID, typed UUID)
│   │   ├── tracking.go          ← value object (TrackingResult) + ClassifyFindings()
│   │   ├── ruling.go            ← value object (Ruling)
│   │   ├── verdict.go           ← value object (Verdict)
│   │   ├── events.go            ← EvidenceCollected, VerdictRendered, QualityGateFailed
│   │   ├── errors.go            ← domain errors
│   │   └── evidence/            ← sub-model: child entity + supporting VOs
│   │       ├── evidence.go          ← entity (Evidence)
│   │       ├── evidence_id.go       ← value object (EvidenceID, typed UUID)
│   │       ├── evidence_content.go  ← interface (EvidenceContent) + FindingsContent, CoverageContent, LicenseContent, NewCodeCoverageContent
│   │       ├── architecture_content.go  ← value object (ArchitectureContent)
│   │       ├── architecture_violation.go← value object (ArchitectureViolation)
│   │       ├── new_code_coverage_content.go ← value object (NewCodeCoverageContent)
│   │       ├── evidence_type.go     ← value object (EvidenceType)
│   │       ├── evidence_subtype.go  ← value object (EvidenceSubtype, owns Type() method)
│   │       ├── finding.go           ← value object (Finding)
│   │       ├── fingerprint.go       ← value object (Fingerprint)
│   │       ├── severity.go          ← value object (Severity)
│   │       ├── language_coverage.go ← value object (LanguageCoverage)
│   │       ├── dependency_license.go← value object (DependencyLicense)
│   │       ├── language.go          ← value object (Language)
│   │       └── errors.go            ← domain errors
│   └── service/
│       └── repository.go        ← CaseFileRepository (Save, FindByID, FindByProject, FindLatestByBranch, FindFingerprintsByBranch)
├── project/                     ← project configuration (aggregate)
│   ├── model/
│   │   ├── project.go           ← aggregate root (Project)
│   │   ├── project_id.go        ← value object (ProjectID, typed UUID)
│   │   ├── events.go            ← QualityGateUpdated, LanguagesUpdated, ArchitecturePolicyUpdated
│   │   ├── errors.go            ← domain errors
│   │   ├── archpolicy/          ← sub-model: architecture policy (optional)
│   │   │   ├── architecture_policy.go ← value object (ArchitecturePolicy)
│   │   │   ├── layer.go              ← value object (Layer)
│   │   │   ├── deny_rule.go          ← value object (DenyRule)
│   │   │   └── errors.go             ← domain errors
│   │   └── qualitygate/         ← sub-model: VO family owned by Project
│   │       ├── quality_gate.go      ← value object (QualityGate)
│   │       ├── quality_gate_rule.go ← value object (QualityGateRule)
│   │       ├── strategy.go          ← interface (EvaluationStrategy)
│   │       ├── strategy_count.go    ← value object (CountBySeverity)
│   │       ├── strategy_zero.go     ← value object (ZeroTolerance)
│   │       ├── strategy_percent.go  ← value object (MinPercentage)
│   │       ├── strategy_forbidden.go← value object (ForbiddenList)
│   │       ├── strategy_max_violations.go  ← value object (MaxViolations)
│   │       ├── strategy_new_code_coverage.go ← value object (MinNewCodeCoverage)
│   │       ├── rule_outcome.go      ← value object (RuleOutcome)
│   │       └── errors.go            ← domain errors
│   └── service/
│       └── repository.go        ← ProjectRepository (Save, FindByID, FindByName)
└── gavelspace/                  ← monorepo container (aggregate)
    ├── model/
    │   ├── gavelspace.go        ← aggregate root (Gavelspace)
    │   ├── project_ref.go      ← value object (ProjectRef)
    │   ├── gavelspace_name.go   ← value object (GavelspaceID)
    │   ├── events.go            ← ProjectAdded, ProjectRemoved (implement DomainEvent)
    │   └── errors.go            ← domain errors
    └── service/
        └── repository.go        ← GavelspaceRepository (Save, FindByName)
```

**Sub-package convention:**

- `model/` — the whole domain model for the aggregate: root, immediate
  VOs, events, errors. Sub-models nest as `model/<concept>/` when they
  grow large enough to deserve their own folder (e.g. `model/evidence/`,
  `model/qualitygate/`).
- `service/` — domain service interfaces (ports). Infrastructure adapters
  implement these. Present in aggregate packages that require persistence
  or external communication.

**Dependencies (unidirectional, no cycles):**

```
event/                              → (stdlib only)
failure/                            → (stdlib only)
casefile/model/evidence/            → (none from core)
project/model/qualitygate/          → casefile/model/evidence/
project/model/                      → casefile/model/evidence/ + project/model/qualitygate/ + event/
project/service/                    → project/model/
gavelspace/model/                   → project/model/ (for ProjectID) + event/
gavelspace/service/                 → gavelspace/model/
casefile/model/                     → casefile/model/evidence/ + project/model/qualitygate/ + project/model/ (for ProjectID) + event/
casefile/service/                   → casefile/model/ + casefile/model/evidence/ + project/model/
```

Notes:
- `casefile/model/ → project/model/` is a Vernon "reference by identity"
  edge: `CaseFile` stores a `ProjectID` field but never holds a `Project`
  reference. Same pattern between `gavelspace/model/` and `project/model/`.
- `project/qualitygate/ → casefile/evidence/` is a value-flow edge: a
  strategy needs `EvidenceContent` to evaluate. It is not an aggregate
  reference — evidence types cross as VOs.
- The graph is acyclic.

**Why `evidence/` lives under `casefile/model/`:**

Evidence is a child entity of CaseFile. It has identity (`EvidenceID`) and
its lifecycle is bound to the CaseFile that contains it: an unattached
Evidence has no persistent home and no meaning. Placing it inside
`casefile/model/evidence/` makes that ownership explicit at the package
level — the sub-model lives where the rest of the CaseFile model lives.

CaseFile does not need access to Evidence's unexported fields — its
constructor validates its own invariants and CaseFile only consumes
exported methods. Nesting under `model/` keeps the root's file count
bounded while keeping Evidence inside the aggregate's boundary.

The ingest use cases (`ingest/findings`, `ingest/coverage`) construct
Evidence before any CaseFile exists; this matches Vernon's pattern of
external value-object construction handed to an aggregate (e.g. his
`RouteSpecification` for Cargo) — the package layout follows ownership,
not construction site.

**Why `qualitygate/` lives under `project/model/`:**

QualityGate is a value object of Project, replaced wholesale via
`UpdateQualityGate`. It is owned by Project and has no independent
identity or persistence. `project/model/qualitygate/` mirrors the same
pattern as `casefile/model/evidence/`: ownership reflected in the package
tree, sub-model under `model/`.

**Aggregate boundaries:**
- `gavelspace/` — manages project collection, enforces target pattern uniqueness.
- `project/` — owns quality gate config, languages, defaultBranch. References Gavelspace by name.
- `casefile/` — analysis session lifecycle. Orchestrates judging using
  evidence/ (data) and qualitygate/ (rules). References Project by ID.

---

## Construction patterns

All aggregates, entities, and value objects follow two construction paths.
Exact signatures live in the code; this section documents the discipline.
Grep `^func New` and `^func Reconstitute` under `core/domain/` for the
authoritative list.

### `NewXxx()` — creation

Used when the domain creates a new instance. Generates identity via the
typed ID's `Generate*` constructor (UUID), validates all invariants, and
may emit domain events. Returns `(T, error)`.

Called by application services (use cases).

### `ReconstituteXxx()` — reconstitution from persistence

Used when rebuilding an instance from stored data. Receives the existing
typed identity instead of generating one. Validates the same invariants as
`NewXxx()` — persisted data is not blindly trusted. Does not emit domain
events (the creation already happened). Returns `(T, error)`.

Called by repository implementations in the infrastructure layer when
materializing from DB rows.

### Differences between the two paths

| Concern | `NewXxx()` | `ReconstituteXxx()` |
|---|---|---|
| Identity | Generates (UUID) | Receives existing |
| Invariant validation | Yes | Yes |
| Domain events | May emit | Never emits |
| Called by | Application layer (use cases) | Infrastructure layer (repositories) |

Both return `(T, error)`. If invariants fail on reconstitution, it signals
data corruption — fail fast rather than propagate an invalid aggregate.

Value objects that have no identity (Finding, Severity, Fingerprint, etc.)
use `NewXxx()` for both paths since there is no identity to distinguish. The
`Reconstitute` pattern applies to types with identity: aggregates (CaseFile,
Project, Gavelspace) and entities (Evidence).

### Why typed IDs in constructors

Vernon (IDDD ch. 5, §"Implementing Unique Identity") argues against primitive
obsession: identity types as VOs prevent passing the wrong ID where a method
expects a different one. The compiler enforces:

```go
cf, err := NewCaseFile(projectID, commitSHA, ...) // OK
cf, err := NewCaseFile(someStringID, commitSHA, ...) // COMPILE ERROR
cf, err := NewCaseFile(caseFileID, commitSHA, ...) // COMPILE ERROR — type mismatch
```

Repository ports also speak in typed IDs:

```go
type CaseFileRepository interface {
    FindByID(ctx context.Context, id CaseFileID) (CaseFile, error)
    FindByProject(ctx context.Context, projectID ProjectID) ([]CaseFile, error)
    ...
}
```

---

## Known limitations

- **Tracking across file renames:** When a file is renamed, the fingerprint
  changes (includes file path). The old finding appears as "resolved" and the
  same finding at the new path appears as "new". This is a false positive in
  tracking — the finding didn't actually get resolved and reintroduced.
  SonarQube handles this with moved-block detection, added years after their
  initial release. Gavel accepts this limitation for now. The design does not
  preclude adding rename detection later: fingerprints are overridable (SARIF
  tools can provide their own), and `ClassifyFindings()` could accept a rename
  map in the future.
