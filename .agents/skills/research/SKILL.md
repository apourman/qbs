---
name: research
description: Investigate a question against high-trust primary sources and capture the findings as a Markdown file in the repo. Use when the user wants a topic researched, docs or API facts gathered, or reading legwork delegated to a background agent.
---

Spin up the named **`researcher`** agent in the background to do the research, so you keep working while it reads.

Its job:

1. Investigate the question against **primary sources** (official docs, source code, specs, first-party APIs), not a secondary write-up of them. Follow every claim back to the source that owns it.
2. Write the findings to a single Markdown file, citing each claim's source.
3. Save it under `.research/<domain-or-feature>/<research-slug>.md`. Choose a stable, lowercase kebab-case domain or feature directory that identifies the subject of the research. Create the directory when needed. Keep one research question and its findings in one file, and say which path you wrote.

## Research paths

`.research/` is the canonical home for research notes. The first path component
after `.research/` is the domain or feature the research informs; it is not a
date, ticket number, or agent name. Reuse an existing domain or feature
directory when one applies. A research note may be referenced by a spec, but it
does not replace the spec.
