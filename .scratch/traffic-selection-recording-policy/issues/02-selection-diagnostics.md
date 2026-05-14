# Issue 02: Record Ruleset Selection Diagnostics

Status: completed
Blocked by: 01-selector-required.md

## Scope

- Add explicit ruleset candidate and winner diagnostics to match results.
- Preserve current single winner ruleset behavior.
- Carry diagnostics internally through runtime decision.
- Persist diagnostics into traffic `explain_json`.

## Acceptance

- Diagnostics list selector-matched ruleset candidates.
- Diagnostics identify the winner ruleset.
- Diagnostics include candidate rule IDs and winner rule ID when present.
- SDK response JSON does not expose internal diagnostics.

## Result

- Match explanation now includes selector-matched ruleset candidates and the selected winner.
- Runtime decisions carry internal diagnostics using `json:"-"`.
- Traffic `explain_json` persists `ruleset_selection` and `rule_selection`.
- Rule miss keeps the selected ruleset in trace and diagnostics without trying another ruleset.
