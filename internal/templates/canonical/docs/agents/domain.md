# Domain documentation

Keep the repository's durable domain context in `context.md` at the repository
root. `qbs init` excludes this file through the repository-local Git common
exclude, but does not create it. Record significant architectural decisions in
`docs/adr/` using one Markdown file per decision.

These documents are created lazily by the domain-modeling workflow. Do not add
placeholder files just to satisfy this layout.
