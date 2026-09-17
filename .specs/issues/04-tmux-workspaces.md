# 04: Open task worktrees in tmux

**What to build:** Task creation automatically opens the new worktree in its own tmux pane without launching an AI harness.

**Blocked by:** 03: Create provisioned task worktrees.

**Status:** ready-for-agent

- [ ] Running task creation inside tmux opens a new pane in the task worktree.
- [ ] Running task creation outside tmux creates or attaches to a QBS-managed tmux session.
- [ ] The new pane starts with the task worktree as its working directory.
- [ ] No Codex, Claude, OpenCode, or other agent process is launched automatically.
- [ ] Missing tmux produces a clear actionable error.
- [ ] Tmux command construction is tested without requiring an actual interactive harness.
