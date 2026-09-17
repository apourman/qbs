---
name: implement-spec
description: "Implement a specification in code."
disable-model-invocation: true
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

2. (optional) Use the **`explorer`** agent to conduct any exploration required
   by the tickets - relevant codebase files or external documentation. Apply
   the selected cost level before dispatching it. Ensure the explorer can save
   files - it should save its markdown notes in a directory outside the repo,
   accessible by all future agents. This lets **`implementer`** focus on
   implementation rather than exploration.

3. Create a dedicated branch and worktree for the spec, then create a draft PR
   which references the spec and its tickets.

4. Start one bounded implementation task per ready ticket. Give each
   implementer its own branch and worktree. Set the task's model and reasoning
   effort explicitly when supported. Use the named **`implementer`** when its
   fixed profile fits the ticket; otherwise select a suitable available model.

5. Once an implementation task completes, report its revision and validation to
   the main task, then use the named **`merger`** to inspect and merge it into
   the spec branch.

6. If this changes the **frontier** of available tickets, report the newly ready
   tickets and dispatch them as dependencies and the remaining budget permit.

7. Once all tickets are complete, run /code-review on the supplied workspace or
   revision. Fix its
   findings in one bounded implementation task.

8. Mark the PR as ready for review once the review findings are fixed.

9. Clean up completed implementer worktrees and branches. Return the final
   revision, checks, and remaining gates.
