# 02. Traffic Recorder And Query API

Status: completed

Blocked by: .scratch/sdk-traffic-inspector/issues/01-mysql-traffic-event-schema.md

## Scope

Record `/mockserver/api/v1/sdk/decision` traffic and expose admin APIs for list, stats, and detail retrieval.

## Requirements

- Record SDK decisions after `RuntimeView.DecidePublished` returns or errors.
- Do not record admin simulate traffic.
- Persist event JSON, decision JSON, optional explain JSON, extracted index fields, duration, outcome, decision kind, ruleset/rule, snapshot, fallback reason, and errors.
- Provide query filters for time range, trace, protocol, namespace, outcome, decision kind, ruleset, rule, fallback reason, and indexed field equality.
- Extract protocol indexes for HTTP method, host, path, query keys, headers, decision response status; cache operation/key/value when present; generic operation fields for future protocols.

## Verification

- Controller tests prove SDK traffic is recorded and simulate traffic is not.
- Repository/service tests cover hash extraction and query filtering.

## Completion Notes

- SDK decision endpoint records traffic after decision execution.
- Admin simulate endpoints remain outside the recorder path.
- Added `GET /mockserver/api/v1/admin/traffic/events` with time, protocol, namespace, outcome, ruleset/rule, trace, and indexed field filters.
- Added service/controller tests for SDK recorder behavior and index extraction.
