---
title: Spec design
type: explanation
description: How to write specification documents that LLMs can follow reliably.
---

# Spec Design

How to write specification documents that LLMs can follow reliably.
Based on research from Zietsman 2026, Monperrus 2026, Osmani 2026,
and industry experience with spec-driven AI development.

---

## The problem with vague specs

LLMs are pattern completers, not mind readers. A vague spec forces the
model to guess unstated requirements — and its guesses come from training
data, not from your project's intent. The result: plausible-looking code
that violates conventions you never wrote down.

> "The only difference between a spec that is ignored and one that works
> is precision. Vague instructions for an LLM are like vague tickets for
> a junior developer — both hallucinate an interpretation."

Compliance correlates inversely with ambiguity. When a spec permits
multiple interpretations, the model selects unpredictably. Clear
constraints yield predictable behavior.

## The spec as ground truth

The specification — not the implementation — is the stable artifact of
record. Improving a codebase means improving its specification; the
implementation is regenerable at any time (Monperrus 2026).

This means specs must be:
- **Auditable** — a human can read and validate them in one sitting.
- **Behaviorally complete** — every capability the code exercises is
  described in the spec.
- **Independent of implementation** — they describe what, not how.
- **Convergent** — two independent implementations of the same spec
  should produce consistent observable behavior.

## Rules for writing specs

### R1: One concern per document

Each spec covers one topic. `05-go-style.md` covers Go idioms;
`08-security.md` covers security. Not "Go style and security."

