# Backlog

`.backlog/items/` holds the project's work: bugs, play-test and user
reports, feature ideas, polish, and the slices of each spec. Like `.specs/`,
it is shared, Git-ignored context in the primary checkout. Run `qbs backlog`
to browse and edit it at http://localhost:4317.

## Item format

Each item is one file, `.backlog/items/<NNN>-<slug>.md`, where `NNN` is the
next unused number padded to three digits. Create it with
`qbs backlog new "<title>"`, which claims the number safely when several
sessions add items at once and prints the new file's path; then fill in the
file:

```markdown
---
title: Clicking coordinates in chat no longer sets the arrow
kind: bug
area: chat
blocked-by: 009, 011
status: open
source: play-test, 21 Oct 2026
created: 2026-10-21
---

What was seen or asked for, with links and screenshots.
```

- `status` is one of `open`, `doing`, `done`, or `dropped`.
- `kind` and `area` are free text. Reuse values that other items already use.
  When the item belongs to a domain or feature, set `area` to the directory
  name used in `.specs/` and `.research/`.
- `blocked-by` is optional: the numbers of items that must be `done` before
  this one starts. Leave it out when nothing blocks the item.
- Store images in `.backlog/items/assets/` as `<NNN>-<n>.<ext>` and embed them
  with a relative path, such as `![Arrow stays put](assets/012-1.png)`, so they
  render in `qbs backlog` and any Markdown preview.
- Never renumber or rename item files. Other records refer to items by number,
  for example "Backlog 012".

## Working an item

- Read the item first, and set it to `doing` when work starts.
- Keep the original report. Add what you learn under a dated
  `## Findings (<date>)` heading. Put long notes in
  `.research/<area>/<slug>.md` and link to them.
- Do not start an item while any item in its `blocked-by` is not `done`.
- A small fix can go straight from the item. For larger work, write
  `.specs/<area>/spec.md`, then add its slices as items in that area, with
  `blocked-by` giving their order. Link the spec from the original item.
- When an item's area has a spec, read the spec before working the item.
- When the work is finished, set the item to `done` and say what resolved it,
  such as a commit or PR. Set an item to `dropped` only with a stated reason.

## Specs and areas

A spec's area is its `.specs/` directory name. `.specs/<area>/spec.md` is the
design; the items with `area: <area>` are its to-do list. Specs get no
separate ticket files. Older `.specs/<area>/tickets/` folders are legacy
context: leave finished ones in place, and convert open ones to items only
when asked.
