# Issue tracker

Backlog items in `.backlog/items/` are the authoritative tickets for project
work, regardless of which repository remotes are configured. A spec's tickets
are the items whose `area` is the spec's `.specs/` directory name.
`docs/agents/backlog.md` owns the item format, `blocked-by` ordering, and
numbering.

## Standard operations

- List a spec's tickets by filtering items on its `area`, or with
  `qbs backlog`.
- Read an item before changing or implementing it, and read its area's spec
  when one exists.
- Create a ticket with `qbs backlog new "<title>"` after confirming the
  requested scope, then link the relevant research or spec from it.
- Update the item when its scope, acceptance conditions, ownership, or status
  changes.
- Record progress, decisions, and requested follow-up in the item rather than
  assuming a hosted comment stream exists.
- Mark completed or declined work `done` or `dropped` and preserve the closing
  rationale in the item.

## Repository workflow

Keep research in `.research/<domain-or-feature>/`, specs in the matching
`.specs/<domain-or-feature>/` directory, and their tickets as items with that
`area`. Use the same stable domain-or-feature name across those artifacts.
This local workflow performs no remote publishing and provides no remote
comments, labels, assignees, or notifications.

## Spec future-work backlogs

Each specification may have one companion backlog beside it, named by adding
`.backlog.md` to the spec filename. For example, `workflow.md` may have
`workflow.backlog.md` in the same `.specs/<domain-or-feature>/` directory.
Keep backlog entries as ordinary unchecked Markdown tasks. They record deferred
or possible future work and are not implementation tickets or automatic
`ready-for-agent` work.

When an entry becomes actionable, create a backlog item in the spec's area
and link the item from the entry. If the idea came from another item,
preserve that originating-item link too.

Treat GitHub, Jira, Linear, GitLab, and other hosted trackers only as optional
projection targets. Publishing or synchronizing a local record requires a
separate, explicit request and is not part of initialization.
