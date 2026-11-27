---
name: update-readmes
description: Bring the READMEs touched by the current work up to the README standard — short, current, one README per module, each parent linking to its children's READMEs — walking from the changed directories up to the repository root. Use when the user says "update the readmes", asks for a README standard or a README cleanup, or a README has grown bloated with history or agent notes.
---

# update-readmes

A README is the **front door** of one module: it tells a reader who just
cloned the repository what the module is, how to use it, and where to go
next. It describes the module as it stands today. Its history lives in git
and the changelog; notes for agents working locally live beside it.

## The standard

**One README per module.** A module is a directory a reader would open on its
own: a package, a project with its own build file, an app, a tool. Plain
source folders inside a module get no README.

**Parents link down.** A README whose directory holds modules lists them in
a table, one row each: the module's name, linked to its README when it has
one, and one line on what it is for. Everything else about a child module
lives in the child's README, so each fact has one home.

**Sections follow the reader's path.** Open with one paragraph: what the
module is, who it is for, and what it needs to run. Then only the sections
the module has a reader for, in this order: Install or Use, Options, Modules
(the table above), Build, Tests, Releases. Each section is the few lines a
reader acts on: a command, a setting, a table. A section that outgrows a
screen moves to a child README or `docs/`, linked from one line.

**Present tense, current state.** An unfinished module says so in one
sentence of status: what works now and what is still missing.

**What moves out**, and where it goes:

| Material | Home |
| --- | --- |
| How the module got here: spikes, tried and dropped designs, "began as" | git history; the changelog for released changes |
| Backlog, ticket and issue numbers | the backlog item or issue itself |
| A file-by-file tour of the source | doc comments in the code |
| Notes for agents or developers working locally: hooks, test seams, gotchas, environment variables for one-off checks | an `AGENTS.md` beside the README, linked from one line at the README's end |

`.backlog/`, `.research/`, `.specs/` and `.artifacts/` are Git-ignored, so a
README never links into them: a fresh clone has no such files.

## 1. Scope

List the directories the work changed:
`git diff --name-only <base>...HEAD` plus the uncommitted changes, where
`<base>` is the branch the work will merge into. When the user names a
directory instead, the scope is every module under it.

For each changed directory, take the README of its nearest module, then each
ancestor README up to the repository root. A changed directory inside a new
module with no README gets one when it meets the module test above. Ancestors
are in scope only for the rows and links that point at their children.

## 2. Rewrite, deepest first

Work through the scoped READMEs from the deepest directory up, so each parent
links to a child that is already finished. For each one:

- Bring every fact the work changed up to date, checking it against the code,
  build files and `--help` output rather than the old text.
- Move out what the table above sends elsewhere. When an `AGENTS.md` already
  sits beside the README, add to it; keep only what is still true.
- Add, fix or remove the module rows for its children.

## 3. Check links

Every relative link in each README you touched resolves to a file in the
repository, and every `#anchor` matches a heading in its target.

## 4. Report

Done when every scoped README meets the standard and every link resolves.
List each README and `AGENTS.md` created or changed, with a line on what moved
out and where. Leave the changes uncommitted unless the user asked for a
commit.
