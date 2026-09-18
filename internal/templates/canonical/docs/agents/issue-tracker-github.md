# Issue tracker

This repository uses GitHub Issues. Run the `gh` CLI from the repository root;
it discovers the repository from the configured GitHub `origin` remote. Check
authentication and repository access before attempting a remote operation.

## Standard operations

- Check the current tracker state with `gh issue status`.
- List open issues with `gh issue list`; use `--state all`, `--label`, or
  `--assignee` when narrowing the list.
- Read an issue with `gh issue view <number>` before changing or implementing
  it.
- Publish a new issue with `gh issue create --title <title> --body-file <file>`
  after confirming the requested scope and preparing the local context.
- Edit an issue with `gh issue edit <number> --title <title> --body-file <file>`
  or update its assignee and labels with the corresponding `gh issue edit`
  options.
- Add progress or decision context with `gh issue comment <number> --body-file <file>`.
- Close completed or declined work with `gh issue close <number> --comment <reason>`;
  reopen it with `gh issue reopen <number>` when the work becomes actionable
  again.

## Repository workflow

Keep the related research note or spec under `.research/` or `.specs/`, and
link that path from the issue body. Before publishing, make sure the title,
scope, acceptance conditions, and requested labels are clear. When working on
an existing issue, read it first, keep material decisions in comments, and
update the issue when its status changes. Use the repository's existing labels;
do not invent or create remote labels during provisioning or triage.
