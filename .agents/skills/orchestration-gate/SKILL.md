---
name: orchestration-gate
description: "Shared preflight and postflight contract for skills that orchestrate sub-agents."
---

# Orchestration gate

Use this contract when a skill dispatches one or more sub-agents. The skill
owns the meaning of its work and roles; this gate owns the common model-plan,
cost, confidence, approval, and evidence vocabulary. A skill with no delegated
agents skips model selection and dispatch approval.

## Preflight

Complete preflight after understanding the request and before the first
sub-agent is dispatched. Present one dispatch summary containing the following
fields.

This is an executable workflow requirement, not explanatory prose: resolve the
declared roster through `internal/agents`, display every resolved assignment,
and record the user's approval before dispatch. A summary that does not show a
real role-to-model resolution, approved concurrency/budget constraints, and the
approval interaction is an incomplete preflight and must not dispatch.

### Request and confidence

- `request`: the user's requested outcome and explicit scope.
- `confidence`: an integer from 1 to 5, with the following fixed meanings:
  - `1 Guess`: the request or repository facts are substantially unknown.
  - `2 Fragile`: a plausible interpretation exists, but important assumptions
    or unresolved questions could change the work.
  - `3 Plausible`: the intent and approach are mostly understood, with material
    risks or verification gaps remaining.
  - `4 Strong`: the scope, approach, and relevant evidence are well supported;
    only limited risks remain.
  - `5 Verified`: the requirements and approach are directly supported by
    evidence and checked against the relevant repository facts.
- `confidence_rationale`: evidence supporting the score.
- `assumptions`: assumptions made in interpreting the request.
- `unresolved_questions`: decisions or missing information that could affect
  intent, scope, safety, or the dispatch plan.

The score is a skeptical state, not a probability. A score without evidence,
assumptions, and unresolved questions is invalid. Missing information that
changes the task's intent always requires clarification, regardless of score.

### Agent roster

Declare every planned role by name. Each role includes:

- `name`: stable role name used in the dispatch summary and handoffs.
- `purpose`: the bounded outcome the role is responsible for.
- `capability`: the capability needed, expressed using the repository's neutral
  agent/model vocabulary.
- `optional`: whether the role can be added after the initial approval.
- `reasoning`: the expected reasoning effort.
- `isolation`: the required workspace isolation.
- `concurrent`: whether the role may run concurrently with other named roles.
- `estimated_tokens`: a range, including the role's expected work and a note
  when the estimate is especially uncertain.

The roster is the complete planned dispatch scope. Optional roles and the
reserved review, fixes, and integration budget must be visible before approval.

### Model and cost plan

Offer exactly these plan choices when the roster is non-empty:

- `Recommended`: role-sensitive neutral profiles selected from the role's
  difficulty and risk.
- `Economy`: the least expensive reliable profiles and lower reasoning effort
  where the role permits it.
- `Deep`: stronger profiles or reasoning for difficult, cross-cutting, or
  high-risk roles.
- `Customize`: explicit profile and reasoning assignments for every named role.

For each role, show:

- `model_profile`: the neutral catalog profile (for example, fast, balanced,
  or capable) selected for the role.
- `resolved_model`: the harness model ID when resolution is available.
- `reasoning`: the approved reasoning effort.
- `concurrency`: the approved concurrency limit and the roles eligible to use
  it.
- `token_budget`: the approved token range, including reserved review,
  fixes, and integration capacity.

Resolve neutral profiles through the existing agent catalog at dispatch time.
An unavailable or missing resolution is a gate failure with an actionable
selection error. Never silently substitute another profile or model.

Keep these cost dimensions separate:

- `token_estimate`: the work estimate and approved token budget range.
- `account_quota`: expected account/quota impact, marked `measured`,
  `advisory`, or `unavailable`.
- `monetary_cost`: expected monetary cost, marked `measured`, `advisory`, or
  `unavailable`.

Estimates are ranges, not exact billing. An unavailable measurement must never
be presented as a zero or as an exact price.

### Approval gates

Apply the confidence gate before asking for cost-plan approval:

| Confidence | Required behavior |
| --- | --- |
| 1 Guess | Stop and request clarification. Do not dispatch. |
| 2 Fragile | Pause for clarification or an explicit user override. |
| 3 Plausible | Show risks and ask whether to proceed. |
| 4 Strong | Proceed only after the selected cost plan is approved. |
| 5 Verified | Proceed only after the selected cost plan is approved; retain the evidence supporting the score. |

