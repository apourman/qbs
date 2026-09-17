---
name: implement
description: "Implement a piece of work based on a spec or set of tickets."
disable-model-invocation: true
---

Implement the work described by the user in the spec or tickets.

Use /tdd when test-first work helps, at pre-agreed seams. Add only tests that
protect meaningful behavior or a regression; prefer existing public interfaces
and avoid covering the same behavior at multiple layers.

When the work repeats operational logic across flows, call `codebase-design`
to choose the seam and separate caller policy from shared mechanics. Do not add
a layer or directory solely because two files look similar.

Use focused typechecking and tests as the work progresses. Run the broader suite
when required by the repository or when focused checks leave material risk.
Keep roughly 25% of an available task budget for review, fixes, CI, and
integration; distinguish an enforceable token limit from account quota and
monetary cost.

Once done, use /code-review to review the work.

Work in an isolated branch and worktree when this task is dispatched as part of
a larger implementation. Create the branch and worktree yourself when one was
not supplied. Commit the completed work and report the revision. Clean up the
worktree only after the revision has been integrated.
