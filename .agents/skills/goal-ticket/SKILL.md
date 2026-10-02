---
name: goal-ticket
description: Prepare a bounded `/goal` prompt to implement one ready ticket from a QBS feature spec in `.specs/`, with approved subagent model assignments. Use when the user wants a goal prompt for one ticket; this skill prepares the prompt and never implements the ticket.
---

Prepare a goal prompt for one ticket in a spec. The user starts `/goal` in a goal-capable session, then pastes the generated prompt. Keep the prompt text separate from the `/goal` command. Discover facts before asking questions; ask only for choices or facts that cannot be established. Never dispatch agents or implement the ticket in this preparation session.

Use the shared [orchestration-gate](../orchestration-gate/SKILL.md) for confidence, model planning, approval, cost status, and evidence vocabulary, and read the shared [goal prompt orchestration contract](../../goal-prompt-orchestration.md). The generated prompt must identify the active goal session as orchestrator and assign models to its worker subagents without naming or assuming a particular harness.

## Discover

Find the repository root, its primary worktree, default branch, and current branch. QBS keeps canonical `.specs/` and `.research/` in the primary worktree and exposes them to linked worktrees; use those shared paths as authoritative. Find every `spec.md` below `.specs/`, including existing nested specs; note whether `.specs` is ignored. For every relevant spec, inspect tickets for number/title/status/blockers/research references, any implementation state or handoffs, and research files under `.research/<feature>/` or explicitly cited elsewhere. Compute readiness yourself: not done, and all blockers done and recorded as merged. If no run state exists, only tickets without blockers are ready. Compare your frontier with any recorded frontier and identify omissions.

Read the chosen spec's build/test decisions, README, CI/build configuration, project files, and relevant `AGENTS.md`/`CLAUDE.md`. Check only prerequisites and test data needed by the selected ticket. Read cited research and later ticket findings relevant to it, plus available handoffs for blockers. Treat code locations in research as leads pinned to a commit: verify drift against current code. Resolve environment facts from the spec context and handoffs; explicitly rule out a wrong resource if research names one. Derive only the test seams and project rules needed for this ticket.

## Ask once

Use the available question tool for one round. Ask which spec only if there are multiple; ask which ready ticket, recommending the lowest number first. Show blocked tickets and blockers in the question text, but do not offer blocked tickets unless the user chooses Other. Offer no more than four ticket options. Also ask the user to confirm/correct the discovered build, tests, prerequisites, base/integration branches, and turn limit. Present the subagent roster and ask the user to choose a model profile and reasoning level for every worker role: `implementer`, `review-standards`, `review-spec`, and `merger` (also the serial recorder). Record the active goal model as the orchestrator when visible, separately from these worker assignments. Offer `Recommended`, `Economy`, `Deep`, and `Customize` plans as defined by orchestration-gate; resolve every selected role through `internal/agents` and obtain approval before generating the command. Use 60 turns by default; suggest 80 for substantial work and 100 for especially large tickets. Ask only unresolved questions that change the output. If no ticket is ready, report blockers and generate no command.

Confirm the integration and base branches, defaulting to the spec run state (or spec folder name) and repository default branch. For a first run, explain that the integration branch will start at the base branch's current SHA. The goal should adopt an existing integration branch only if it is at that SHA.

## Generate

Fill [template.md](template.md) with discovered facts and the user's answers. Include absolute paths from the primary worktree, exact build/test commands, relevant research files or sections, blocker handoffs, prerequisites, needed test seams, 2–3 compact project rules, and the approved orchestrator/subagent role split with resolved models and reasoning. Include ignored-file setup and baseline failures from run state/handoffs when available, not from assumptions. Do not assume README alone contains those facts. Do not include `/goal` in the generated prompt; the user starts `/goal` and then pastes the prompt.

This skill establishes `S/implementation-state.md` and `S/handoffs/NN.md` as its own execution journal. Discover them if present; on the first run, have the goal create the run state and record the starting SHA and baseline. Create the selected ticket's handoff at completion. QBS does not provide these files by default. Preserve any existing spec-level handoff and use it as context rather than replacing it.

Preserve these workflow invariants in every generated command:

- Work on one ticket only. Never implement the ticket in this preparation session.
- `.specs/` and `.research/` are shared, ignored QBS context. Keep updates local and uncommitted. The implementer and reviewers treat them as read-only; the merger/recorder is the single writer for the selected ticket, its handoff, run state, and ticket findings note. The orchestrator reads and verifies those records but does not author code or project documents. Use the primary worktree's absolute paths.
- Use ticket branch `B-NN` (not `B/NN`). Work in the session's own worktree, create the branch from latest `B`, run the baseline, and use baseline results from run state. Never recreate or rebase `B`. On first run, record BASE's SHA and adopt an existing `B` only if it is at that SHA.
- Research step A reads the ticket's cited files or sections and relevant later findings; verify stale code pointers and record drift. Include an explicit `BLOCKED` outcome for unresolved blockers, unavailable required prerequisites, or impossible spec decisions.
- Map every acceptance criterion to a change and an appropriate test seam before implementation. Test first, run the build and all relevant suites, review with `/code-review`, fix findings, then integrate with `--no-ff` into `B` in the primary worktree. If that checkout is on another branch, switch it to `B` only when clean, otherwise block; restore its branch afterwards.
- Update ticket status, handoff, run state, and frontier; include what the ticket unblocks. Before the final report, create `.research/<feature>/ticket-NN-implementation-findings.md` with the merge SHA, evidence-backed findings, and a `Still unclear or open` list. Choose the research feature directory by subject, even when the spec path is nested. Clean up by switching the worktree back to its session branch and deleting `B-NN`; done condition checks `git branch --list B-NN` is empty.
- Never weaken, skip, or delete tests. No push, no PR, and never touch BASE. Do not claim baseline failures as regressions.
- Keep the completed prompt at 4,000 bytes or fewer, excluding the `/goal` command and any surrounding newline. Trim project rules, then parentheticals, then filler. Never remove a done condition, workflow step, research read/update instruction, or `BLOCKED` clause.

Measure the final prompt with a byte count. Print only the prompt in one fenced `text` block, without a leading `/goal`. Below it, report the count; tell the user to start `/goal` in a fresh goal-capable session opened in the primary worktree and then paste the prompt; list other ready tickets not to run concurrently (they share run state); say what finishing this ticket unblocks; ask them to run this prompt again for the next ticket; and briefly list what's different from the previous prompt, if available.

If the chosen ticket is not ready, do not generate a prompt. Explain its blockers. Do not modify project files or run the generated goal.
