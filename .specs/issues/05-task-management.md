# 05: Manage task worktrees

**What to build:** Developers can inspect and safely remove QBS-created task worktrees.

**Blocked by:** 03: Create provisioned task worktrees.

**Status:** ready-for-agent

- [ ] A task-listing command shows task branches and worktree locations.
- [ ] Task listing works from the main worktree and linked worktrees.
- [ ] A task-removal command removes the selected worktree safely.
- [ ] Removal protects modified tracked files by default.
- [ ] Removal clearly warns that ignored local AI files and research notes will be deleted.
- [ ] Explicit force behavior is available for intentionally destructive cleanup.
- [ ] Task listing and removal are covered by temporary-repository integration tests.
