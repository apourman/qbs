---
name: update-docs
description: Bring every document touched by this session's work up to date — repository docs plus the shared `.research/`, `.specs/`, `.artifacts/`, and `.backlog/` context, including republishing changed artifacts. Use when the user says "update docs", "sync the docs", or wants documentation caught up with what was done in a session, typically before shipping or handing off.
---

# update-docs

Make the documentation match what this session actually did. Change only
what the work made stale or left unrecorded; do not rewrite docs for style.

## 1. Inventory the session's work

List what changed and what was decided, from:

- The conversation: decisions, rejected options, facts verified, open
  questions, tickets started or finished.
- Git: commits made this session (`git log --oneline <base>..HEAD`), the
  working-tree diff, and branches merged.

## 2. Find the affected documents

Shared context lives in the primary checkout, even from a linked worktree:
`dirname "$(git rev-parse --path-format=absolute --git-common-dir)"`.

- `.specs/<domain>/spec.md` — the spec the work belongs to; its tickets are
  the backlog items with `area: <domain>`.
- `.research/<domain>/` — notes whose findings the work confirmed, refuted,
  or extended.
- `.backlog/items/` — items the work started, investigated, or resolved,
  and items for bugs or ideas the session found but did not act on.
- `.artifacts/index.md` and `.artifacts/<domain>/*.html` — pages that show
  behavior, plans, or status the work changed.
- Repository docs: `README.md`, `CONTRIBUTING.md`, `INSTALL.md`, `docs/`,
  `CONTEXT.md`, `docs/adr/`, `AGENTS.md`/`CLAUDE.md`, and skill files whose
  instructions the work changed.

Grep these for the names of changed commands, flags, files, and concepts.

## 3. Update them

- **Specs**: record decisions and scope changes; add follow-up items in the
  spec's area for deferred work.
- **Backlog**: follow `docs/agents/backlog.md` — set worked items to `doing`
  or `done` with a pointer to the commit or PR, add a dated `## Findings`
  section, and link any spec created from the item. Add new items
  for unrecorded bugs or ideas.
- **Research**: add verified findings with sources; mark superseded claims.
- **Artifacts**: follow the canonical layout rules — read the live version,
  keep the newer copy, edit the local file, bump its version in the page
  header and in `.artifacts/index.md`, then publish that file to its existing
  URL. If this account cannot publish it, update the local file and index and
  tell the user it needs republishing from the owning account.
- **Repository docs**: fix stale usage, examples, and paths; record
  architectural decisions as ADRs when the repo keeps them.

New work with no home gets one in the canonical layout
(`.<kind>/<domain>/...`); report each created path.

## 4. Report

List each document updated, created, or published (with URL and version),
and anything left stale and why. Never commit `.specs/`, `.research/`,
`.artifacts/`, or `.backlog/`. Leave repository doc changes uncommitted unless the user asked
for a commit.
