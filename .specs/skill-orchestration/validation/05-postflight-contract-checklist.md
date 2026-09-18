# Ticket 05 validation checklist

This checklist is the focused validation artifact for the postflight contract.
It is intentionally behavior-oriented so it can be applied to `to-spec`,
`to-tickets`, and `implement-spec` reports.

- [ ] Every source requirement has one coverage status: `complete`, `partial`,
      `missing`, `ambiguous`, or `unrequested`, with a precise pointer.
- [ ] Verified evidence names an actual check, result, and inspection pointer.
- [ ] Inferred conclusions are listed separately and are not used as evidence.
- [ ] Assumptions, low-confidence areas, risks, and likely failure modes are
      explicit.
- [ ] Confidence-raising checks are minimal and ordered by expected value.
- [ ] Complexity findings name the preserved requirements and their checks;
      no required behavior is removed for simplification.
- [ ] If red-team review ran, its prompt contained only the requirements and
      artifact, its report is read-only, and every finding has a requirement
      pointer, failure scenario, evidence, severity, and corrective check.
- [ ] If red-team review did not run, `red_team_findings` is explicitly empty.

Repository validation for this ticket:

- `git diff --check`
- `go test ./...`
