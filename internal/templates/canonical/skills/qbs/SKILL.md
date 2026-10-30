---
name: qbs
description: Use QBS to initialize local AI workspace state or manage the user's global curated skill catalog. Applies to QBS initialization, `.specs/`, `.research/`, `.artifacts/`, `.backlog/`, and global skill synchronization.
---

# QBS

Use `qbs init` for the current repository. Run `qbs backlog` to browse and
edit the repository's `.backlog/items/` in a browser, and
`qbs backlog new "<title>"` to create an item with the next free number.

Manage reusable skills globally with `qbs skills import`, `list`, `sync`, and
`remove`. `qbs sync` also installs the shared QBS instructions globally for
Claude, Codex, and OpenCode. Keep repository-specific context in `AGENTS.md`, `CLAUDE.md`,
`.specs/<domain-or-feature>/`, `.research/<domain-or-feature>/`,
`.artifacts/<domain-or-feature>/`, and
project-local files created by the project owner. `qbs init` does not create or
modify project-local agents or skills. Track a spec's tickets as backlog items
whose `area` is the spec's directory name.
