# /goal command template

Replace every placeholder. Omit the research-read wording when no prior notes exist; still create the ticket findings file. Keep the entire command at or below 4,000 bytes.

```text
/goal Ticket NN of spec S is built, reviewed, merged into branch `B` and handed off. Done only when this conversation shows ALL of: (1) passing output of BUILD and every test suite touched; (2) a /code-review of the ticket branch, with every finding fixed or given a reason; (3) `git log --oneline -3 B` showing "Merge B NN", and tests passing on that merge; (4) a final report marking every acceptance criterion complete/partial/missing with evidence (test name or file:symbol) and confidence 1-5; (5) ticket Status done, handoff, run state and ticket findings updated; (6) `git branch --list B-NN` empty. Also done if this session prints "BLOCKED: <reason>" (a blocker is not done and merged, a required prerequisite is unavailable, or a spec decision is impossible in code). Stop after TURNS turns.

Primary worktree MAIN (S=SPEC): spec S/spec.md; ticket S/tickets/NN-*.md; prior research RESEARCH_PATHS; run state S/implementation-state.md; handoff S/handoffs/NN.md; findings MAIN/.research/FEATURE/ticket-NN-implementation-findings.md. `.specs` and `.research` are ignored shared QBS context; use these absolute paths, edit only this ticket and its execution records, never commit them. BUILD: BUILD. TEST: TEST.

Steps:
A. Read the spec, ticket, cited research files/sections RESEARCH and relevant later findings, available run state and blocker handoffs HANDOFFS. Verify code leads pinned to COMMIT against current code and note drift. Every blocker must be done and merged, else print BLOCKED. Check relevant prerequisites PREREQUISITES and clean tracked status.
B. First run only: create the run state with BASE's SHA; adopt B if already at that SHA, otherwise create B there. Later runs: never recreate or rebase B.
C. In this session's worktree, with clean `git status`, create branch `B-NN` from latest `B`. Set up ignored inputs INPUTS. Run TEST; record first-run baseline in run state.
D. Explore, then print a plan mapping every acceptance criterion to a change and test at needed seams SEAMS, with confidence, assumptions and risks. If confidence <3 or intent may change, print BLOCKED.
E. Test first: per criterion add a test in project style, observe failure, implement. Run BUILD and TEST. Commit "B NN: <title>".
F. Run /code-review against the base commit; fix findings in one pass, re-test, commit "B NN: review fixes".
G. If B moved, merge it into the ticket branch and re-test. In the primary worktree (switch to B only if clean, else BLOCKED; restore branch after), merge --no-ff as "Merge B NN: <title>", then build and test the merge.
H. Tick met criteria in ticket; Status done only if all met, otherwise in-progress with reason. Write handoff with implementation, uncovered decisions, research drift, review, acceptance report and later-ticket notes. Update run state with status, merge SHA, handoff and recomputed frontier (include omitted ready tickets and what NN unblocks).
I. Switch this worktree back to its session branch; delete B-NN.
J. Write the ticket findings file with merge SHA, evidence-backed findings, unaddressed items and a `Still unclear or open` list. Then print final report.

Rules: this ticket only; stub minimum later-ticket needs and note them. Spec decisions are settled; if one is wrong, print BLOCKED. Never weaken, skip or delete tests; baseline failures are not regressions. No push, no PR, never touch BASE.
PROJECT_RULES
```
