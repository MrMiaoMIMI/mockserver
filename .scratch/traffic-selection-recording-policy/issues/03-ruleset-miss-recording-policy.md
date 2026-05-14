# Issue 03: Skip Raw Traffic For Ruleset Miss

Status: completed
Blocked by: 02-selection-diagnostics.md

## Scope

- Do not persist raw SDK traffic when decision fallback reason is `ruleset_miss`.
- Keep `rule_miss`, `matched`, and error traffic persisted.
- Record ruleset miss through in-memory runtime metrics.

## Acceptance

- `ruleset_miss` calls do not call the traffic repository create path.
- Metrics record fallback reason `ruleset_miss`.
- `rule_miss` still persists raw traffic and includes winner ruleset diagnostics.

## Result

- `TrafficService.RecordSDKDecision` skips repository persistence for successful `ruleset_miss`.
- Runtime controller observes suppressed SDK `ruleset_miss` traffic in in-memory metrics.
- `rule_miss` persistence remains enabled and covered by service tests.
