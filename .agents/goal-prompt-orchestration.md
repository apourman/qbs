# Goal prompt orchestration

Use this contract in every QBS skill that prepares a `/goal` prompt.

## Separate orchestration from work

Treat the active goal session as the orchestrator. Record its model and reasoning effort when visible, but do not ask the user to assign that model as though it were a worker. The orchestrator reads, plans, delegates, monitors, and reviews agent results; it does not author implementation code or project documents.

Before generating the goal, ask the user to choose the model profile and reasoning effort for every selected subagent role that performs work, including code and test implementation, reviews, integration, exploration, and project-record writing. Use the shared `orchestration-gate` and the `internal/agents` catalog to resolve role assignments for the active environment. Show the resolved assignment for each role, get approval, and include the approved assignments in the generated prompt. If the active environment cannot apply an assignment, stop with an actionable explanation; never silently substitute.

The generated prompt must explicitly identify the orchestrator and name the subagents, their tasks, and their approved model assignments. Delegate all repository writes to the assigned subagents. Give shared files a single serial writer, normally the merger role with recorder duties after integration; other agents return findings and keep shared records read-only. Have that same writer record blockers, turn-limit state, and the remaining frontier. The orchestrator verifies those updates without authoring them.

Use the active environment's subagent mechanism without naming or assuming a particular harness. Keep this language separate from harness-specific invocation details, which belong to the active runtime.

## Write the plan, paste a pointer

Write the full run instructions to a goal plan file in the primary checkout, at the path the skill names: `S/goal-plan.md` beside `spec.md` for a whole spec, or `S/tickets/goal-plan-NN.md` for one ticket. Goal plan files are not tickets; skip them when discovering tickets. Then generate a short goal prompt that points to it. The plan file is shared, Git-ignored `.specs/` context: never commit it. Start the plan with a `# <Feature> goal plan` heading and the line `Full run instructions for the /goal. The pasted goal prompt points here; follow this file exactly.` If a plan file already exists, read it first; replace it only with a newly approved plan and say what changed.

Keep the pasted prompt at 4,000 bytes or fewer, excluding the `/goal` command. It must state the objective, the plan's absolute path with an instruction to read it in full before acting and follow it exactly, the done condition, the orchestrator and its no-authoring rule, the `BLOCKED` outcome, the turn limit, and the hard safety rules (never weaken tests; no push, pull request, or base-branch change; keep `.specs/`, `.research/`, and `.artifacts/` uncommitted). It must also say to stop without dispatch if the plan file is missing or its approved assignments cannot be honored. Everything else, including the role and ticket dispatch tables, checks, and steps, lives only in the plan.

Measure both files with a byte count. Print only the prompt in one fenced `text` block, without a leading `/goal`, and report the plan path and both byte counts below it.
