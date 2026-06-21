---
title: Application model
type: reference
description: The use cases — Simple CQRS commands and queries — that orchestrate the domain aggregates, per Vernon IDDD.
---

# Gavel Core — Application Model

Application Services orchestrate domain aggregates to fulfill user intent.
This document follows Vernon's *Implementing Domain-Driven Design* (IDDD)
strict interpretation, chapter 14 ("Application").

## Vernon discipline

Key rules applied throughout the application layer:

1. **Application Services as façades** — Commands and Results expose only
   primitives and value objects to outer layers. Domain aggregates never
   leak through. Identity is referenced by typed IDs.
2. **Application Services compose other Application Services** when one
   represents the canonical entry point for an operation. Coordination
   policy lives in the orchestrator, not in the composed handlers.
3. **One use case = one handler type** — focused on a single user intent.
4. **Dependencies through constructor injection** — handlers depend on
   repository interfaces (defined in `domain/`), never on concrete
   implementations.
5. **One aggregate per transaction** — when an operation touches two
   aggregates, the application service saves them sequentially, accepting
   the eventual consistency window.
6. **Two-level DTOs** — Application Commands/Results are pure Go types (no
   `json:` tags). The Published Language layer in `core/userinterface/api/v1/<bc>/ (per BC)` defines
   JSON-tagged DTOs and translates between the two.

## Use cases

Use cases live in `core/application/`, organized **per Bounded Context**
(Simple CQRS — commands AND queries together in the same BC tree). The
`shared/` and `supporting/` siblings carry the Vernon Shared Kernel and
Supporting Subdomain pieces respectively:

```
core/application/
├── casefile/
│   ├── judge/, submit/, classify/                # commands
│   ├── ingestfindings/, ingestcoverage/          # commands (parsers)
│   ├── baseline/, get/, list/, listfindings/     # queries
│   └── evidencedto/                              # BC-shared Evidence DTOs
├── gavelspace/
│   ├── create/, registerproject/, removeproject/ # commands
│   └── get/, list/                               # queries
├── iam/
│   ├── login/, changepassword/, createuser/, deactivateuser/,
│   │   issuetoken/, revoketoken/, listmytokens/, resolveprincipal/
│   └── tenant/{create,suspend}/
├── pleading/
│   ├── file/, resolve/                           # commands
│   ├── get/, list/                               # queries
│   └── gateview/                                 # BC-shared view DTOs
├── project/
│   ├── create/, updatequalitygate/, updatelanguages/, updatetargetpattern/
│   ├── get/, getbykey/, list/                    # queries
│   └── projectview/                              # BC-shared view DTOs
├── shared/
│   └── event/                                    # Shared Kernel: Event DTO
└── supporting/
    └── search/                                   # Supporting Subdomain
```

Cada query usa el patrón **Finder** (Evans/Vernon CQRS read side): un
`finder.go` con un `type Finder interface` de un solo método. La
implementación Postgres vive en `core/infrastructure/<bc>/postgres/`
(o `core/infrastructure/supporting/search/` para el search cross-BC).

### Configuration use cases (Gavelspace / Project)

| Path | Purpose | Aggregates touched |
|------|---------|---------------------|
| `project/create` + `gavelspace/registerproject` | Two-step ingest: create Project (tx 1) then register ProjectRef (tx 2) | Project, Gavelspace |
| `gavelspace/removeproject` | Remove ProjectRef from Gavelspace (Project entity preserved) | Gavelspace |
| `project/updatequalitygate` | Replace a Project's quality gate | Project |
| `project/updatelanguages` | Replace a Project's language list | Project |

### Analysis use cases (CaseFile)

