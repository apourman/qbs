# Issue tracker

Local Markdown issue records under `.specs/<domain-or-feature>/issues/` are the
authoritative source for project work, regardless of which repository remotes
are configured.

## Standard operations

- List issues by browsing the domain's `issues/` directory.
- Read an issue file before changing or implementing it.
- Create a ticket as the next available `<NN>-<ticket-slug>.md` file after
  confirming the requested scope and linking the relevant research or spec.
- Update the issue file when its scope, acceptance conditions, ownership, or
  status changes.
- Record progress, decisions, and requested follow-up in the issue file rather
  than assuming a hosted comment stream exists.
- Mark completed or declined work in the issue's status and preserve the
  closing rationale in the file.

## Repository workflow

Keep research in `.research/<domain-or-feature>/` and specs and their tickets
in the matching `.specs/<domain-or-feature>/` directory. Use the same stable
domain-or-feature name across those artifacts. This local workflow performs no
remote publishing and provides no remote comments, labels, assignees, or
notifications.

Treat GitHub, Jira, Linear, GitLab, and other hosted trackers only as optional
projection targets. Publishing or synchronizing a local record requires a
separate, explicit request and is not part of initialization.
