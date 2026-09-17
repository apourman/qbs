---
name: ship-current-changes
description: Finish the current worktree branch and merge it back into its base branch locally or through a GitHub pull request.
---

# ship-current-changes

Use this skill when work has been completed in the current worktree or topic
branch and needs to be delivered to a base branch.

Choose one delivery mode:

- **Local:** rebase the work branch onto the base branch, then fast-forward the
  base branch.
- **GitHub:** push the branch, create and merge a pull request targeting the
  base branch, then update the local base branch.

## Workflow

### 1. Inspect the worktree

Before changing anything, inspect:

```bash
git status --short --branch
git diff
git diff --cached
git log --oneline -10
git worktree list
```

Treat staged, unstaged, and untracked changes as candidates for this shipment.
Review all of them. If anything appears unrelated, sensitive, or generated,
ask whether to include it before staging.

Record the current branch:

```bash
work_branch="$(git branch --show-current)"
```

Ask the user which base branch should receive the changes. Use the current
branch's upstream only as a suggestion. Do not guess based only on the branch
name. Confirm the selected base branch before staging or delivering anything.

### 2. Choose delivery mode

Ask whether to deliver locally or through GitHub:

- **Local merge** when delivery should happen entirely on the machine.
- **GitHub pull request** when delivery should happen through a GitHub remote.

Do not require GitHub CLI authentication for a local merge.

### 3. Finish the worktree branch

Stage the reviewed current changes, including staged, unstaged, and untracked
files. Review the staged diff for secrets, local environment data, generated
noise, and unrelated changes before committing.

Commit the intended changes using a concise conventional commit message:

```text
<type>: <short imperative summary>
```

Run the relevant tests before delivery. Do not amend existing commits unless
the user explicitly asks.

### 4. Deliver locally

Find the worktree where the base branch is checked out using `git worktree
list`. If it is a different worktree, perform the merge there. Otherwise use
the current worktree only after ensuring it is safe to switch branches.

The base worktree must be clean before integration. Rebase the work branch onto
the selected base branch:

```bash
git rebase <base-branch>
```

If the rebase has conflicts, stop and report them. Do not resolve, reset, or
discard changes without the user's direction. After a successful rebase,
fast-forward the base branch:

```bash
git switch <base-branch>
git merge --ff-only <work-branch>
```

After a successful fast-forward, report the base branch and new tip. Leave the
repository on the base branch.

### 5. Deliver through GitHub

Confirm that `gh` is installed and authenticated, and that the repository has
an `origin` remote:

```bash
gh auth status
git remote get-url origin
```

Push the worktree branch:

```bash
git push -u origin HEAD
```

Create a pull request targeting the explicitly confirmed base branch:

```bash
body_file="$(git rev-parse --git-path qbs-pr-body.md)"
gh pr create --base <base-branch> --title "<type>: <short summary>" --body-file "$body_file"
rm -f "$body_file"
```

The pull request body should contain only:

```md
## Summary

- Why the branch exists.
- What outcome it delivers.

## Changes

- Meaningful final change 1.
- Meaningful final change 2.
```

After the pull request is created, wait for its checks and merge it:

```bash
gh pr checks <number> --watch
gh pr merge <number> --squash
```

Then update the local base branch in its worktree:

```bash
git fetch origin <base-branch>
git pull --ff-only origin <base-branch>
```

Report the pull request URL and the synchronized local base branch.

### 6. Stay within scope

Only finish and deliver the current branch. Ask before force-pushing, resetting,
deleting branches or worktrees, or changing unrelated files. Rebasing is part of
the local delivery flow after the user has selected that mode and base branch.

## Safety rules

- Inspect the complete worktree before changing it.
- Preserve unrelated user changes.
- Confirm the base branch; never infer it from a naming convention alone.
- Include all reviewed current changes in the shipment.
- Keep the base worktree clean before a local merge.
- Use `gh` only for GitHub delivery.
- Stop on merge conflicts or failed checks and report the state.