| Path | Purpose | Notes |
|------|---------|-------|
| `casefile/ingestfindings` | Parse bytes via format parser → Evidence (findings-based subtype) | CLI-facing; uses `ingestfindings.Parser` interface |
| `casefile/ingestcoverage` | Parse bytes via format parser → Evidence (coverage subtype) | CLI-facing; uses `ingestcoverage.Parser` interface |
| `casefile/classify` | Compare current findings against a branch baseline → TrackingResult | Internal orchestration primitive |
| `casefile/judge` | Load CaseFile + load Project's QualityGate + Judge + Save | Canonical "evaluate this case file" |
| `casefile/submit` | End-to-end server flow: create CaseFile, add evidences, classify, judge, save | Composes `judge` + `classify` |

## Use case structure

Every use case follows the same skeleton (Vernon, IDDD §14 "Anatomy").
Commands and queries share the `Handler` + `Execute` pattern but differ in
their input type name:

```
# Command use case
core/application/<group>/<usecase>/
├── errors.go       ← ErrInvalidCommand (validation sentinel)
├── command.go      ← Command (primitives + VOs only)
├── command_test.go
├── result.go       ← Result (may be omitted if void)
├── handler.go      ← Handler struct, NewHandler(deps...), Execute(ctx, cmd)
└── handler_test.go (+ fakes_test.go for in-memory repositories)

# Query use case
core/application/<bc>/<aggregate>/<operation>/
├── query.go        ← Query struct + interface + Result/DTO types
├── handler.go      ← Handler struct, NewHandler(dep), Execute(ctx, query)
├── handler_test.go
└── fakes_test.go   ← in-memory interface implementation
```

Exported types are short and unqualified (`Command`/`Query`, `Result`,
`Handler`, `ErrInvalidCommand`/`ErrInvalidQuery`) because the package name
already conveys the use case at the call site (e.g. `registerproject.Command`,
`judge.Result`, `list.Query`).

### Command signatures

```go
// Configuration
create.NewCommand(key, name, targetPattern string)                                      // project/create
registerproject.NewCommand(gavelspaceID, projectID, targetPattern string)               // gavelspace/registerproject
removeproject.NewCommand(gavelspaceID, projectID string)
updatequalitygate.NewCommand(projectID string, input updatequalitygate.Input)
updatelanguages.NewCommand(projectID string, languages []string)

// Analysis (CaseFile BC)
ingestfindings.NewCommand(data []byte, format, source, subtype string)
ingestcoverage.NewCommand(data []byte, format, source string)
classify.NewCommand(projectID, branch string, findings []evidence.Finding)
judge.NewCommand(caseFileID string, tracking *casefile.TrackingResult)
submit.NewCommand(projectID, commitSHA, branch string, startedAt time.Time, evidences []evidencedto.Evidence)
```

Vernon-strict observations:
- All IDs cross the command boundary as strings. The handler reconstitutes
  typed IDs (`CaseFileID`, `ProjectID`) inside the domain layer.
- Quality gate and languages enter as application-layer input shapes
  (`updatequalitygate.Input`, `[]string`) — Vernon admits VOs across
  boundaries since they have no identity to protect, and the application
  maps them to domain VOs in the handler.
- Evidence enters `submit` as `[]evidencedto.Evidence` (the application
  Evidence DTO from `core/application/casefile/evidencedto/`). Evidence is
  a child entity of CaseFile, not an independent aggregate, so the crossing
  is permitted; the handler converts to domain `evidence.Evidence` internally.

### Result types

```go
create.Result{ ProjectID string }                                                       // project/create
judge.Result{ CaseFileID string; Verdict VerdictView; Events []event.Event }
ingestfindings.Result{ Evidence evidencedto.Evidence }
ingestcoverage.Result{ Evidence evidencedto.Evidence }
classify.Result{ Tracking casefile.TrackingResult }
submit.Result{ CaseFileID string; Verdict judge.VerdictView; Events []event.Event; EvidenceSummary }
```

Use cases that do not produce domain output beyond events (`removeproject`,
`updatequalitygate`, `updatelanguages`) return `(Result, error)` where
`Result` carries only `Events []event.Event` (from
`core/application/shared/event/`) — the application layer drains domain
events and translates them through `event.EventsFromDomain` into the
Shared Kernel DTO.

