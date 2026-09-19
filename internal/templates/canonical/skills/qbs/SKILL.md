---
name: qbs
description: Use QBS to initialize local AI workspace state or manage the user's global curated skill catalog. Applies to QBS initialization, `.specs/`, `.research/`, and global skill synchronization.
---

# QBS

Use `qbs init` for the current repository.

Manage reusable skills globally with `qbs skills import`, `list`, `sync`, and
`remove`. Keep repository-specific context in `AGENTS.md`, `CLAUDE.md`,
`.specs/<domain-or-feature>/`, `.research/<domain-or-feature>/`, and
project-local files created by the project owner. `qbs init` does not create or
modify project-local agents or skills. Put spec tickets in the spec directory's
`tickets/` subdirectory.
