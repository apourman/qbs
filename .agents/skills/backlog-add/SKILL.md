---
name: backlog-add
description: Add one or more items to `.backlog/` as structured Markdown, with any screenshots or images saved beside them and embedded so they show in the rendered file. Use when the user reports a bug, idea, or play-test note to log, or asks to add something to the backlog.
---

# backlog-add

Turn what the user said (and any images they gave) into backlog items in the
primary checkout's `.backlog/items/`, following the format in
`docs/agents/backlog.md`. Capture only: no fixing, no spec.

## 1. Split

Read the user's words after the skill name and the conversation. Each distinct
bug, ask, or idea becomes its own item. When it is unclear whether two notes
are one item, keep them as one.

## 2. Match

Read the existing items' frontmatter. Reuse their `kind` and `area` values,
and set `area` to the `.specs/` or `.research/` directory name when one fits.
When an open or `doing` item already covers a note, add to that item under a
dated `## Findings (<date>)` heading instead of creating a duplicate, and say
so.

## 3. Write

Give each new item the next unused number and a short kebab-case slug, then
write it with every frontmatter field filled: a plain-language `title`,
`status: open`, `source` (who or what reported it, and when), and `created`
as today's date. Structure the body under these headings, leaving out any
with nothing to say:

```markdown
## What happened
What was seen or asked for, in the reporter's terms.

## Expected
What should happen instead.

## Steps to reproduce
1. Numbered steps, when known.

## Screenshots
![What the image shows](assets/012-1.png)

## Notes
Links, environment, guesses about cause, related items ("Backlog 007").
```

Keep the reporter's own words where they matter; your summary goes in the
title.

## 4. Images

Save every image the user gave to `.backlog/items/assets/` as
`<NNN>-<n>.<ext>` (item number, then 1, 2, … per item), keeping the original
extension, and embed it with the relative path `assets/<NNN>-<n>.<ext>` and alt
text describing what it shows. Copy from the path the user gave. When an image
was pasted with no file path you can reach, ask the user for its path or to
save it somewhere, and write the item meanwhile with a `TODO: screenshot`
line in its place.

## 5. Report

Done when every note is in an item and every embedded image path exists on
disk. List each item as `Backlog NNN: <title>` with its file path, plus any
item you appended to instead.
