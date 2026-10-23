---
name: backlog-session
description: Start a session that works one `.backlog/` item end to end and ships it with `ship-staged`. Picks the easiest open item unless told otherwise. Use when the user wants to work on a backlog item or ticket, or asks for something easy to pick up.
---

# backlog-session

One session, one backlog item, built on staging from the first edit and
shipped through it. The user's words after the skill name narrow the pick: an
item number, a `kind` or `area`, or a size ("easy" is the default; "meaty"
flips it).

## 1. Load ship-staged

Before anything else, invoke `ship-staged` (with the Skill tool) and run its
**Preflight**. Every step below follows its branch model: work starts on a
feature branch cut from `origin/<staging>` (`next` by default) and reaches
staging through its **Ship** action.

## 2. Pick

Read every item in the primary checkout's `.backlog/items/` (see
`docs/agents/backlog.md` for the format). Candidates have `status: open`.
Rank them by **ease**: a clear repro or ask, one `area`, no spec or design
decision needed, a small touched surface you can confirm by a quick look at
the code. Skip items already `doing`; another session may own them.

Ask once with the question tool: up to four candidates, easiest first and
marked recommended, each with its number, title, and a one-line reason it is
easy. If no item is open, say so and stop.

## 3. Branch

Before the first code edit:

1. Cut the feature branch from `origin/<staging>`, named after the item, e.g.
   `fix/012-chat-coordinates`. When this session already runs in a worktree,
   cut the branch there.
2. Set the item to `status: doing`.
3. Rename the session to `Backlog <NNN>: <title>` with the harness's
   session-title tool, so the session list shows which item each session
   owns. When no such tool exists, skip this.

## 4. Build

Work the item as `docs/agents/backlog.md` describes, using the `ponytail`
skill and the repository's tests. When the item turns out larger than it
looked (it needs a spec, or a decision only the user can make), stop and
report: offer to record findings on the item and return it to `open`, or to
continue.

## 5. Ship

Run ship-staged's **Ship** action on the feature branch. It owns docs, the
final review, the pull request, merging, and marking the item `done`; follow
its asks.
