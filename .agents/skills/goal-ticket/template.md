# Goal plan and prompt templates

Replace every placeholder. Omit prior-research wording when no prior notes exist, but still create the ticket findings file. The plan holds the full run instructions; the prompt only points to it.

## Plan

Write to `S/goal-plan-NN.md` (PLAN).

```markdown
# Ticket NN goal plan

Full run instructions for the /goal. The pasted goal prompt points here; follow this file exactly.

Complete ticket NN from spec S, review it, merge it into `B`, and hand it off. Done only when: BUILD and touched suites pass; `/code-review` against BASE is complete and findings are fixed or explained; `git log --oneline -3 B` shows `Merge B NN` and merge checks pass; every acceptance criterion is complete/partial/missing with evidence and confidence 1–5; ticket status, handoff, run state and findings are updated; and `git branch --list B-NN` is empty. If a blocker is not merged, a prerequisite is unavailable, or a spec decision is impossible, record `BLOCKED: <reason>`. Stop after TURNS turns.

Orchestrator: ORCHESTRATOR_MODEL. It reads, plans, delegates, monitors and reviews; it writes no repository files. Approved subagents and assignments (profile/resolved model/reasoning/budget/isolation/concurrency): implementer=IMPLEMENTER; review-standards=REVIEW_STANDARDS; review-spec=REVIEW_SPEC; merger (integration and serial recorder)=MERGER. User approved these assignments. Pass each through the active subagent mechanism; if unavailable, stop without substitution or dispatch.

Primary worktree MAIN; spec S/spec.md; ticket MAIN/.backlog/items/NN-*.md (status and blocked-by in its frontmatter); prior research RESEARCH; blocker handoffs HANDOFFS; state S/implementation-state.md; handoff S/handoffs/NN.md; findings MAIN/.research/FEATURE/ticket-NN-implementation-findings.md. `.specs`/`.research` are shared, ignored, local context; do not commit. BUILD: BUILD. Targeted tests: TEST. Expensive suites: EXPENSIVE_SUITES. Project rules: PROJECT_RULES.

Testing scope: run only the tests covering code this ticket touches, using the targeted commands above. Never run the full suite, including for the baseline, review fixes, or merge checks. Run tests from an expensive suite only when the change directly affects code they exercise, and then only those tests. The full suite is left to CI as a remaining gate.

A. Read the spec, ticket, cited research and relevant later findings, run state and blocker handoffs. Verify code pointers against current code and note drift. Every blocker must be done and merged. Check prerequisites PREREQUISITES and tracked status; otherwise BLOCKED.
B. First run: record BASE SHA in the state; do not run a full-suite baseline. Create `B` at that SHA or adopt it only if its recorded SHA matches; otherwise BLOCKED. Later runs never recreate or rebase `B`.
C. Create `B-NN` from latest `B` in an isolated worktree, or resume it if the state records it, and delegate it to implementer. Set up ignored inputs INPUTS. Run the targeted tests before changing code and return them as the ticket baseline. Map every acceptance criterion to a change and test seam; test first, implement, run BUILD and the targeted tests for touched code, commit `B NN: <title>`. Implementer does not edit shared spec or research records.
D. Run `/code-review` against BASE with the approved review-standards and review-spec assignments. Delegate fixes to implementer, re-test, and commit review fixes.
E. If `B` moved, merge it into `B-NN` and re-test. Merger switches the primary worktree to `B` only when clean (otherwise BLOCKED), merges with `--no-ff` as `Merge B NN: <title>`, runs the targeted tests on the merge, then restores the prior branch.
F. Merger alone updates ticket status, handoff, state, frontier, and findings. Include acceptance evidence, implementation/review results, research drift, later-ticket notes, and what NN unblocks. Findings include merge SHA, evidence-backed results, open items, and `Still unclear or open`. On a blocker or turn limit, record status and remaining frontier; do not create findings without a merge SHA.
G. Merger returns the worktree to its session branch and deletes `B-NN`; verify it is absent. Orchestrator reports the result without repository edits.

Never weaken, skip, or delete tests; baseline failures are not regressions. Test only touched code; never run the full suite. This ticket only; note later-ticket needs. No push or PR; never modify BASE.
```

## Prompt

Write to `S/goal-prompt-NN.md` and paste into `/goal`; paste it again in a new session to resume. Keep it at 4,000 bytes or fewer.

```text
Complete ticket NN from spec S, review it, merge it into `B`, and hand it off, following the goal plan at PLAN. Read that plan in full before doing anything else and follow it exactly: it holds the approved subagent assignments, branches, checks, and steps. If it is missing, or any approved assignment cannot be honored, stop and report without dispatch or substitution. If S/implementation-state.md already records progress on this ticket, resume from it rather than starting over.

Done only when: BUILD and touched suites pass; `/code-review` findings are fixed or explained; `B` has the `Merge B NN` commit and merge checks pass; every acceptance criterion is evidenced; ticket status, S/handoffs/NN.md, S/implementation-state.md and MAIN/.research/FEATURE/ticket-NN-implementation-findings.md are updated; and `git branch --list B-NN` is empty. If a blocker is not merged, a prerequisite is unavailable, or a spec decision is impossible, have the merger record status, then print `BLOCKED: <reason>`. Stop after TURNS turns.

Orchestrator: ORCHESTRATOR_MODEL. It reads, plans, delegates, monitors and reviews; it writes no repository files. The merger is the single writer of shared records.

Never weaken, skip, or delete tests; baseline failures are not regressions. Test only touched code; never run the full suite. This ticket only. Keep .specs, .research and .artifacts local and uncommitted. No push or PR; never modify BASE.
```
