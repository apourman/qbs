# 02: Implement repository initialization

**What to build:** `qbs init` prepares the current Git repository for local-only Agentic AI tooling.

**Blocked by:** 01: Create the Go CLI foundation.

**Status:** ready-for-agent

- [ ] Initialization adds the AI tooling paths to the repository's local Git exclude configuration.
- [ ] Re-running initialization does not duplicate exclude entries.
- [ ] Initialization provisions shared agent instruction files.
- [ ] Initialization provisions skills for Codex, Claude, and OpenCode.
- [ ] Initialization creates local research storage.
- [ ] Existing local instruction files are not silently overwritten.
- [ ] All provisioned files are ignored by Git and do not appear as untracked changes.
- [ ] Initialization behavior is covered by temporary-repository integration tests.
