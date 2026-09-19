---
name: implement-spec
description: "Implement a specification in code."
---

You have been provided a spec. This spec should have tickets associated with it, describing how to implement the spec.

The goal is a coherent change which implements the entire spec. You own the
spec branch, worktrees, merges, review, and final handoff for the change.

Use the Branch and delivery invariants below as the single source of truth for
the integration base, worktree ancestry, cleanup, and delivery artifact.

The tickets are not a list of steps. They are a **task graph** with blocking relationships between them. This means there is always a **frontier** of tickets which are ready to be grabbed.

Communication to and from subagents should be sparse. Communicate primarily through **context pointers**: to the spec, tickets, research notes, and previous commits. Don't duplicate information already available via pointers.

Use the named agents **`implementer`**, **`merger`**, and **`explorer`**
where they fit the work. Before dispatch, check available usage if supported,
reserve roughly 25% of the available task budget for review, fixes, CI, and
integration, and apply the cost gate below. Keep token budgets, account quota,
and monetary cost separate; mark limits as advisory when the tool cannot
enforce them.

This skill also uses the shared `orchestration-gate` contract. The contract is
the source of truth for confidence vocabulary, approval gates, neutral model
profiles, cost-status fields, evidence reporting, and handoff context. This
skill supplies the implementation-specific roster and dispatch sequencing.

## Cost gate

Before dispatching any agent, choose a cost level. If the user has not chosen
one, use **Economy** and say so. Build a dispatch summary listing each agent,
model, reasoning effort, concurrency, and estimated token budget. Ask the user
before exceeding the selected level or adding an optional agent.

- **Economy**: choose the cheapest reliable model, use minimal or low reasoning,
  run at most two implementation tasks concurrently, and skip optional
  exploration unless a ticket is blocked without it.
- **Standard**: use low or medium reasoning, allow concurrency where ticket
  dependencies permit it, and reserve 25% of the budget for review and
  integration.
- **Deep**: use higher-cost models or reasoning, or broader concurrency, only
  after the user explicitly selects it or approves the exception.

Never silently upgrade a model, reasoning level, concurrency, or budget. If
usage or cost cannot be measured or enforced by the available tools, state that
limitation and enforce the selected level through the dispatch choices.

## Steps

1. Read the spec and tickets. Read enough to understand the task graph.

2. Complete orchestration-gate preflight before dispatching any agent. Build
   one user-visible dispatch summary containing:

   - the requested implementation scope, confidence score and rationale,
     assumptions, and unresolved questions;
   - the complete role roster, including the required `explorer`,
     `implementer`, and `merger` roles, plus `reviewer` and `red-team` only
     when the workflow selects them;
   - each role's purpose, capability, optionality, reasoning effort,
     isolation, concurrency eligibility, and token range;
   - the `Recommended`, `Economy`, `Deep`, and `Customize` plan choices;
   - each role's neutral model profile and harness-resolved model, approved
     reasoning, concurrency limit, and token budget; and
   - separate token-estimate, account-quota, and monetary-cost statuses,
     including the reserved review, fixes, CI, and integration capacity.

   Use the role-sensitive catalog resolver from `internal/agents/dispatch.go`
   to resolve every selected profile. An unavailable profile or harness
   resolution is an actionable gate failure. Do not substitute a different
   profile or model. Scores 1 and 2 stop or pause for clarification/override;
   score 3 exposes risks and asks whether to proceed; scores 4 and 5 still
   require approval of the selected cost plan. Missing information that could
   change intent always requires clarification.

   Approval is a binding constraint. Record the selected plan, role
   assignments, resolved models, reasoning, isolation, concurrency, token
   budget, optional roles, and cost-status fields in the implementation
   context. Do not dispatch until the required approval is present.

3. Apply the Branch and delivery invariants before dispatching any role. Resolve
   the spec directory and starting revision, create the spec branch and its
   integration worktree, and record the branch, immutable starting SHA, spec
   path, and worktree path in the implementation context. The spec branch must
   exist before any role receives a repository worktree.

4. (optional) Use the **`explorer`** agent to conduct any exploration required
   by the tickets - relevant codebase files or external documentation. Apply
   the selected cost level before dispatching it. If the explorer receives a
   repository worktree, create it from the current spec branch. It should save
   markdown notes in a directory outside the repo when possible, accessible by
   all future agents. Pass the approved assignment and the implementation
   context to it. This lets **`implementer`** focus on implementation rather
   than exploration. If exploration is not selected, omit the role from the
   roster and do not dispatch it.

