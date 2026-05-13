# SDK Traffic Inspector PRD

Status: completed

## Background

MockServer currently exposes a Runtime Diagnostics page backed by in-memory HTTP runtime metrics. That model is useful for local HTTP debugging, but it does not represent the primary user path after `mocksdk` was extracted: injected business code calls `/mockserver/api/v1/sdk/decision` and expects MockServer to return a decision.

Users need a persistent, multi-protocol Traffic Inspector for real SDK decision traffic. Manual simulate traffic from the ruleset/rules page already renders its result inline and must not be persisted as traffic.

## Goals

- Persist only real SDK decision traffic from `/mockserver/api/v1/sdk/decision`.
- Do not persist ruleset/rules page simulate requests.
- Support HTTP, cache, and future protocols through a protocol-neutral event table plus protocol-specific search index table.
- Keep MySQL schema compliant with the local `mysql-database-design-guide`.
- Let users query traffic by time, trace, protocol, namespace, outcome, decision kind, ruleset, rule, fallback reason, and selected protocol fields.
- Let users inspect the original event, decision, optional match explanation, and indexed fields.
- Replace the user-facing Runtime Diagnostics page with Traffic Inspector language and behavior.

## Non-Goals

- No compatibility with the previous in-memory Runtime Diagnostics UI is required.
- No persistence for admin simulate, draft simulate, or published simulate.
- No protocol-specific traffic tables.
- No MySQL foreign keys, triggers, stored routines, or database-side hash calculation.

## Data Model

### `mockserver_traffic_event_tab`

Stores protocol-neutral SDK decision facts and full JSON payloads. HTTP status is not a first-class column because it is protocol-specific; it lives in `decision_json` and may be indexed in `mockserver_traffic_event_index_tab` as `decision.response.status`.

### `mockserver_traffic_event_index_tab`

Stores selected searchable fields extracted from the event and decision. It stores the full value in `field_value_text`, a display-only preview in `field_value_preview`, and a `BIGINT UNSIGNED` application-computed hash in `field_value_hash` for compact indexed equality lookup.

## User Experience

The Traffic Inspector page shows:

- Recent SDK traffic rows with source, protocol, namespace, operation, outcome, decision kind, ruleset/rule, fallback reason, duration, and trace.
- Filter controls for time range, trace, protocol, namespace, outcome, decision kind, ruleset, rule, fallback reason, and protocol field filters.
- Summary cards for total traffic, match rate, fallback count, errors, and average duration.
- Detail drawer with event JSON, decision JSON, explanation JSON, and extracted index fields.
- Actions to open the matched ruleset and copy trace/event details.

## Acceptance Criteria

- SDK decision endpoint records both response and forward/fallback decisions.
- Manual simulate endpoints do not create traffic records.
- HTTP status is stored inside `decision_json` and indexed only as a protocol field.
- Long indexed values cannot fail insertion because the full value is stored in `TEXT`.
- `field_value_hash` is `BIGINT UNSIGNED` and computed by application code.
- Frontend no longer presents the module as Runtime Diagnostics; it presents SDK Traffic Inspector.
- Go tests, frontend tests/build, and schema checks pass.
