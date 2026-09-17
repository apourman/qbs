---
name: qbs
description: Use QBS to provision local AI workspace state or manage the user's global curated skill catalog. Applies to QBS initialization, external worktrees, `.specs/`, `.research/`, and global skill synchronization.
---

# QBS

Use `qbs init` for the current repository and `qbs provision <worktree>` after
Firstmate, Treehouse, or another tool creates a worktree. Use `qbs task` only
when QBS owns the standalone task lifecycle.

Manage reusable skills globally with `qbs skills import`, `list`, `sync`, and
`remove`. Keep repository-specific context in `AGENTS.md`, `CLAUDE.md`,
`.specs/<domain-or-feature>/`, `.research/<domain-or-feature>/`, and
project-local skill directories. Put spec tickets in the spec directory's
`issues/` subdirectory.
