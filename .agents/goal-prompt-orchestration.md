# Goal prompt orchestration

Use this contract in every QBS skill that prepares a `/goal` prompt.

## Separate orchestration from work

Treat the active goal session as the orchestrator. Record its model and reasoning effort when visible, but do not ask the user to assign that model as though it were a worker. The orchestrator reads, plans, delegates, monitors, and reviews agent results; it does not author implementation code or project documents.

Before generating the goal, ask the user to choose the model profile and reasoning effort for every selected subagent role that performs work, including code and test implementation, reviews, integration, exploration, and project-record writing. Use the shared `orchestration-gate` and the `internal/agents` catalog to resolve role assignments for the active environment. Show the resolved assignment for each role, get approval, and include the approved assignments in the generated prompt. If the active environment cannot apply an assignment, stop with an actionable explanation; never silently substitute.

The generated prompt must explicitly identify the orchestrator and name the subagents, their tasks, and their approved model assignments. Delegate all repository writes to the assigned subagents. Give shared files a single serial writer, normally the merger role with recorder duties after integration; other agents return findings and keep shared records read-only. Have that same writer record blockers, turn-limit state, and the remaining frontier. The orchestrator verifies those updates without authoring them.

Use the active environment's subagent mechanism without naming or assuming a particular harness. Keep this language separate from harness-specific invocation details, which belong to the active runtime.
