# Local triage

Triage the authoritative Markdown record: a backlog item in
`.backlog/items/` (see `docs/agents/backlog.md`). Store its triage state in a
`triage:` frontmatter field, separate from `status`, using one of these
values:

- `needs-triage`: the issue has not been assessed.
- `needs-info`: more information is required before the issue can proceed.
- `ready-for-agent`: the issue is sufficiently specified for implementation.
- `ready-for-human`: the issue needs a maintainer decision or action.
- `wontfix`: the issue will not be pursued.

## Local record workflow

- Read the issue and its related local specification before changing status.
- Update the item's `triage` field when triage changes its state.
- Append questions, answers, investigation notes, and decisions to a
  `## Triage notes` section in the issue record.
- Store an implementation-ready brief in an `## Agent brief` section of the
  same issue record.
- Preserve prior notes so the reasoning behind status changes remains durable.

Do not create or manage hosted labels, comments, or tickets during local
triage. Publishing a local issue to a hosted tracker requires a separate,
explicit request.
