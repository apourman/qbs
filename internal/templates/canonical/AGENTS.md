<!-- qbs-managed-instruction: v2 -->

# QBS workspace

Read `CLAUDE.md` for shared project instructions. Reusable skills are installed
globally; project-only skills may live in the harness-native local skill
directory.

Use the canonical domain/feature layout for project knowledge:

- Research: `.research/<domain-or-feature>/<research-slug>.md`
- Specs: `.specs/<domain-or-feature>/spec.md` (exactly one per directory)
- Tickets for a spec: `.specs/<domain-or-feature>/tickets/<NN>-<ticket-slug>.md`
- Future work for a spec: a companion backlog file beside `.specs/<domain-or-feature>/spec.md`

Keep deferred or possible future work in the spec's companion `.backlog.md`
file. Use ordinary unchecked Markdown tasks, and keep the backlog beside its
source spec. A backlog records context; it is not a queue of implementation
tickets or an indication that every item is ready for an agent. When an item
becomes actionable, create a separate numbered issue and link it from the
backlog entry. Preserve any link to the ticket or decision that originally
raised the idea.

Choose a stable, lowercase kebab-case domain or feature name and reuse it for
the related research, spec, and tickets. New specs and tickets belong in these
domain directories, not directly at the root of `.specs/`.

## Agent skills

- Issue tracker: see `docs/agents/issue-tracker.md` for the repository's tracker workflow.
- Domain documentation: see `docs/agents/domain.md` for `CONTEXT.md` and `docs/adr/`.
- Local triage: when triage is installed, see `docs/agents/triage.md` for the status workflow.