Approval records the selected plan, role assignments, resolved models,
reasoning, concurrency, token budget, optional roles, and cost-status fields.
Do not dispatch until required clarification and approval are complete.

After approval, the plan is a constraint. A new approval is required before
any of the following changes:

- model profile or resolved model;
- reasoning effort;
- concurrency or isolation;
- token budget or reserved review/fix capacity;
- adding an optional agent or expanding review scope.

Never silently upgrade, fall back, add work, or exceed the approved plan.

## Postflight

After the primary work completes, produce a postflight report with all of the
following sections, even when a section is empty.

The report must be generated from the completed artifact and actual checks. It
must include concrete evidence pointers (tests, command results, or inspected
files), not only a restatement of this contract. If a skill cannot produce a
section, it must say that the evidence is unavailable and lower confidence.

Build the report from the original request, specification, or ticket rather
than from the primary agent's summary. For every requirement, record a
requirement identifier or precise pointer, the observed result, and exactly
one of `complete`, `partial`, `missing`, `ambiguous`, or `unrequested`.
Unrequested behavior is not a defect, but it must be visible so scope drift
can be distinguished from an intentional omission. A requirement cannot be
classified as `complete` from narrative plausibility alone: cite a check,
artifact, test, or source pointer that directly supports it.

- `confidence_and_rationale`: a new 1-to-5 score with evidence explaining the
  result, not a restatement of the preflight score.
- `verified_evidence`: checks actually performed, their results, and precise
  evidence pointers.
- `inferred_conclusions`: conclusions not directly verified, clearly marked as
  inference.
- `requirement_coverage`: each original requirement classified as `complete`,
  `partial`, `missing`, `ambiguous`, or `unrequested`.
- `assumptions`: assumptions still relied upon after the work.
- `low_confidence_areas`: unresolved or weakly evidenced parts of the result.
- `risks_and_likely_failure_modes`: concrete ways the result could fail or
  regress.
- `unnecessary_complexity`: abstractions, layers, or work that may be
  removable; required functionality must remain in scope.
- `red_team_findings`: findings only when an independently prompted red-team
  role was selected and run.
- `confidence_raising_checks`: the smallest follow-up checks most likely to
  increase confidence, ordered by expected value.

Keep `verified_evidence` and `inferred_conclusions` disjoint. Verified
evidence names the check that was actually performed, its result, and where
the result can be inspected. Inferred conclusions explain what that evidence
suggests but was not itself able to establish. Do not promote an inference to
verified evidence merely because the primary agent described it confidently.

Include assumptions and low-confidence areas even when the result is scored
4 or 5. For each material risk, state the likely failure mode and the
smallest check that would expose or reduce it. Prefer checks that inspect
requirements, execute the relevant behavior, or exercise an integration seam
over additional prose review.

### Complexity preservation check

Review the result for unnecessary abstractions, layers, duplication, and
ceremony. For every proposed simplification, name the requirement(s) it
preserves and the verification that proves they remain covered. A complexity
finding may recommend removal, but postflight must not remove or downgrade a
required behavior merely because the implementation can be made smaller. If
the simplification has not been applied, report it as a finding rather than
silently changing the artifact during postflight.

Use these postflight interpretations:

- scores `1` or `2`: report the result as incomplete or unsafe and identify
  blockers;
- score `3`: report material assumptions and request targeted follow-up;
- score `4`: report the limited risks that remain;
- score `5`: require direct requirement coverage and verification evidence,
  never merely a confident narrative.

The optional red-team review is recommended for broad, cross-cutting,
low-confidence, or implementation-producing work. It receives the relevant
requirements and artifact with minimal framing from the primary agent, reports
concrete counterexamples, and does not modify the implementation.

When selected, dispatch the red-team role with an independent prompt that
contains only the original requirements, the relevant artifact or diff, and
the expected output below. Do not include the primary agent's confidence,
interpretation, risk list, or proposed conclusions; those are framing that
can cause the reviewer to repeat the same assumptions.

The red-team report must identify concrete counterexamples or state that none
were found after naming the examined requirements. For each finding, include
the requirement pointer, failure scenario, evidence, severity, and the
smallest corrective check or change. The red-team agent is read-only: it may
not modify the artifact, broaden the requested scope, or silently become a
general reviewer. If the role is not selected, emit an empty
`red_team_findings` section rather than adding the review implicitly.

Carry the approved plan, confidence, assumptions, unresolved risks, and final
evidence into the skill handoff or final report so later agents do not need to
re-derive them.
