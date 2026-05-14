# Traffic Selection Recording Policy PRD

Status: completed
Owner: Codex
Created: 2026-05-14

## Background

SDK traffic can include requests that were not intended to be handled by MockServer. If a ruleset has an empty selector, it behaves like a protocol + namespace catch-all and can turn unrelated traffic into rule misses or even unintended mock responses. Even when no ruleset selector matches, recording every raw ruleset miss pollutes the traffic log and grows storage quickly.

Compatibility is not required. The new behavior should make ruleset ownership explicit and keep the raw traffic table focused on useful diagnostic records.

## Goals

- Require every ruleset to declare at least one selector condition.
- Keep the current single winner ruleset model; do not fall through to another ruleset after the selected ruleset has no matching rule.
- Do not persist raw traffic for `ruleset_miss`; record it only through lightweight metrics.
- Persist `rule_miss`, `matched`, and decision errors.
- Persist matching process diagnostics for recorded traffic, including selector-matched ruleset candidates, the selected winner ruleset, candidate rules, and the winner rule when present.
- Keep SDK response JSON unchanged; matching diagnostics are internal to MockServer traffic recording.

## Non-Goals

- Do not implement namespace-level exclusion rules in this slice.
- Do not add a new DB aggregate table in this slice.
- Do not change rule matching to cross-ruleset fallthrough.
- Do not expose internal diagnostics to `mocksdk` response payloads.

## Functional Requirements

- Ruleset validation fails when `selector.all` is empty.
- A published or simulated ruleset with an empty selector cannot compile.
- The engine records selector-matched ruleset candidates before it selects the single winner.
- The engine records the winner ruleset even when no rule inside that ruleset matches.
- `RuntimeDecision` carries internal diagnostics with `json:"-"` so SDK clients do not receive internal selection data.
- Traffic `explain_json` includes:
  - `trace`
  - `ruleset_selection.winner_ruleset_id`
  - `ruleset_selection.candidates[]`
  - `rule_selection.candidate_rule_ids[]`
  - `rule_selection.winner_rule_id`
- `ruleset_miss` SDK decisions do not create rows in `mockserver_traffic_event_tab`.
- `ruleset_miss` SDK decisions increment runtime metrics with fallback reason and event summary.

## Acceptance Criteria

- `engine.ValidateRuleSet` rejects empty selectors.
- Existing matching still uses a single selected ruleset and does not try the next candidate after rule miss.
- Recorded `rule_miss` traffic includes the winner ruleset in trace/diagnostics.
- Recorded matched traffic includes ruleset candidates and winner rule diagnostics.
- `ruleset_miss` SDK traffic is absent from raw traffic rows but visible in metrics counters.
- Go tests and frontend build/tests pass.

## Execution Result

- Implemented on 2026-05-14.
- Backend selector validation now rejects empty `selector.all`.
- Runtime matching records selector-matched ruleset candidates, winner ruleset, candidate rules, and winner rule diagnostics.
- SDK `ruleset_miss` traffic skips raw persistence and is counted in runtime metrics.
- Frontend ruleset settings now blocks saving without selector conditions.
- Documentation updated in `README.md` and `docs/mocksdk.md`.
