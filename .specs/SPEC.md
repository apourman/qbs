# QBS — Local AI Workspace Provisioner

## Problem Statement

When working on multiple plans or tasks in parallel, each task needs an isolated Git worktree so that changes do not conflict. Each worktree also needs local Agentic AI tooling: agent instructions, skills, and research notes.

Git worktrees do not copy ignored or untracked files. As a result, provisioning AI tooling once in a main worktree does not make it available in later worktrees. Users currently need to repeat this setup manually, which is inconvenient and makes parallel work in separate tmux panes harder to manage.

## Solution

Build a small, distributable `qbs` CLI binary. It configures a Git repository for local-only AI tooling and creates fully provisioned task worktrees.

The primary workflow is:

```text
qbs init
qbs task <name>
```

`qbs init` configures the repository and provisions the current worktree. `qbs task <name>` creates a task branch and worktree, installs the local AI files into that worktree, and opens a tmux pane with its working directory set to the new worktree.

The CLI prepares workspaces for Codex, Claude, and OpenCode, but does not launch any AI harness. The user chooses which harness to run from the prepared pane.

All provisioned AI files are local and ignored by Git. Each worktree receives independent copies of the files so that plans can have independent instructions, skills, and research notes.

## User Stories

1. As a developer, I want to initialize a repository for local AI tooling, so that I do not have to configure ignore rules and agent files manually.
2. As a developer, I want `qbs init` to be safe to run repeatedly, so that I can repair or refresh setup without duplicated rules or unexpected changes.
3. As a developer, I want AI tooling to remain untracked, so that local agent state does not appear in project commits.
4. As a developer, I want to create a task with one command, so that starting isolated work is fast.
5. As a developer, I want each task to have its own Git worktree, so that parallel plans cannot overwrite one another's working files.
6. As a developer, I want each task to have a predictable branch, so that task work is easy to identify and manage.
7. As a developer, I want each task worktree to receive `AGENTS.md`, so that Codex and compatible agents can discover project instructions.
8. As a developer, I want each task worktree to receive `CLAUDE.md`, so that Claude can discover project instructions.
9. As a developer, I want `AGENTS.md` to direct agents to `CLAUDE.md`, so that shared instructions do not diverge unnecessarily.
10. As a developer, I want skills available to Codex, Claude, and OpenCode, so that I can choose a harness without manually copying skill files.
11. As a developer, I want each harness to receive skills in its native skills directory, so that each harness can discover the same portable skills.
12. As a developer, I want each task to have a separate research directory, so that research from one plan does not pollute another plan.
13. As a developer, I want task creation to open a tmux pane automatically, so that the task is immediately ready for interactive work.
14. As a developer, I want the tmux pane to start in the new worktree, so that commands run against the intended task by default.
15. As a developer, I want QBS to work both inside and outside tmux, so that I can use it from a normal terminal or an existing tmux session.
16. As a developer, I want QBS to create a tmux session when no tmux session exists, so that automatic pane setup remains consistent.
17. As a developer, I want QBS to avoid launching Codex, Claude, or OpenCode, so that I remain in control of which harness runs.
18. As a developer, I want to list task worktrees, so that I can find active plans and their locations.
19. As a developer, I want to remove a task worktree safely, so that completed plans do not leave unused worktrees behind.
20. As a developer, I want QBS to protect modified tracked files during removal, so that cleanup cannot silently destroy work.
21. As a developer, I want clear errors when Git or tmux is unavailable, so that I know how to correct the environment.
22. As a developer, I want QBS distributed as a single binary, so that installing it does not require a language runtime.
23. As a developer, I want QBS binaries for common Linux, macOS, and Windows platforms, so that the workflow is portable.
24. As a maintainer, I want canonical AI templates embedded in the binary, so that users do not need to clone or download a separate template repository.
25. As a maintainer, I want portable skills separated from harness-specific metadata, so that the same skill behavior can be maintained across supported harnesses.

## Implementation Decisions

