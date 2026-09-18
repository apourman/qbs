# Install QBS

QBS prepares local AI workspaces and manages a user-wide catalog of curated
skills. Go is only needed to build QBS from source. Git is required for
project commands.

## Download a release artifact

Choose the release archive matching the host operating system and CPU. The
local `make release-artifacts` command writes the binaries named below to
`dist/`; published releases package each binary in a `.tar.gz` archive, or a
`.zip` archive on Windows.

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

## Install or update from a release

With Go installed, the simplest global installation is:

```text
go install github.com/trues/qbs/cmd/qbs@latest
```

Go places the executable in your Go binary directory (`GOBIN`, or otherwise
`GOPATH/bin`). Add that directory to `PATH` if `qbs` is not found. To update
an existing installation to the newest release, run:

```text
qbs update
```

QBS downloads the matching release, verifies its checksum, and replaces the
installed executable. This self-update command currently supports Linux and
macOS. On Windows, download the new release archive and replace `qbs.exe`
after closing QBS.

To install a specific release, replace `latest` with its tag, for example
`@v0.0.1`.

The installed executable reports the release selected by Go:

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

## Project initialization

Initialize the current repository with local instructions, specifications,
research storage, and shared Git excludes:

```text
qbs init
```

Initialization creates `.research/` for local research and `.specs/` for local
specifications. It also generates the shared QBS agent catalog into the native
project directories for Codex, OpenCode, and Claude Code. These local files are
ignored by the repository's shared Git exclude file and refreshed by later
`qbs init` runs. A same-named unmanaged agent is preserved
with a warning, and unrelated files in those directories are left alone. The
tracked source of truth is `internal/agents/definitions.yaml` in QBS itself.

Initialization also creates local engineering guidance in `docs/agents/`,
including the issue-tracker and domain-documentation configuration. GitHub
remotes receive the GitHub Issues guidance; repositories without a supported
GitHub remote receive a conservative local-Markdown fallback. Existing files
are preserved, and the directory is excluded through the repository-local Git
common exclude file.

## Global curated skills

Import either one skill directory containing `SKILL.md`, or a collection whose
immediate child directories each contain `SKILL.md`:

```text
qbs skills import /path/to/skill
qbs skills import /path/to/collection
```

The canonical catalog lives at `~/.qbs/skills/`. By default, imports are
synchronized to Codex's shared global agent skill directory, `~/.agents/skills/`,
as well as Claude's `~/.claude/skills/` and OpenCode's
`~/.config/opencode/skills/`. QBS marks its synchronized copies and refuses to
replace an unmanaged skill with the same name. Use `--force` only to replace an
existing catalog entry; it does not override unmanaged destinations.

```text
qbs skills list
qbs skills sync
qbs skills remove <name>
```

Set `QBS_HOME` to relocate the catalog. Set `QBS_SKILL_TARGETS` to an
OS-path-list of global directories when harness discovery paths differ from the
defaults. Curated skills remain global; project-only skills may still live in
the project-local harness directories provisioned and excluded by QBS.

## Build release artifacts locally

The project uses Conventional Commits and release-please for published
releases. Merges to `master` are analyzed to maintain a release PR, which is
automatically merged into `master` after updating `VERSION` and
`CHANGELOG.md`. The follow-up workflow creates a `v<version>` tag with a draft
release. GoReleaser then builds the tagged binaries for Linux, macOS, and
Windows on amd64 and arm64 and publishes the release after uploading them.

For local packaging, from a checkout with Go installed, run:

```text
make release-artifacts
```

This produces standalone, CGO-free binaries for Linux, macOS, and Windows,
plus `dist/SHA256SUMS`. The version is embedded in each binary and is shown by
`qbs --version`. Run `make smoke-release` to build the artifacts and exercise
initialization and skill synchronization from a fresh temporary environment.
