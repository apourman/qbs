---
name: explain
description: Explain concepts in plain language using a large, clear diagram first.
---

# Explain

Use this skill when the user wants something explained, clarified, visualized,
or broken down for a non-technical reader.

## Instructions

1. Start with one large, clear ASCII or Mermaid diagram that shows the main
   parts and how they connect.
2. Use generous spacing, simple labels, and arrows that show movement or cause.
3. Explain technical words with everyday comparisons.
4. Follow the diagram with a short explanation in plain language.
5. Keep the answer focused on the user's question. Do not add technical detail
   that the diagram and short explanation do not need.

Use ASCII by default because it works everywhere. Use Mermaid only when the
destination is known to render Mermaid. Consult `EXAMPLES.md` for diagram
style, and `references/mermaid.md` only when Mermaid is appropriate.

## Quality check

Before answering, make sure the diagram teaches the relationship by itself,
the labels are understandable to a layman, and the prose stays brief.