## Handler dependency injection

Every handler is constructed via `NewHandler(deps...)`. Dependencies are
domain service interfaces (ports), never concrete implementations:

```go
judge.NewHandler(
    caseFiles caseservice.CaseFileRepository,
    projects  projectservice.ProjectRepository,
) *Handler

submit.NewHandler(
    caseFiles       caseservice.CaseFileRepository,
    projects        projectservice.ProjectRepository,
    judgeHandler    *judge.Handler,    // composed application service
    classifyHandler *classify.Handler, // composed application service
) *Handler
```

Wiring lives at the deployment edge (`apps/server/cmd/gavel-server/main.go`,
`apps/cli/cmd/gavel/main.go`). Each binary instantiates SQLite / in-memory /
HTTP-client repository adapters and injects them into handlers.

## The submit flow

`casefile/submit` is the canonical server-side entry point. It implements
the use case flow described in `domain-model.md` (steps 3–10) by composing
smaller use cases:

```
1. Validate projectID + load Project (for DefaultBranch)
2. NewCaseFile(projectID, commitSHA, branch, startedAt)
3. AddEvidence loop (each emits EvidenceCollected)
4. caseFiles.Save(cf)              ← first persistence: collecting state
5. classify.Execute(projectID,     ← PR delta baseline against DefaultBranch
                    project.DefaultBranch(),
                    extractFindings(evidences))
6. judge.Execute(cf.ID(), &tracking) ← loads, judges, saves (second persistence)
7. Build EvidenceSummary from tracking + evidences
8. Return submit.Result{ CaseFileID, Verdict, Events, EvidenceSummary }
```

Two writes per submission (collecting → judged). Vernon-acceptable cost: the
two states are semantically distinct, and the intermediate save provides a
recovery point if judging fails.

### CLI–Server integration (Vernon context mapping)

CLI and Server share `core/` as **Shared Kernel**. In server mode, the CLI
acts as a driving adapter of the server's **Open Host Service**
(`POST /api/analyses`), not as an independent bounded context.

- **Server mode** (`--server URL`): CLI collects evidence (Bazel aspects →
  SARIF/LCOV) and delegates to the server via `SubmitAnalysis()`. The server
  runs the full pipeline (parse → classify → judge) and returns the
  authoritative verdict + `EvidenceSummary`. CLI displays the server's result.
  If the server is unreachable and `--require-submit=false` (default), CLI
  falls back to the local pipeline with a visible warning.
- **Local mode** (no `--server`): CLI collects evidence and runs the full
  core pipeline locally (submit → classify → judge) using in-memory repos.
  Baselines saved to `.gavel/baseline/`.

**Deferred** (not implemented yet): branch evolution baseline (paso 5a + 9 of
the doc). Requires a TrackingResult store for dashboards. Will be added when
the dashboard query handlers come online.

## Repository ports consumed

All ports live in `core/domain/<aggregate>/service/`:

| Port | Methods | Used by |
|------|---------|---------|
| `CaseFileRepository` | Save, FindByID, FindByProject, FindLatestByBranch, FindFingerprintsByBranch | `casefile/judge`, `casefile/classify`, `casefile/submit` |
| `ProjectRepository` | Save, FindByID, FindByName | `project/create` + `gavelspace/registerproject`, `project/updatequalitygate`, `project/updatelanguages`, `casefile/judge`, `casefile/submit` |
| `GavelspaceRepository` | Save, FindByName | `project/create` + `gavelspace/registerproject`, `gavelspace/removeproject` |
| `ingestfindings.Parser` | Parse([]byte) → []findings.Parsed | `casefile/ingestfindings` (format dispatch by map) |
| `ingestcoverage.Parser` | Parse([]byte) → coverage.Parsed | `casefile/ingestcoverage` |

