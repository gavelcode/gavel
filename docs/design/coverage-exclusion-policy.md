---
title: Coverage exclusion policy
type: explanation
description: Why Gavel measures coverage as a faithful mirror of the tools — no per-line pragma, no editing of the denominator — and where legitimate exclusions belong instead.
---

# Coverage Exclusion Policy

## Principle

> **Gavel reflects what the tools measure; it neither adds nor removes
> exclusions of its own.**
>
> Gavel does not edit, does not embellish, does not fight the project's own
> configuration.

Coverage is reported exactly as the underlying tool produced it. Gavel is a
normalizer and a gate, not a second opinion on what "should" count.

## Decision

| Decision | Choice |
|----------|--------|
| Per-line coverage pragma (`// gavel:unreachable`) | **Will not build** |
| Honoring / stripping the tools' own pragmas | **Neither** — pass through faithfully |
| Sanctioned way to exclude code | **Project scope** in `gavel.yaml` (file/dir/target), proposed separately |
| Existing debt | Protected by the **baseline**, never by editing the number |

## Context

Some lines are unreachable by design — defensive guards the domain already
guarantees can't fire (`if err != nil` over a factory that cannot fail,
re-validation in `Reconstitute*`). They never execute, so they sit as
uncovered and cap a file below 100%. The recurring temptation is to let authors
mark such lines so they drop out of the denominator and a clean "100%" becomes
reachable.

We considered three positions:

- **(a) Faithful mirror.** Gavel reports exactly what the coverage tool emits.
  Zero machinery; the exclusion decision (if any) lives in the project's own
  tool config, which is itself visible in the repo.
- **(b) Active honesty.** Gavel configures every upstream tool to ignore
  exclusions so the denominator is always complete. Defeats gaming, but costs
  per-language plumbing in the analysis aspects.
- **(c) Gavel's own per-line pragma.** A `// gavel:unreachable` comment that
  Gavel scans for and subtracts from the denominator. Most machinery, and it
  reintroduces exactly the escape hatch we want to avoid.

**We choose (a).**

## Rationale

**A per-line pragma is the wrong tool — the evidence is against it.**
Hora's empirical study of deliberate coverage exclusion across 55 projects
found the **single most common reason for excluding a line is that it was
simply "already untested" (22%)** — i.e. the pragma is used far more to hide
debt than to mark the genuinely unreachable. The reviewer's takeaway is
sharper still: the healthy move is to *design code so there is no temptation to
exclude*, not to normalize an inline escape hatch. In a Vernon-strict domain
that means asking whether a given defensive guard should exist at all, not
how to hide it from the report.

**Go itself deliberately declined this.** Go's coverage tool (`go test
-cover`) has no line-exclusion pragma, and the requests for one have stalled by
design: [#31280](https://github.com/golang/go/issues/31280) sits `Unplanned`,
and the concrete [#53271](https://github.com/golang/go/issues/53271) proposal
for a `//go:cover ignore` comment was frozen. The central objection in that
thread — a covered-but-ignored line is a silent lie — is the same one that
makes us refuse it. Go is Gavel's primary language, so for Go the mirror is
honest by construction: there is no upstream pragma to leak.

**"Honest" here means faithful reporting, not a guaranteed-full denominator.**
For Python/Java, a tool-level pragma (`# pragma: no cover`,
`@codeCoverageIgnore`) is applied before the LCOV reaches Gavel, so those lines
never appear and Gavel cannot see them. We accept that: it is the project's own
visible configuration, not a hidden Gavel mechanism, and the **baseline**
already prevents coverage from silently dropping run to run. We do **not**
claim every reachable line is counted; we claim Gavel adds no escape hatch of
its own and edits nothing.

## Consequences

- Files with genuinely-unreachable defensive guards stay **below 100%**, and
  that is accepted. A 97.7% on such a file is honest, not a defect to engineer
  away with a comment.
- Existing debt is handled where it already is: the **baseline** protects it;
  only *new* uncovered code moves the gate. We never lower the number to make a
  gate pass.
- The gate's `min: 90%` is read against whatever the tools report. For Go that
  is the full reachable denominator; for other languages it inherits the
  project's own tool config — a transparent, repo-visible decision.

## The one sanctioned exclusion: project scope (proposed)

Refusing a per-line pragma is **not** "no exclusions ever." Generated and
vendored code (OpenAPI `gen/`, protobuf, mocks) legitimately should not be
gated — for findings and architecture, not just coverage. The right shape is
**scope, not a coverage hack**: an `exclude` list next to a project's `pattern`
in `gavel.yaml`, applied to the whole gate.

```yaml
- name: core
  pattern: //core/...
  exclude:
    - //core/**/gen/...     # generated (OpenAPI, protobuf) — out of gate scope
```

This is visible, auditable in review, coarse (file/dir/target, never per-line),
and maps onto Bazel's native negative target patterns
(`//core/... -//core/gen/...`). It says "this is not my code to gate" rather
than "subtract this from the coverage denominator" — which is precisely the
distinction that keeps it out of the gaming category the research flags.

This mechanism is **implemented**: `projects[].exclude` in `gavel.yaml` (see
[configuration](../configuration.md)). Each exclude must resolve within the
project's `pattern`; excluded targets are dropped before analysis via Bazel
negative target patterns, so they affect no gate dimension.

## References

- A. Hora, *What Code Is Deliberately Excluded from Test Coverage and Why?*
  (MSR 2021) and *Excluding Code from Test Coverage: Practices, Motivations,
  and Impact* (EMSE 2023) —
  [paper](https://homepages.dcc.ufmg.br/~andrehora/pub/2023-emse-test-coverage-exclusion.pdf),
  [critical review](https://neverworkintheory.org/2021/09/01/what-code-is-deliberately-excluded-from-test-coverage-and-why.html).
- golang/go [#31280](https://github.com/golang/go/issues/31280) — exclude
  statically unreachable code (Unplanned).
- golang/go [#53271](https://github.com/golang/go/issues/53271) — proposal for
  a `//go:cover ignore` comment (frozen).
- [baseline-strategy.md](baseline-strategy.md),
  [incrementality-decision.md](incrementality-decision.md) — how existing debt
  is handled without editing the number.
