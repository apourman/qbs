# 03: Create provisioned task worktrees

**What to build:** `qbs task <name>` creates an isolated task branch and Git worktree with independent local AI tooling.

**Blocked by:** 01: Create the Go CLI foundation; 02: Implement repository initialization.

**Status:** ready-for-agent

- [ ] The task name is validated before any repository changes are made.
- [ ] A predictable task branch is created.
- [ ] A predictable sibling worktree is created.
- [ ] The new worktree receives shared agent instructions.
- [ ] The new worktree receives Codex, Claude, and OpenCode skills.
- [ ] The new worktree receives independent research storage.
- [ ] The new worktree remains clean from Git's perspective because the local AI files are ignored.
- [ ] Multiple tasks can be created without sharing their local AI files.
- [ ] Branch collisions and existing worktree paths produce clear errors.
