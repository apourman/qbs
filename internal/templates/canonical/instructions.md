# QBS shared instructions

These instructions are installed globally by `qbs sync`; rerun it to update them.

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

`.specs/`, `.research/`, `.artifacts/`, and `.backlog/` are shared, Git-ignored context that
lives in the primary checkout. From a linked worktree, read and write the primary checkout's
copies, not the worktree's, and never commit them.

## Ponytail

Use the `ponytail` skill for all coding work: writing, fixing, refactoring,
designing, and reviewing code, and choosing dependencies. When reviewing a
diff, also run `ponytail-review` alongside any correctness review. Use
`ponytail-debt` to list the deliberate `ponytail:` shortcuts left in the code.

## Canonical knowledge layout

Organize project knowledge by a stable, lowercase kebab-case domain or feature
name:

- Research notes go in `.research/<domain-or-feature>/<research-slug>.md`.
- Specs go in `.specs/<domain-or-feature>/spec.md`, with exactly one spec per directory.
- Tickets created for a spec go in `.specs/<domain-or-feature>/tickets/<NN>-<ticket-slug>.md`.
- Backlog items (bugs, reports, and ideas not yet in a spec) go in
  `.backlog/items/<NNN>-<slug>.md`; see `docs/agents/backlog.md`.
- Visual pages (HTML artifacts) go in `.artifacts/<domain-or-feature>/<slug>.html`,
  listed in `.artifacts/index.md` with one line per page: title, path, what it
  is the source of truth for, and, when published, its URL, owning account and
  current version. Keep them as HTML and edit them in place. Show the version
  in the page header too.
- A published artifact and its local file stay identical. Edit the local file,
  then publish that same file to the artifact's existing URL. Never publish
  edits that are not in the local file, and never create a second URL for the
  same page.
- Before editing a published artifact, read the live version and compare it
  with the local file, ignoring the hosting wrapper the service adds before
  the page. Keep the newer one; if the live copy is newer, save it locally
  first. If a published artifact has no local copy, save one and add its index
  line.
- If the current account or runner cannot read or publish an artifact, still
  update the local file and its index version, and tell the user it needs
  republishing from the owning account.

Use the same domain-or-feature directory across related research, the spec, its
tickets, and its artifacts. When creating any of these artifacts, create the directory if it
does not exist and report the path. Existing files outside this layout are
legacy context; do not move them unless asked.

## Agent skills

When the repository has a `docs/agents/` directory:

- Issue tracker: see `docs/agents/issue-tracker.md` for the repository's tracker workflow.
- Backlog: see `docs/agents/backlog.md` for the item format and how items become specs or tickets.
- Domain documentation: see `docs/agents/domain.md` for `context.md` and `docs/adr/`.
- Local triage: when triage is installed, see `docs/agents/triage.md` for the status workflow.
