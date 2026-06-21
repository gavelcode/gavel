---
title: Web and server alignment
type: explanation
description: Decision record for aligning the web frontend against core server DTOs and query ports.
---

# Web-Server Alignment — Decision Record

Executed May 2026. Aligned the web frontend (`apps/web/`) against the core
server DTOs and query ports in 5 phases.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| IDs | UUID strings everywhere | Core uses typed UUIDs (`ProjectID`, `CaseFileID`). |
| Severities | SARIF: `error` / `warning` / `note` | Core uses SARIF terminology. SonarQube naming dropped. |
| Quality Gates | Embedded in Project (core model) | Not standalone entities. Standalone entity and page deleted. |
| API naming | `/api/casefiles` (domain term) | "Case file" is the judicial metaphor. "Judgment" removed. |
| Orphan pages | Keep with inline dummy data | Pages for unimplemented server features stay visible with demo data. |
| Project routing | By key (`/projects/:key`) not ID | Bazel pattern is the natural identifier. URL-encoded. |
| Finding identity | Fingerprint, not numeric ID | Core has no stable finding ID. Fingerprint is the dedup key. |

## Contracts aligned to

### CaseFile (was "Judgment")

```
CaseFileSummary:
  ID, ProjectID, CommitSHA, Branch, StartedAt, VerdictOutcome,
  TotalFindings, NewFindings, ExistingFindings, ResolvedFindings, CreatedAt

CaseFileDetail (extends Summary):
  Evidences []EvidenceSummary
  Rulings   []RulingView
```

### Finding

```
FindingView:
  Tool, RuleID, Severity, FilePath, Line, Message, Fingerprint, Status, Source

Filters: ProjectID, CaseFileID, Tool, Severity, Status, FilePath
```

### Project

```
ProjectSummary:
  ID, Key, Name, DefaultBranch, LatestVerdict, TotalFindings, CreatedAt

ProjectDetail (extends Summary):
  TargetPattern, Languages, QualityGateRules[], SeverityCounts map[string]int
```

### API routes (final state)

| Endpoint | Notes |
|----------|-------|
| `GET /api/projects` | Paginated list |
| `GET /api/projects/{key}` | Detail by URL-encoded Bazel key |
| `POST /api/projects` | Returns `{projectId}` |
| `PUT /api/projects/{key}/quality-gate` | Update rules |
| `PUT /api/projects/{key}/languages` | Update languages |
| `GET /api/casefiles?project_id=` | Paginated list |
| `GET /api/casefiles/{id}` | Detail |
| `GET /api/findings?casefile_id=&severity=&tool=&file_path=` | Flat with query params |