Mixing concerns creates contradictions. Integration test rules contradict
unit test rules by design (one uses real systems, the other doesn't). If
they live in the same document, the model has to reconcile them — and it
will reconcile them wrong.

**Test**: if you cannot describe the document's scope in one sentence
without using "and," split it.

### R2: Every rule must be falsifiable

A rule that cannot be violated is not a rule — it is decoration.

> Bad: "Write clean code."
> Good: "Never discard errors with `_` (except documented no-error APIs)."

The good rule has a clear violation signal: if `_, err :=` appears and
`err` is unused, the rule is broken. The bad rule has no observable
criterion for failure.

**Test**: can you grep or read code and determine pass/fail? If not,
the rule is not falsifiable.

### R3: Use directive language, not suggestions

When a spec says "prefer," "try to," or "consider," the model treats it
as optional. It may or may not follow through.

> Bad: "Prefer returning errors over panicking."
> Good: "No `panic` in library or domain code."

Use "always," "never," or state the constraint as a fact. Reserve "prefer"
for genuinely optional stylistic choices where either answer is acceptable.

**Test**: replace your verb with "must" or "must not." If the sentence
still expresses your intent, use the directive form. If "must" feels
too strong, the rule is probably a suggestion — either strengthen it or
drop it.

### R4: Show, don't describe

One code example beats three paragraphs of explanation. LLMs learn from
pattern matching — a concrete example is a stronger signal than an
abstract description.

```go
// Wrong — description only
"Wrap errors with context using fmt.Errorf and the %w verb."

// Right — description + example
"Wrap errors with context:
  return fmt.Errorf("save analysis %s: %w", id, err)"
```

Include both the pattern (what to do) and the anti-pattern (what not to
do) when the distinction is non-obvious. Label them clearly.

### R5: Make the scope explicit

Every rule should make clear where it applies. "All code," "domain layer
only," "tests only," "production code only."

> Bad: "Use `context.Background()` sparingly."
> Good: "`context.Background()` only in `main()` and tests. No
> `context.TODO()` in production code."

Without scope, the model applies the rule everywhere — or nowhere.

### R6: State the why, not just the what

A rule without justification feels arbitrary. LLMs (and humans) are more
likely to follow rules they can reason about.

> Bad: "No JWT for sessions."
> Good: "No JWT for sessions — opaque tokens allow server-side revocation.
> JWT revocation requires a deny list that defeats the purpose."

The why also helps the model make correct judgment calls in edge cases
the spec didn't anticipate.

### R7: Keep it under reading distance

A spec that exceeds 1,500 words or 15 minutes of reading time loses
compliance. LLMs attend to ~150 instructions reliably; beyond that,
later rules get lower attention.

If a spec is growing long:
- Split by concern (R1).
- Move code examples to a linked reference.
- Eliminate redundancy — if the same rule appears in two places, one
  of them should be a reference, not a copy.

### R8: Separate rules that contradict by context

Some rules are only valid within a specific context and conflict with
rules in another context. Common examples:

- Unit tests vs. integration tests (mocks vs. real systems).
- Domain layer vs. infrastructure layer (no imports vs. import everything).
- Production code vs. test code (no inner classes vs. inner classes allowed).

If conflicting rules live in the same document, the model resolves the
contradiction unpredictably. Separate them into context-specific sections
or documents.

### R9: Name what "done" looks like

Every spec should make it obvious what code that follows the spec looks
like. If a reader has to interpret or weigh trade-offs to decide compliance,
the spec is too vague.

> Bad: "Good separation of concerns in the CLI."
> Good: "CLI commands parse flags, call core handlers, format output.
> If a command calculates a verdict or classifies a finding, it is a
> domain leak."

The "done" criterion is what connects specs to reviews. Each review
question should trace to a spec; each spec should be verifiable by a
review question.

### R10: Version and maintain

Specs are living documents. When the code changes, the spec must change.
A stale spec is worse than no spec — it teaches the model to follow
rules that no longer apply.

- Update the spec in the same commit as the code change.
- If a spec and the code disagree, the spec is wrong until proven
  otherwise. Fix the spec, then fix the code.

## Spec structure template

```markdown
# [Topic Name]

[One paragraph explaining what this spec covers and why it exists.]

## [Section]

[Rules as short, falsifiable statements. Directive language.]

[Code example showing the right pattern.]

[Code example showing the wrong pattern, labeled clearly.]

## [Section]

...
```

Sections should follow a natural reading order: most important rules
first, edge cases last. Within a section, start with the rule, then
the rationale, then the example.

## Anti-patterns

### The encyclopedia

A 3,000-word document covering everything from naming to deployment.
The model reads the first half carefully and the second half at a skim.
Split by concern.

### The suggestion box

"Consider using..." / "It might be helpful to..." / "You could try..."
These are prompts for brainstorming, not constraints for code. The
model treats them as optional. Use directives.

### The orphan rule

A rule with no code example and no justification. The model has to
infer what the rule means and why it exists. It will infer wrong.
Always show and explain.

### The fossil

A rule that applied to a previous architecture but was never removed.
The model follows it faithfully, producing code that conflicts with
the current design. Audit specs when architecture changes.

### The universal rule

"Always do X in all contexts." If X has exceptions (and almost everything
does), state them. Otherwise the model applies the rule where it
shouldn't, and when corrected, loses confidence in the rule entirely.

### The implementation leak

A spec that describes how to implement something instead of what the
behavior should be. "Use a map with string keys" instead of "cache
results with O(1) lookup." Implementation specs are brittle — they
break when the implementation changes. Behavioral specs survive.

## Connecting specs to reviews

Specs define WHAT the code should be. Reviews verify that it does.
The connection is explicit:

1. Every review question traces to a spec section (R1 of
   `question-design.md`).
2. Every falsifiable rule in a spec (R2 above) should be verifiable by
   a review question.
3. If a spec has rules that no review question checks, those rules are
   unenforced — they exist on paper but have no guardrail.

This creates a closed loop: spec → code → review → spec update.

## Sources

- [The Specification as Quality Gate: Three Hypotheses on AI-Assisted Code Review](https://arxiv.org/abs/2603.25773) — Zietsman, March 2026. The residual taxonomy, correlated failures in LLM pipelines, and specification-first architecture.
- [Bootstrapping Coding Agents: The Specification Is the Program](https://arxiv.org/abs/2603.17399) — Monperrus, March 2026. Specs as the stable artifact; implementations as regenerable.
- [Spec-Driven Development: From Code to Contract](https://arxiv.org/html/2602.00180v1) — February 2026. Empirical evidence on spec quality → code quality correlation.
- [How to Write a Good Spec for AI Agents](https://addyosmani.com/blog/good-spec/) — Osmani, 2026. Practical rules: falsifiability, directive language, context budgeting.
- [Building Shared Coding Guidelines for AI (and People Too)](https://stackoverflow.blog/2026/03/26/coding-guidelines-for-ai-agents-and-people-too/) — Stack Overflow, March 2026. Making tacit knowledge explicit for both humans and agents.
- [How to Write Rules for AI Coding Tools](https://virtuslab.com/blog/ai/how-to-write-rules-for-ai) — VirtusLab, 2026. Rule structure, anti-patterns, context-specific rule separation.
- [Spec-Driven Development: Unpacking 2025's Key AI-Assisted Engineering Practice](https://www.thoughtworks.com/en-us/insights/blog/agile-engineering-practices/spec-driven-development-unpacking-2025-new-engineering-practices) — Thoughtworks, 2025. Spec structure, domain language, quality gates.
