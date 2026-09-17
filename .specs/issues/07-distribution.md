# 07: Package and distribute QBS

**What to build:** Release-ready QBS binaries that include the canonical AI templates and can be installed on common developer platforms.

**Blocked by:** 02: Implement repository initialization; 03: Create provisioned task worktrees; 04: Open task worktrees in tmux; 05: Manage task worktrees; 06: Harden workspace safety and error handling.

**Status:** ready-for-agent

- [ ] Release builds produce standalone binaries for Linux, macOS, and Windows targets.
- [ ] Release binaries contain the canonical agent instructions and portable skills.
- [ ] A fresh installation can initialize and provision a repository without access to the QBS source tree.
- [ ] Build metadata exposes the QBS version.
- [ ] Release artifacts and installation instructions are documented.
- [ ] A release build smoke test exercises initialization and task creation with embedded templates.
