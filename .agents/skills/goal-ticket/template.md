# Goal prompt template

Replace every placeholder. Omit prior-research wording when no prior notes exist, but still create the ticket findings file. Keep the completed prompt at or below 4,000 bytes.

```text
Complete ticket NN from spec S, review it, merge it into `B`, and hand it off. Done only when: BUILD and touched suites pass; `/code-review` against BASE is complete and findings are fixed or explained; `git log --oneline -3 B` shows `Merge B NN` and merge checks pass; every acceptance criterion is complete/partial/missing with evidence and confidence 1–5; ticket status, handoff, run state and findings are updated; and `git branch --list B-NN` is empty. If a blocker is not merged, a prerequisite is unavailable, or a spec decision is impossible, record `BLOCKED: <reason>`. Stop after TURNS turns.

Orchestrator: ORCHESTRATOR_MODEL. It reads, plans, delegates, monitors and reviews; it writes no repository files. Approved subagents and assignments (profile/resolved model/reasoning/budget/isolation/concurrency): implementer=IMPLEMENTER; review-standards=REVIEW_STANDARDS; review-spec=REVIEW_SPEC; merger (integration and serial recorder)=MERGER. User approved these assignments. Pass each through the active subagent mechanism; if unavailable, stop without substitution or dispatch.

Primary worktree MAIN; spec S/spec.md; ticket S/tickets/NN-*.md; prior research RESEARCH; blocker handoffs HANDOFFS; state S/implementation-state.md; handoff S/handoffs/NN.md; findings MAIN/.research/FEATURE/ticket-NN-implementation-findings.md. `.specs`/`.research` are shared, ignored, local context; do not commit. BUILD: BUILD. TEST: TEST. Project rules: PROJECT_RULES.

A. Read the spec, ticket, cited research and relevant later findings, run state and blocker handoffs. Verify code pointers against current code and note drift. Every blocker must be done and merged. Check prerequisites PREREQUISITES and tracked status; otherwise BLOCKED.
B. First run: record BASE SHA and baseline in the state. Create `B` at that SHA or adopt it only if its recorded SHA matches; otherwise BLOCKED. Later runs never recreate or rebase `B`.
C. Create `B-NN` from latest `B` in an isolated worktree and delegate it to implementer. Set up ignored inputs INPUTS. Map every acceptance criterion to a change and test seam; test first, implement, run BUILD/TEST, commit `B NN: <title>`. Implementer does not edit shared spec or research records.
D. Run `/code-review` against BASE with the approved review-standards and review-spec assignments. Delegate fixes to implementer, re-test, and commit review fixes.
E. If `B` moved, merge it into `B-NN` and re-test. Merger switches the primary worktree to `B` only when clean (otherwise BLOCKED), merges with `--no-ff` as `Merge B NN: <title>`, tests the merge, then restores the prior branch.
F. Merger alone updates ticket status, handoff, state, frontier, and findings. Include acceptance evidence, implementation/review results, research drift, later-ticket notes, and what NN unblocks. Findings include merge SHA, evidence-backed results, open items, and `Still unclear or open`. On a blocker or turn limit, record status and remaining frontier; do not create findings without a merge SHA.
G. Merger returns the worktree to its session branch and deletes `B-NN`; verify it is absent. Orchestrator reports the result without repository edits.

Never weaken, skip, or delete tests; baseline failures are not regressions. This ticket only; note later-ticket needs. No push or PR; never modify BASE.
```
