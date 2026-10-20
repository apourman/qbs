---
name: ship-staged
description: Ship work through a long-lived staging branch (`next` by default). Squash-merge a feature branch into staging, promote staging to the default branch with a merge commit, or hotfix the default branch, each through a GitHub pull request. Use when the user wants to ship, promote, release, or hotfix a branch.
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
**Promote** staging to the default branch, or **Hotfix** a branch straight to
the default branch.

## Preflight

```bash
git status --short --branch
git worktree list
gh auth status
git fetch origin
```

The branch being shipped is clean and fully committed. Show any uncommitted
changes and ask before committing them. If `origin/<staging>` does not exist,
offer to create it from `origin/<default>` and push it.

## Ship

1. Merge `origin/<staging>` into the feature branch; stop and report on
   conflicts. A merge keeps the push fast-forward, and the squash discards the
   history anyway. Run the repository's tests.
2. Push and open the pull request:

   ```bash
   git push -u origin HEAD
   gh pr create --base <staging> --title "<type>: <feature summary>" --body-file <body>
   ```

   Body: a `## Summary` (why the feature exists, what it delivers) and a
   `## Changes` list.
3. Show the pull request URL and title, and ask before merging. Then:

   ```bash
   gh pr checks <number> --watch
   gh pr merge <number> --squash --subject "<title> (#<number>)"
   ```

4. Report the features now waiting on staging:
   `git log --oneline --no-merges origin/<default>..origin/<staging>`.

## Promote

1. Resync staging (below) so it contains every commit on the default branch.
2. List the waiting features with the `git log` from Ship step 4. Stop if
   there are none.
3. Open the pull request with a fixed, non-conventional title:

   ```bash
   gh pr create --base <default> --head <staging> --title "Promote <staging> to <default>" --body-file <body>
   ```

   GitHub copies this title into the merge commit, and release tools such as
   release-please turn a `feat:` or `fix:` title into an extra changelog
   entry. Body: the waiting features, one per line.
4. Ask before merging. Then merge with a merge commit, which keeps each
   feature commit visible to the release tooling:

   ```bash
   gh pr checks <number> --watch
   gh pr merge <number> --merge
   ```

   If the repository disallows merge commits, stop and tell the user.
5. Resync staging.
6. If release-please (or similar) opens a release pull request, tell the user
   to check that each feature appears once in its changelog before merging
   it, and to resync staging after it merges.

## Hotfix

Cut the branch from `origin/<default>`, then follow Ship with the default
branch in place of staging throughout. Resync staging after it merges.

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

- Ask before every merge into staging or the default branch, any force-push,
  and deleting branches or worktrees.
- Stop on conflicts or failed checks and report the state.
- Preserve unrelated user changes.
