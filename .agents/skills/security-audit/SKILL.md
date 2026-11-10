---
name: security-audit
description: Audit an entire codebase for security vulnerabilities, including vulnerable dependencies and secrets in git history, and write a private report. Use when the user wants a security audit or assessment of existing code. For only a branch's changes, use `review-security`.
---

Audit all the code at one pinned commit, try to disprove every finding, and
write a private report. The audit leaves the code untouched: the report is the
only file it adds to the project.

## 1. Pin and divide

Run `git rev-parse HEAD`, `git status --short`, and `git ls-files`. The audit
covers tracked files at that commit; note any uncommitted changes as outside it.

Group the tracked files into areas of one concern each, small enough for one
agent to read closely. The step is done when every tracked file is in an area or
on a skip list with a reason (vendored, generated, or binary only). With more
than 20 areas, show the list and ask before continuing, since each area costs an
agent.

## 2. Ask once about scanners

Check which of these are installed:

- **Dependencies**: `osv-scanner`, which covers most languages, or the project's
  own tool (`npm audit`, `pip-audit`, `cargo audit`, `govulncheck`).
- **Secrets in git history**: `gitleaks`.

Ask once whether to install what is missing and run them, saying that dependency
scanners send package names and versions to a public vulnerability database. Use
each tool's `--help` for current flags, and run `gitleaks` over the full history
with redaction on. Anything declined or unavailable goes in the report as not
covered.

## 3. Find

Spawn one `review-security` agent per area, all in parallel. Each prompt carries
that area's file list, the **Checklist** section of
[`review-security`](../review-security/SKILL.md) pasted in full, **Audit
additions** and **Exclusions** below pasted in full, and this brief:

"Audit these files as they stand now. Apply every Checklist item to every file.
Find where untrusted input enters, then trace it to where it is used, reading
into other areas when the data flows there. For each candidate give the file and
line, category, severity, and exploit scenario. Under 600 words."

The step is done when every area has returned.

## 4. Disprove

For each candidate, spawn one `review-security` agent, all in parallel. Each
prompt carries that candidate, **Exclusions** pasted in full, and this brief:

"Try to disprove this finding by reading the code. Check whether the input is
really attacker-controlled, whether something upstream already neutralizes it,
and whether an Exclusion applies. Return a confidence from 1 to 10 that it is a
real, exploitable vulnerability, with your reasoning in under 100 words."

A score of 8 or higher is **confirmed**, 5 to 7 is **needs a human look**, and
below 5 is dropped.

## 5. Report

Write `.research/security/audit-<YYYY-MM-DD>.md` in the primary checkout, after
`git check-ignore` confirms that path is ignored. If it is not ignored, ask where
to put the report, because a vulnerability list should stay out of the
repository. Name leaked secrets by file, commit, and type, and leave out their
values.

Sections, in order:

1. **Scope**: the commit, the date, areas audited, and the skip list.
2. **Confirmed findings**, High first, each with severity, file and line, what
   is wrong, how it could be exploited, and the fix.
3. **Needs a human look**, in the same format.
4. **Dependencies** and **Secrets in git history**: scanner results.
5. **Not covered**: scanners that did not run, uncommitted changes, deployed
   configuration and infrastructure, behavior under live attack, and compliance.

Finish in chat with the report path, counts by severity, and the worst finding.
Offer to log the findings with `backlog-add`.

## Audit additions

Apply these along with the Checklist:

- **Denial of service**: unbounded input, recursion, or allocation that an
  outside attacker can trigger; missing limits on expensive public endpoints.

## Exclusions

Report a finding only with a concrete attack path from a named source of
untrusted input. These do not count:

- Missing hardening with no specific exploit.
- Dependency versions, which the scanners cover.
- Files used only by tests, unless they hold real credentials.
- Prose documentation for people. Agent instructions stay in scope.
- Theoretical race conditions.
- Attacks that require already controlling environment variables, CLI flags, or
  the local account running the program.
