# Traffic Observability Hardening PRD

Status: completed
Owner: Codex
Created: 2026-05-14

## Background

The current traffic recording policy now requires ruleset selectors, skips raw SDK traffic for `ruleset_miss`, and persists rule-selection diagnostics for recorded traffic. A follow-up review found several remaining gaps: suppressed `ruleset_miss` metrics are not protocol-aware, traffic stats are sampled, indexed-field filtering can generate large ID sets, selection diagnostics are not preserved in summary explain output, and the Traffic detail page still shows the new diagnostics only as raw JSON.

Compatibility is not required. This slice should optimize the current design for the intended user workflow: operators need to understand unexpected SDK traffic without storing raw `ruleset_miss`, and they need reliable, scalable traffic queries for persisted events.

## Goals

- Make suppressed SDK `ruleset_miss` visible through protocol/source/operation-aware runtime metrics.
- Keep raw `ruleset_miss` traffic suppressed.
- Make traffic list stats exact for the current filter instead of sampling recent rows.
- Reduce indexed-field filtering risk by bounding ID materialization and by using indexed query constraints first.
- Preserve selection diagnostics in summarized explain output.
- Show ruleset/rule selection diagnostics as structured information in Traffic detail.
- Tighten raw-skip guard and add regression coverage for edge cases.

## Non-Goals

- Do not add a new persistent aggregate table in this slice.
- Do not change DB schema.
- Do not reintroduce raw `ruleset_miss` persistence.
- Do not change the single winner ruleset model.

## Functional Requirements

- Runtime metrics records include `source`, `protocol`, and `operation`.
- Runtime metrics expose fallback breakdown by protocol, namespace, and operation.
- Suppressed SDK `ruleset_miss` records use source `sdk_decision`, protocol from event, operation from event operation/request operation/request method, and protocol-specific target path/cmd/key.
- `RecordSDKDecision` skips raw persistence only when the decision is a successful fallback with `fallback_reason=ruleset_miss`.
- Traffic stats count all rows matching the current filters.
- Indexed-field filters use query constraints before ID materialization and cap the number of materialized IDs to protect the event query.
- `summarizeExplain` keeps `winner_ruleset_id` and `ruleset_candidates`.
- Traffic detail shows `ruleset_selection` and `rule_selection` in a structured diagnostics section.

## Acceptance Criteria

- Go unit tests cover protocol-aware suppressed ruleset-miss metrics.
- Go unit tests cover exact traffic stats and bounded indexed-filter behavior.
- Go unit tests cover `ruleset_miss` response fallback raw suppression and diagnostics-hidden SDK response JSON.
- Go unit tests cover explain summary preserving selection diagnostics.
- Frontend tests cover extraction/render model for traffic selection diagnostics.
- `go test ./...`, `npm test`, `npm run build`, and `git diff --check` pass.

## Execution Result

- Implemented on 2026-05-14.
- Runtime metrics now include source/protocol/operation dimensions and fallback stats by protocol, namespace, and operation.
- Suppressed SDK `ruleset_miss` records are counted in metrics with protocol-aware summaries while raw traffic remains suppressed.
- Traffic stats are computed with exact grouped DB queries instead of a recent-row sample.
- Indexed-field filtering now caps materialized event IDs and returns a clear error when a filter is too broad.
- Traffic ruleset/namespace filters preserve both DB ID and business code to avoid missing historical rows after delete/recreate.
- Explain summary keeps ruleset winner/candidate diagnostics.
- Traffic detail now shows structured selection diagnostics in addition to raw JSON.
