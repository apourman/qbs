# Install QBS

QBS prepares local AI workspaces and manages a user-wide catalog of curated
skills. Go is only needed to build QBS from source. Git is required for
project commands, and tmux is required only by the standalone `qbs task`
workflow.

## Download a release artifact

Choose the artifact matching the host operating system and CPU from `dist/`:

| Host | Artifact |
| --- | --- |
| Linux x86-64 | `qbs_<version>_linux_amd64` |
| Linux ARM64 | `qbs_<version>_linux_arm64` |
| macOS Intel | `qbs_<version>_darwin_amd64` |
| macOS Apple Silicon | `qbs_<version>_darwin_arm64` |
| Windows x86-64 | `qbs_<version>_windows_amd64.exe` |
| Windows ARM64 | `qbs_<version>_windows_arm64.exe` |

Verify the download against `SHA256SUMS`, then place the executable on
`PATH`. On macOS and Linux, mark it executable with `chmod +x`. On Windows,
place the `.exe` in a directory on `PATH`.

Check the installed build with:

```text
qbs --version
```

To build and install QBS into your user-local bin directory on Linux or macOS:

```text
make install PREFIX="$HOME/.local"
export PATH="$HOME/.local/bin:$PATH"
```

For a system-wide install, use `PREFIX=/usr/local` with the permissions
required by your system.

To update the local installation from the current `main` checkout:

```text
make update-local
```

This rebuilds QBS into `~/.local/bin/qbs` and synchronizes the global curated
skill catalog. To run that update automatically after commits on `main`, opt
in once with:

```text
make enable-local-updates
```

The hook is local to this Git checkout. It never prevents a commit from
completing; if an update fails, it prints a warning and can be rerun manually.
Set `QBS_UPDATE_BRANCH=master` when the checkout uses `master` instead of
`main`.

## Project and worktree provisioning

Initialize the current repository with local instructions, specifications,
research storage, and shared Git excludes:

```text
qbs init
```

Initialization creates `.research/` for local research and `.specs/` for local
specifications. It also generates the shared QBS agent catalog into the native
project directories for Codex, OpenCode, and Claude Code. These local files are
ignored by the repository's shared Git exclude file and refreshed by later
`qbs init` or `qbs provision` runs. A same-named unmanaged agent is preserved
with a warning, and unrelated files in those directories are left alone. The
tracked source of truth is `internal/agents/definitions.yaml` in QBS itself.

Provision an existing worktree created by Firstmate, Treehouse, or another
tool without asking QBS to create or open it:

```text
qbs provision /path/to/worktree
```

QBS retains its standalone task workflow for use without an external
orchestrator:

```text
qbs task <name>
qbs tasks
qbs task remove <name>
```

Do not have QBS and another orchestrator create a worktree for the same task.
Let the orchestrator own the task lifecycle and use `qbs provision` after its
worktree exists.

## Global curated skills

Import either one skill directory containing `SKILL.md`, or a collection whose
immediate child directories each contain `SKILL.md`:

```text
qbs skills import /path/to/skill
qbs skills import /path/to/collection
```

The canonical catalog lives at `~/.qbs/skills/`. Imports are synchronized to
the global skill directories for Codex, Claude, and OpenCode. QBS marks
its synchronized copies and refuses to replace an unmanaged skill with the
same name. Use `--force` only to replace an existing catalog entry; it does not
override unmanaged destinations.

```text
qbs skills list
qbs skills sync
qbs skills remove <name>
```

Set `QBS_HOME` to relocate the catalog. Set `QBS_SKILL_TARGETS` to an
OS-path-list of directories when harness discovery paths differ from the
defaults. Curated skills remain global; project-only skills may still live in
the project-local harness directories provisioned and excluded by QBS.

## Build release artifacts

From a checkout with Go installed:

```text
VERSION=0.2.0 make release
```

This produces standalone, CGO-free binaries for Linux, macOS, and Windows,
plus `dist/SHA256SUMS`. The version is embedded in each binary and is shown by
`qbs --version`. Run `make smoke-release` to build the artifacts and exercise
initialization and task creation from a fresh temporary environment.