5. Start one bounded implementation task per ready ticket. For each ticket,
   create its branch and worktree from the current spec branch, and give the
   implementer that isolated worktree. Set the task's model and reasoning
   effort explicitly when supported, and apply the approved isolation,
   concurrency, and budget limits. Use the named **`implementer`** when its
   fixed profile fits the ticket; otherwise select a suitable available model
   only if that selection was approved in preflight.

6. Once an implementation task completes, report its revision and validation to
   the main task, then use the named **`merger`** to inspect and merge it into
   the spec branch. Every handoff must carry the approved plan, preflight and
   postflight confidence, assumptions, unresolved risks, selected model and
   reasoning, isolation/concurrency/budget constraints, and evidence pointers.
   The merger may not broaden scope or change the approved dispatch plan.

7. If this changes the **frontier** of available tickets, report the newly ready
   tickets and dispatch them as dependencies and the remaining budget permit.

8. Once the primary implementation work completes, dispatch the selected
   `reviewer` role with its approved assignment and reserved review budget.
   If a red-team pass was selected, include a separately prompted `red-team`
   role in the original summary and dispatch it independently with the
   approved assignment. It receives the requirements and artifact with
   minimal framing, reports concrete counterexamples, and does not modify the
   implementation. If either role was not selected, do not add it implicitly.

   Run `/code-review` on the supplied workspace or revision when that is the
   selected reviewer workflow, then fix findings in one bounded implementation
   task using the reserved capacity. A new approval is required before adding
   any optional reviewer/red-team work, expanding review scope, exceeding the
   approved model/reasoning/isolation/concurrency/budget plan, or changing a
   resolved model. Never silently upgrade or fall back.

9. Produce the orchestration-gate postflight report, even when a section is
   empty. Compare every original requirement with the implementation and
   classify it as `complete`, `partial`, `missing`, `ambiguous`, or
   `unrequested`, with precise evidence pointers. Keep verified checks
   separate from inferred conclusions; report assumptions, low-confidence
   areas, concrete failure modes, unnecessary complexity, and the smallest
   confidence-raising checks. Complexity findings may not remove required
   behavior. When the red-team role ran, include its read-only findings from
   the independent minimal-context prompt; otherwise leave
   `red_team_findings` empty. Scores 1/2 are incomplete or unsafe, score 3
   requests targeted follow-up, score 4 reports remaining limited risks, and
   score 5 requires direct coverage and verification evidence.

10. After postflight, write the finished review handoff to the spec directory.
    Include the spec path, spec branch and final revision, requirement
    coverage, checks, review findings and resolutions, approved plan,
    confidence report, remaining risks or gates, and pointers to the relevant
    commits and artifacts. Use a stable descriptive filename such as
    `implementation-review-handoff.md`; preserve an existing handoff by
    updating it only when it is clearly the handoff for this run.

11. Clean up completed implementer worktrees and branches according to the
    Branch and delivery invariants. Return the final revision, checks, approved
    plan, handoff path, handoff evidence, confidence report, and remaining
    gates. Keep token estimates, account quota, and monetary cost distinct;
    unavailable measurements must remain unavailable or advisory, never zero
    or exact.

## Branch and delivery invariants

- The spec branch is created before any ticket branch or implementation
  worktree and is based on the spec run's starting revision.
- The starting revision is recorded as an immutable commit SHA before branch
  creation, using the documented resolution order.
- No role receives a repository worktree before the spec branch exists.
- Every implementation branch and worktree is based on the latest spec branch,
  not directly on the original base branch or another ticket branch.
- The merger integrates completed ticket work into the spec branch before the
  next dependent ticket is dispatched.
- Delivery is the completed review handoff in the spec directory, together
  with the retained spec branch and final revision. The workflow does not
  create, draft, update, or mark any pull request ready.

## Dispatch invariants

- The dispatch summary is built and approved before the first sub-agent starts.
- `explorer`, `implementer`, and `merger` remain the stable named roles. Their
  role semantics are not replaced by a generic agent name.
- `reviewer` and `red-team` are visible in the roster whenever selected, with
  their own model, reasoning, isolation, concurrency, and budget entries.
- Each dispatch receives the approved resolved model and reasoning explicitly,
  and obeys the approved isolation, concurrency, and budget constraints.
- Adding an optional role, changing any approved cost dimension, or expanding
  review scope pauses for renewed approval; it is never an implicit recovery
  path.
- Handoffs are context-complete: later agents receive the plan, confidence,
  assumptions, unresolved risks, and evidence rather than re-deriving them.
