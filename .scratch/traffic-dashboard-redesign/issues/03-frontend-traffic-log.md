# Issue 03: Frontend Traffic Log Redesign

Status: done
Blocked by: 01-backend-summary-detail-contract.md

## Scope

- Replace the two-column dashboard layout with a full-width traffic log.
- Add backend-backed pagination.
- Load detail payloads only when a row is selected.
- Move replay result into the event detail drawer.
- Keep filters compact and scan-friendly.

## Acceptance

- Users can change page and page size.
- Search and filters reset to the first page.
- Detail drawer loads full event data and shows raw Event/Decision/Explain.
- Replay result is shown inside the drawer.
- Short desktop viewport does not clip the old right rail because the rail is removed.

## Result

- Replaced the dashboard with a full-width paginated traffic log.
- Added backend-backed filters for time range, protocol, namespace, outcome, trace/event lookup, and indexed-field exact match.
- Added on-demand detail drawer and moved replay output into the selected event drawer.
- Removed the old right-side outcomes/protocols/namespaces/replay rail.
