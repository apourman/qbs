---
name: ship-staged
description: Ship work through a long-lived staging branch (`next` by default). Squash-merge a feature branch into staging, promote staging to the default branch with a merge commit and published patch notes, or hotfix the default branch, each through a GitHub pull request, and keep in-flight feature branches merged up with staging. Use when the user wants to ship, promote, release, or hotfix a branch, or sync branches with staging.
---

# ship-staged

```
master ◄── merge-commit PR ── next ◄── squash PR ── feature branch ◄── local merges ── ticket worktrees
```

- **Default branch**: what releases. Read it with
  `gh repo view --json defaultBranchRef -q .defaultBranchRef.name`.
- **Staging branch**: `next` unless the user names another. Long-lived; holds
  finished features waiting for the next release.
- **Feature branch**: one integration branch per feature, cut from
  `origin/<staging>`. Ticket worktrees merge into it locally through the
  implement and goal flows; it reaches staging as exactly one squash commit.

Each squash commit on staging becomes one changelog entry once staging is
promoted, so its title is a conventional commit describing the whole feature.

Pick the action from the request: **Ship** a feature branch into staging,
**Promote** staging to the default branch, **Hotfix** a branch straight to
the default branch, or **Sync** in-flight feature branches with staging.

## Preflight

```bash
git status --short --branch
git worktree list
gh auth status
git fetch origin
```

The branch being shipped is clean and fully committed. Show any uncommitted
changes and ask before committing them. If `origin/<staging>` does not exist,
offer to create it from `origin/<default>` and push it. If
`git merge-base --is-ancestor origin/<default> origin/<staging>` fails,
staging is missing work from the default branch, such as a release commit:
resync staging (below) before continuing.

## Ship

1. Invoke the `update-docs` skill (with the Skill tool) for the feature's
   work, so repository docs reach staging in the same pull request. Editing
   docs by hand does not replace it: the skill also walks `.artifacts/index.md`
   and republishes stale pages, which hand edits miss. Show the doc changes
   it makes and ask before committing them to the feature branch.
2. Merge `origin/<staging>` into the feature branch; stop and report on
   conflicts. A merge keeps the push fast-forward, and the squash discards the
   history anyway. Run the repository's tests.
3. Run the final review (below) and commit the fixes it calls for.
4. Hand the branch to the user for live testing. Nothing is pushed and no
   pull request exists yet.
   1. Set up the environment yourself, so the user only has to test:
      install dependencies, build, run migrations or seed data, and start
      any servers or watchers in the background (the `run` skill covers
      launching the app). Check each one is up, for example by hitting its
      URL or reading its log. Ask first for anything you cannot or should
      not do alone: secrets, paid or external services, a device, or a
      step that touches shared data.
   2. Give the user a checklist: the branch and worktree path; what is
      already running, with its URLs, ports, and log locations; any login
      or test data; anything left for them to do, with exact commands; what
      to try, covering the behavior the feature adds and anything nearby it
      could break; and how to stop what you started.
   3. Stop and wait. Passing tests and reviews do not replace this gate;
      only the user's explicit approval opens it. When they report a
      problem, fix it, commit, rerun the tests, restart what the fix
      affects, and hand off again. After approval, stop the processes you
      started.
5. Push and open the pull request:

   ```bash
   git push -u origin HEAD
   gh pr create --base <staging> --title "<type>: <feature summary>" --body-file <body>
   ```

   Body: a `## Summary` (why the feature exists, what it delivers) and a
   `## Changes` list.
6. Show the pull request URL and title, and ask before merging. Then:

   ```bash
   gh pr checks <number> --watch
   gh pr merge <number> --squash --delete-branch --subject "<title> (#<number>)"
   ```

7. Report the features now waiting on staging:
   `git log --oneline --no-merges origin/<default>..origin/<staging>`.
8. Record the merge in shared context: invoke `update-docs` again, limited
   to `.specs/`, `.backlog/`, `.research/`, and `.artifacts/`, so pages that
   list what waits on staging pick up this pull request. Mark the feature's
   tickets and backlog items `done` with the pull request link. Repository
   docs already shipped in step 1. In the final report, say that both
   `update-docs` runs happened, or which one was skipped and why.
9. Clean up. Candidates are the feature branch's worktree and the ticket
   worktrees whose branches are merged into the feature branch
   (`git branch --merged <feature>`). A squash merge hides the feature branch
   from `git branch -d`, so a candidate qualifies only when all hold:
   - `gh pr view <number> --json state,headRefOid` shows `MERGED`, and the
     feature branch tip equals `headRefOid`.
   - `git -C <worktree> status --porcelain` is empty.
   - The worktree was created for this feature; other sessions' and tools'
     worktrees stay.

   List the qualifying worktrees and branches, ask once, then run
   `git worktree remove <path>` and `git branch -D <branch>` for each. The
   worktree this session runs in stays: tell the user to archive the session
   so its host removes it.
