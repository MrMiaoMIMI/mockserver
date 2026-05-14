# Issue 03: Selection Diagnostics UX

Status: completed
Blocked by: none

## Scope

- Preserve ruleset candidates and winner ruleset in summarized explain output.
- Add a structured Traffic detail diagnostics view for `ruleset_selection` and `rule_selection`.
- Add lightweight frontend helper tests for the diagnostics view model.

## Acceptance

- Replay/simulation summary output retains selector candidate context.
- Traffic detail exposes candidate rulesets, winner ruleset, candidate rule IDs, and winner rule ID without requiring raw JSON inspection.
- Frontend tests cover empty and populated diagnostics.

## Result

- `summarizeExplain` now keeps winner ruleset and ruleset candidates.
- Traffic detail renders a structured selection diagnostics section.
- Added frontend utility tests for populated and empty diagnostics.