The two `Parser` ports live in `core/application/casefile/{ingestfindings,ingestcoverage}/`
rather than `domain/`, because their input/output types (`findings.Parsed`,
`coverage.Parsed`) are application-layer DTOs that translate to domain types
inside the handler. The implementations (`core/infrastructure/casefile/sarif/`,
`core/infrastructure/casefile/lcov/`) depend on these ports.

## Eventual consistency between aggregates

Vernon (IDDD ch. 10): one aggregate per transaction. When a use case must
touch two aggregates, the order matters.

The ingest flow that creates a Project and links it to a Gavelspace runs
**two use cases sequentially**, each in its own transaction — there is no
combined `addproject`:

1. `project/create` saves the Project (transaction 1)
2. `gavelspace/registerproject` loads the Gavelspace, validates target
   pattern uniqueness in memory (`Gavelspace.AddProjectRef`), and saves
   the Gavelspace (transaction 2)

If step 2 fails, the Project exists in the store but is not referenced by
any Gavelspace — a benign orphan. The composition root retries
`registerproject` idempotently; reconciliation cleans up orphan Projects
offline.

A transactional Unit of Work spanning both aggregates is **rejected by
Vernon** as a violation of the small-aggregate principle: it conflates the
two consistency boundaries the aggregate design exists to keep apart. The
domain model accepts the brief inconsistency window in exchange for that
clarity.

## Errors and the Published Language

Every sentinel error in the system is declared with
[`core/domain/shared/failure.New(msg, kind)`](../domain/failure/failure.go). Each
sentinel carries its own `Kind` — `Validation`, `NotFound`, `Conflict`,
`Internal` — so no central registry has to enumerate them.

The classifier `failure.Of(err)` (in `core/domain/shared/failure`) walks the error chain with
`errors.As` and returns the first declared `Kind`. `shared.MapDomainError`
switches on that `Kind` to produce the right RFC 7807 Problem Details:

| Kind | HTTP | Problem type |
|------|------|--------------|
| `Validation` | 422 | `https://gavel.dev/problems/validation` |
| `NotFound` | 404 | `https://gavel.dev/problems/not-found` |
| `Conflict` | 409 | `https://gavel.dev/problems/conflict` |
| `Internal` (default) | 500 | `https://gavel.dev/problems/internal` |

Layering: `core/domain/shared/failure` holds `Kind` and the `New` constructor for kinded sentinels;
`core/domain/shared/failure` re-exports `Kind` via type alias plus the `New`
helper and the `Of` classifier. The userinterface layer imports only the
application package, so the "userinterface never imports domain" rule still
holds.

Adding a new sentinel is local to the package that emits it:

```go
// in core/domain/<aggregate>/errors.go
var ErrSomethingWrong = failure.New("something wrong", failure.Validation)
```

No central list to update, no mapping table to keep in sync.

## Testing approach

Each use case ships with:

- `command_test.go` — table-driven validation tests (each invariant rejected,
  each happy path constructed)
- `handler_test.go` — black-box tests (`package <usecase>_test`) with
  hand-rolled fakes
- `fakes_test.go` — in-memory implementations of the repository ports used
  by the handler

Zero mocks (project convention). Fakes implement the same domain interfaces as the
production adapters, with configurable error fields (`findErr`, `saveErr`)
to exercise failure paths. Tests verify behavior, not implementation.

## What is NOT in the application layer

- HTTP routing, multipart parsing, auth middleware → `apps/server/internal/core/userinterface/api/v1/<bc>/ (per BC)`
- SQL queries → `core/infrastructure/<bc>/postgres/`
- Bazel/git invocations → `apps/cli/internal/bazel/`
- Cobra command wiring → `apps/cli/internal/command/`
- JSON serialization → `core/userinterface/api/v1/dto/` (Published Language)
- Format-specific parsing (SARIF, LCOV) → `core/infrastructure/{sarif,lcov}/`

The application layer is pure orchestration: load → mutate → save → dispatch
events. Anything mechanical or technology-specific lives outside it.
