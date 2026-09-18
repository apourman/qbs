---
name: implement-spec
description: "Implement a specification in code."
---

You have been provided a spec. This spec should have tickets associated with it, describing how to implement the spec.

The goal is a coherent change which implements the entire spec. You own the
branch, worktree, PR, and merge setup for the change.

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

3. (optional) Use the **`explorer`** agent to conduct any exploration required
   by the tickets - relevant codebase files or external documentation. Apply
   the selected cost level before dispatching it. Ensure the explorer can save
   files - it should save its markdown notes in a directory outside the repo,
   accessible by all future agents. Pass the approved assignment and the
   implementation context to it. This lets **`implementer`** focus on
   implementation rather than exploration. If exploration is not selected,
   omit the role from the roster and do not dispatch it.

4. Create a dedicated branch and worktree for the spec, then create a draft PR
   which references the spec and its tickets.

5. Start one bounded implementation task per ready ticket. Give each
   implementer its own branch and worktree. Set the task's model and reasoning
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

9. Mark the PR as ready for review once the review findings are fixed.

10. Produce the orchestration-gate postflight report, even when a section is
    empty. Include the new confidence score and rationale; verified evidence
    with precise pointers; inferred conclusions; requirement coverage
    (`complete`, `partial`, `missing`, `ambiguous`, or `unrequested`);
    assumptions; low-confidence areas; risks and likely failure modes;
    unnecessary complexity that can be removed without dropping required
    behavior; red-team findings when that role ran; and the smallest
    confidence-raising checks. Scores 1/2 are incomplete or unsafe, score 3
    requests targeted follow-up, score 4 reports remaining limited risks, and
    score 5 requires direct coverage and verification evidence.

11. Clean up completed implementer worktrees and branches. Return the final
    revision, checks, approved plan, handoff evidence, confidence report, and
    remaining gates. Keep token estimates, account quota, and monetary cost
    distinct; unavailable measurements must remain unavailable or advisory,
    never zero or exact.

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
