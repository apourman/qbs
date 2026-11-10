---
name: review-security
description: Review the changes since a fixed point for exploitable security vulnerabilities, including unsafe instructions in agent skills and rules. Runs the read-only `review-security` agent to find issues, then again to disprove each one. Use when the user wants a security review of a branch, a PR, or work-in-progress changes.
---

Find vulnerabilities a change introduces, and report only the ones that survive
an attempt to disprove them. The caller supplies the fixed point (a commit,
branch, or tag); ask for one if it is missing.

## 1. Pin the diff

```bash
git rev-parse --verify <fixed-point>
git log <fixed-point>..HEAD --oneline
mktemp
git diff <fixed-point>...HEAD > <file mktemp printed>
```

Stop and report a bad ref or an empty diff. The file lives outside the
repository because the review agents have no shell and read the diff from it.

## 2. Find

Spawn one `review-security` agent. Its prompt carries the diff file path, the
commit list, **Checklist** and **Exclusions** pasted in full, and this brief:

"Find vulnerabilities this diff adds. Apply every Checklist item to every
changed file. Read the surrounding code and the project's existing validation
and sanitizing patterns, then trace each suspect value from where it enters to
where it is used. For each candidate give the file and line, Checklist category,
severity, and exploit scenario. Under 600 words."

The step is done when the agent has returned its candidates. No candidates: go
to Report.

## 3. Disprove

For each candidate, spawn one `review-security` agent, all in parallel. Each
prompt carries the diff file path, that candidate, **Exclusions** pasted in
full, and this brief:

"Try to disprove this finding by reading the code. Check whether the input is
really attacker-controlled, whether something upstream already neutralizes it,
and whether an Exclusion applies. Return a confidence from 1 to 10 that it is a
real, exploitable vulnerability, with your reasoning in under 100 words."

Keep candidates scoring 8 or higher. The step is done when every candidate has
a score.

## 4. Report

Under a `## Security` heading, High findings first, one block per finding:

```markdown
### <category>: `path/to/file:line`
- Severity: High or Medium
- What is wrong:
- How it could be exploited:
- Fix:
```

End with one line: the range reviewed, candidates found, candidates confirmed.
With none confirmed, say no exploitable vulnerabilities were found in that
range.

## Checklist

- **Injection**: SQL, NoSQL, command, template, and XXE injection; path
  traversal.
- **Access control**: authentication bypass, privilege escalation, session and
  JWT flaws, missing authorization on a sensitive action.
- **Crypto and secrets**: hardcoded keys, passwords, or tokens; weak algorithms;
  bad randomness; skipped certificate checks.
- **Code execution**: unsafe deserialization, `eval` and similar, XSS.
- **Data exposure**: secrets or personal data in logs, API leaks, debug output.
- **Memory safety**: in C, C++, and `unsafe` blocks, buffer overflows,
  use-after-free, and integer overflows that reach a memory operation.
- **Files and processes**: writes or deletes that escape their target directory
  or follow an attacker-placed symlink, overwriting user files without a check,
  temporary files others can read, running a downloaded file without verifying
  it.
- **Agent instructions**: in skills, `AGENTS.md`, `CLAUDE.md`, rules, agent
  definitions, hooks, and prompt templates, anything that tells an agent to skip
  an approval or safety gate, widen its permissions, run untrusted content, read
  or send secrets, push, publish, or delete without asking, or obey instructions
  found in fetched pages, issues, or file contents.

## Exclusions

Report a finding only with a concrete attack path from a named source of
untrusted input. These do not count:

- Denial of service, rate limits, and resource exhaustion.
- Missing hardening with no specific exploit.
- Outdated dependencies.
- Files used only by tests.
- Prose documentation for people. Agent instructions stay in scope.
- Theoretical race conditions.
- Attacks that require already controlling environment variables, CLI flags, or
  the local account running the program.
- Problems that were already there and the diff does not make worse.
