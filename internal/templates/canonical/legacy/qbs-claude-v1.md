# Shared project instructions

Follow the repository's documented conventions. Verify changes with focused
checks and tests that protect meaningful behavior; avoid redundant tests of
implementation details. Run broader suites when required by the repository or
when focused checks leave a material risk.

For implementation work, plan to reserve roughly 25% of the available task
budget for review, fixes, CI, and integration. Treat this as a planning reserve,
not a guarantee. Keep token budgets, account quota, and monetary cost distinct;
never invent conversions between them or promise completion within a quota.
When delegating, choose the least costly available model and reasoning effort
that can reliably handle the work, subject to the user's model constraints.
When Firstmate or another orchestrator supplies an isolated workspace, it owns
branch, worktree, PR, and merge lifecycle; agents work within that workspace.

Treat `.specs/` and `.research/` as worktree-local context.

## Canonical knowledge layout

Organize project knowledge by a stable, lowercase kebab-case domain or feature
name:

- Research notes go in `.research/<domain-or-feature>/<research-slug>.md`.
- Specs go in `.specs/<domain-or-feature>/<spec-slug>.md`.
- Tickets created for a spec go in `.specs/<domain-or-feature>/issues/<NN>-<ticket-slug>.md`.

Use the same domain-or-feature directory across related research, the spec, and
its tickets. When creating any of these artifacts, create the directory if it
does not exist and report the path. Existing files outside this layout are
legacy context; do not move them unless asked.
