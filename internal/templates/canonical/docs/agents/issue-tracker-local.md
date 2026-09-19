# Issue tracker

Local Markdown ticket records under `.specs/<domain-or-feature>/tickets/` are the
authoritative source for project work, regardless of which repository remotes
are configured.

## Standard operations

- List tickets by browsing the domain's `tickets/` directory.
- Read a ticket file before changing or implementing it.
- Create a ticket as the next available `<NN>-<ticket-slug>.md` file after
  confirming the requested scope and linking the relevant research or spec.
- Update the ticket file when its scope, acceptance conditions, ownership, or
  status changes.
- Record progress, decisions, and requested follow-up in the ticket file rather
  than assuming a hosted comment stream exists.
- Mark completed or declined work in the ticket's status and preserve the
  closing rationale in the file.

## Repository workflow

Keep research in `.research/<domain-or-feature>/` and specs and their tickets
in the matching `.specs/<domain-or-feature>/` directory. Use the same stable
domain-or-feature name across those artifacts. This local workflow performs no
remote publishing and provides no remote comments, labels, assignees, or
notifications.

## Spec future-work backlogs

Each specification may have one companion backlog beside it, named by adding
`.backlog.md` to the spec filename. For example, `workflow.md` may have
`workflow.backlog.md` in the same `.specs/<domain-or-feature>/` directory.
Keep backlog entries as ordinary unchecked Markdown tasks. They record deferred
or possible future work and are not implementation tickets or automatic
`ready-for-agent` work.

When a backlog item becomes actionable, create a separate numbered ticket
under the spec's `tickets/` directory and link that ticket from the item. If
the idea came from another ticket, preserve that originating-ticket link too.
This keeps the local specification and ticket hierarchy authoritative without introducing
a central backlog or hosted synchronization.

Treat GitHub, Jira, Linear, GitLab, and other hosted trackers only as optional
projection targets. Publishing or synchronizing a local record requires a
separate, explicit request and is not part of initialization.