10. Sync the other in-flight feature branches (below), so they build on the
   feature that just landed.

## Final review

Before a branch merges into staging or the default branch, review everything
it adds over its base. Invoke `code-review` and `ponytail-review` (with the
Skill tool) on the changes since `origin/<base>`; they are independent, so
run them in parallel. Show both reports side by side, then ask which findings
to fix. Commit the fixes to the branch and rerun the repository's tests. If a
skill is unavailable, say so and continue with the other. In the final report,
say that both reviews ran and what was fixed or deferred.

## Promote

1. Resync staging (below) so it contains every commit on the default branch.
2. List the waiting features with the `git log` from Ship step 7. Stop if
   there are none.
3. Write the patch notes (below).
4. Open the pull request with a fixed, non-conventional title:

   ```bash
   gh pr create --base <default> --head <staging> --title "Promote <staging> to <default>" --body-file <body>
   ```

   GitHub copies this title into the merge commit, and release tools such as
   release-please turn a `feat:` or `fix:` title into an extra changelog
   entry. Body: the waiting features, one per line, and the patch notes
   link.
5. Ask before merging. Then merge with a merge commit, which keeps each
   feature commit visible to the release tooling:

   ```bash
   gh pr checks <number> --watch
   gh pr merge <number> --merge
   ```

   If the repository disallows merge commits, stop and tell the user.
6. Resync staging.
7. If release-please (or similar) opens a release pull request, tell the user
   to check that each feature appears once in its changelog before merging
   it, and to resync staging after it merges.

## Patch notes

Promote only: a Ship into staging or a Hotfix writes none.

1. Find the last release: `git describe --tags --abbrev=0 origin/<default>`.
   The range is `<tag>..origin/<staging>`; if the command finds no tag, use
   `origin/<staging>` alone and call the range "all history".
2. Gather every change in the range, including hotfixes already on the
   default branch: `git log --no-merges --format='%h %s' <range>`, plus
   `gh pr view <number> --json title,body` for the pull requests those
   commits name.
3. Load `artifact-design`, then write
   `.artifacts/releases/<YYYY-MM-DD>-<staging>-to-<default>.md` in the
   primary checkout, for people who use the project rather than its
   maintainers. A `# Patch notes` heading with the date, the range, and the
   page version (`v1`, bumped on each edit), then `## New`, `## Improved`,
   and `## Fixed`, leaving out empty sections. One plain-language bullet per
   change, saying what users can now do or what no longer goes wrong, ending
   with its pull request link, or its commit link when it has no pull
   request. Rewrite commit titles; never paste them. Leave out changes
   nobody outside the repository notices (tests, CI, refactors, chores).
4. Show the notes and ask for changes.
5. Publish them as a Claude Artifact: this skill asks for a Markdown
   artifact, so publish the `.md` file itself. When `.artifacts/index.md`
   already lists the file, read the live version first, keep the newer
   copy, and republish to its URL. Add or update its index line: title,
   path, "patch notes for this promotion", URL, owning account, and
   version. If the Artifact tool is unavailable, update the local file and
   its index version, and say it needs publishing from the owning account.

## Hotfix

Cut the branch from `origin/<default>`, then follow Ship with the default
branch in place of staging throughout. Resync staging after it merges.

## Sync

Merge `origin/<staging>` into every in-flight feature branch so each one
builds on what has already shipped. Run it after each Ship, or whenever the
user asks. Merge, never rebase: rebasing would force-push the branch and break
ticket worktrees, and the squash at ship time discards the merge commits.

1. `git fetch origin`, then find the feature branches: local branches other
   than staging and the default branch, with commits not on
   `origin/<staging>`, whose upstream is not `[gone]` (`git branch -vv`).
   Leave ticket branches; they pick up staging through their feature branch.
   Ask when it is unclear which branches are features.
2. For each, in the worktree where it is checked out (or a temporary one):
   skip it if `git status --porcelain` is not empty, otherwise run
   `git merge origin/<staging>`. On conflicts, run `git merge --abort` and
   move on.
3. Report each branch as updated, already current, skipped as dirty, or
   conflicting. Do not push; each branch reaches GitHub when it ships.

## Resync staging

Run after anything lands on the default branch: a promotion, a hotfix, or a
release commit. Use the worktree where staging is checked out (`git worktree
list`), or add a temporary one.

```bash
git fetch origin
git merge --ff-only origin/<staging>
git merge origin/<default>
git push origin <staging>
```

The merge fast-forwards when staging holds nothing new. Staging keeps its
commit SHAs, since the default branch already contains them, so it is updated
by merge only. Stop and report on conflicts.

## Safety rules

- Never push or open a pull request for a Ship or Hotfix before the user
  approves the live test.
- Ask before every merge into staging or the default branch, any force-push,
  and deleting branches or worktrees.
- Stop on conflicts or failed checks and report the state.
- Preserve unrelated user changes.
