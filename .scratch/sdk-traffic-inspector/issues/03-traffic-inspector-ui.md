# 03. Traffic Inspector UI

Status: completed

Blocked by: .scratch/sdk-traffic-inspector/issues/02-traffic-recorder-and-query-api.md

## Scope

Replace the Dashboard/Runtime Diagnostics experience with a Traffic Inspector backed by persisted SDK decision traffic.

## Requirements

- Rename the user-facing page to Traffic Inspector.
- Query the new traffic APIs instead of runtime metrics.
- Show traffic rows, protocol-aware filters, summary cards, and detail drawer.
- Show event JSON, decision JSON, explanation JSON, and indexed fields.
- Preserve useful actions: open matched rules, copy trace, and replay event against published simulation.

## Verification

- Frontend unit tests cover traffic row view models and filters.
- `npm run build` passes.

## Completion Notes

- Replaced `/dashboard` user experience with Traffic Inspector backed by persisted SDK traffic.
- Added protocol-aware field filters, summary panels, event detail drawer, replay, open rules, and copy trace actions.
- Verified with frontend tests and production build.