- Implement QBS in Go and distribute it as a standalone executable.
- Use the Go standard library for command parsing, filesystem operations, process execution, and embedded templates. Avoid introducing a CLI framework until the command surface requires one.
- Treat Git and tmux as runtime system dependencies. QBS itself does not launch an AI harness.
- Keep canonical templates in the QBS source and embed them into release binaries. Installed copies in user repositories are generated local state, not source-controlled project files.
- Support these initial commands:
  - `qbs init`
  - `qbs task <name>`
  - `qbs tasks`
  - `qbs task remove <name>`
- Require `qbs init` to run from within a Git repository. Detect the repository root and the common Git directory rather than assuming a particular `.git` layout.
- Add these local paths to the repository's shared Git exclude file, idempotently:
  - `AGENTS.md`
  - `CLAUDE.md`
  - `.agents/skills/`
  - `.claude/skills/`
  - `.opencode/skills/`
  - `.research/`
- Provision `AGENTS.md`, `CLAUDE.md`, all supported skill directories, and `.research/` into the current worktree during initialization.
- Make `AGENTS.md` the general agent entry point and have it reference `CLAUDE.md` for shared instructions.
- Maintain one portable canonical skill set and copy it to the Codex, Claude, and OpenCode skill locations. Add harness-specific metadata only when required by a harness.
- Have `qbs task <name>` create a branch named `qbs/<name>` and a sibling worktree with a predictable task-specific directory name.
- Provision local AI files only after the worktree has been created. Do not depend on Git copying ignored files.
- When invoked inside tmux, create a split pane whose starting directory is the new worktree.
- When invoked outside tmux, create or attach to a QBS-managed tmux session and start the task there.
- Do not automatically start Codex, Claude, OpenCode, or any other agent process.
- Refuse or clearly warn on branch collisions, existing worktree paths, invalid task names, and attempts to overwrite existing local instruction files.
- Make task removal conservative: do not remove a worktree with modified tracked files unless an explicit force option is supplied, and clearly warn that ignored local files will be removed.
- Keep worktree-specific AI files independent. Do not use shared symlinks for instructions, skills, or research notes.

## Testing Decisions

- Test externally observable behavior rather than internal helper structure.
- Use temporary Git repositories for integration tests so tests verify behavior with real repository roots, branches, worktrees, and exclude files.
- Test that initialization creates the expected local files and that Git reports them as ignored rather than untracked.
- Test initialization idempotency, including duplicate exclude prevention and preservation of existing local files.
- Test task creation end-to-end: branch creation, worktree creation, AI-file provisioning, and clean Git status.
- Test multiple tasks to verify independent worktree paths and independent research directories.
- Test invalid names, branch collisions, existing paths, missing Git, and missing tmux.
- Test repository detection from both the main worktree and a linked worktree.
- Test tmux integration through a controllable fake executable or a dedicated integration environment; verify commands and starting directories without requiring an AI harness.
- Test template embedding by building the binary in tests and verifying that a fresh environment can provision files without access to the source repository.
- Test task removal protection for modified tracked files and explicit force behavior.
- Test cross-platform path and process behavior where practical, with platform-specific tmux expectations isolated from common workspace logic.

## Out of Scope

- Launching Codex, Claude, OpenCode, or any other AI harness.
- Managing AI credentials, providers, models, or API keys.
- Tracking or synchronizing research notes between worktrees.
- Committing, merging, pushing, or reviewing task branches.
- Automatically installing Git or tmux.
- Supporting non-Git repositories.
- Making ignored AI files available automatically to worktrees created by raw `git worktree add`; users should use `qbs task` for provisioned workspaces.
- Building a general-purpose plugin system for arbitrary AI harnesses in the first release.
- Synchronizing changes made to local `AGENTS.md`, `CLAUDE.md`, or skills back into the QBS templates automatically.

## Further Notes

The central product boundary is the provisioned workspace, not the AI harness. A task is considered ready when its branch, worktree, local AI files, and tmux pane exist; the developer remains responsible for selecting and launching a harness.

The first release should favor predictable behavior and a small command surface over extensive configuration. Harness-specific improvements can be added later without changing the core task-worktree workflow.
