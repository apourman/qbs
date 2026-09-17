# 06: Harden workspace safety and error handling

**What to build:** QBS handles invalid environments and dangerous workspace operations predictably, with clear user-facing errors.

**Blocked by:** 03: Create provisioned task worktrees; 04: Open task worktrees in tmux; 05: Manage task worktrees.

**Status:** ready-for-agent

- [ ] Missing Git is reported clearly.
- [ ] Missing tmux is reported clearly when tmux setup is required.
- [ ] Invalid task names are rejected before changes begin.
- [ ] Existing branches and worktree paths are handled without partial setup.
- [ ] Existing local AI files are protected from silent replacement.
- [ ] Partial failures report what was completed and what remains to be cleaned up.
- [ ] Error behavior is covered across initialization, task creation, tmux setup, and removal.
