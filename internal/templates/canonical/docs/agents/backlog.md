# Backlog

`.backlog/items/` holds the project's intake: bugs, play-test and user
reports, feature ideas, and polish that are not yet part of a spec. Like
`.specs/`, it is shared, Git-ignored context in the primary checkout. Run
`qbs backlog` to browse and edit it at http://localhost:4317.

## Item format

Each item is one file, `.backlog/items/<NNN>-<slug>.md`, where `NNN` is the
next unused number padded to three digits:

```markdown
---
title: Clicking coordinates in chat no longer sets the arrow
kind: bug
area: chat
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
- Never renumber or rename item files. Other records refer to items by number,
  for example "Backlog 012".

## Working an item

- Read the item first, and set it to `doing` when work starts.
- Keep the original report. Add what you learn under a dated
  `## Findings (<date>)` heading. Put long notes in
  `.research/<area>/<slug>.md` and link to them.
- A small fix can go straight from the item. For larger work, create a spec or
  ticket as `docs/agents/issue-tracker.md` describes. Link the spec or ticket
  from the item, and the item from the spec or ticket.
- When the work is finished, set the item to `done` and say what resolved it,
  such as a commit, PR, or ticket. Set an item to `dropped` only with a stated
  reason.

A spec's `spec.backlog.md` is separate. It holds work deferred from that spec,
while `.backlog/` holds intake that no spec owns yet.
