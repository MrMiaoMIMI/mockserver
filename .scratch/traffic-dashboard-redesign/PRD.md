# Traffic Dashboard Redesign PRD

Status: done
Owner: Codex
Created: 2026-05-13

## Background

The current `/dashboard` page is used as the SDK decision traffic console. It mixes live traffic listing, aggregate cards, indexed-field filtering, event detail, and replay output in a two-column dashboard. This works for small traffic volumes, but it is not ergonomic for fast-growing traffic data:

- The page fetches only the first 100 rows and has no user-visible pagination.
- Search filters only the loaded rows, so users can miss records outside the current page.
- The right-side `outcomes`, `protocols`, `namespaces`, and `replay result` panels can be clipped on shorter viewports and compete with the primary traffic table.
- The list request includes detail-heavy payloads and indexes even though they are only needed after a user opens one event.
- The traffic table needs indexes optimized for the main time-ordered lookup path.

Compatibility is not required. The new design should optimize the user-facing workflow and long-term data growth.

## Goals

- Make the page a traffic log first: dense, paginated, filterable, and easy to scan.
- Keep backend list responses lightweight by default.
- Provide a dedicated event detail path that loads full event, decision, explain, and indexed fields on demand.
- Move replay output into the selected event detail drawer.
- Avoid nested page/side-panel scrolling traps.
- Keep MySQL table/index design compliant with the MySQL Database Design Guide.

## Non-Goals

- Do not keep the existing two-column dashboard layout.
- Do not preserve the old list payload shape if it conflicts with performance.
- Do not implement retention management or async export in this slice.
- Do not add authentication or RBAC changes.

## User Experience

- Users land on a `Traffic Log` page with summary metrics, filters, a full-width table, and pagination.
- The table shows summary columns only: time, protocol, namespace, operation, outcome, decision, rule, trace, latency.
- Users can filter by time range, protocol, namespace, outcome, indexed field exact match, and search text.
- Pagination exposes page size and page number using backend `limit` and `offset`.
- Selecting a row opens a drawer with event facts, indexed fields, raw Event/Decision/Explain tabs, and Replay.
- Replay result is shown inside the drawer for the selected event.

## Backend Contract

- `GET /mockserver/api/v1/admin/traffic/events`
  - Returns paginated summary rows by default.
  - Supports `limit`, `offset`, time range, protocol, namespace, outcome, decision kind, ruleset, rule, fallback reason, and indexed-field exact filters.
  - Does not include raw event/decision/explain payloads or indexes by default.
- `GET /mockserver/api/v1/admin/traffic/events/{id}`
  - Returns one full event with raw event, decision, explain, and indexes.
  - Uses the numeric `id` primary key because this is an internal admin inspection endpoint.

## Database Requirements

- Keep `mockserver_traffic_event_tab` with `id BIGINT UNSIGNED AUTO_INCREMENT`.
- Add an index for the main timeline path: `idx_traffic_source_event_time_id (traffic_source, event_time, id)`.
- Preserve existing protocol, namespace, outcome, ruleset, fallback, trace, expire, and index-table lookup indexes.
- Do not add foreign keys, triggers, stored procedures, nullable fields, or timestamp/datetime fields.

## Acceptance Criteria

- Traffic list request does not include raw payloads or indexes unless the detail endpoint is called.
- The frontend has visible pagination and does not pretend the first page is the full dataset.
- The right-side dashboard rail is removed; short viewports do not clip critical controls.
- Detail drawer loads full event data on demand and can replay the selected event.
- Go tests, frontend tests, frontend build, and `git diff --check` pass.

## Verification

- `go test ./...`
- `npm test`
- `npm run build` (passes with existing Vite large chunk warning)
- API smoke against `MOCKSERVER_ADDR=:18080 go run ./cmd/server`
- Playwright screenshots:
  - `/tmp/mockserver-traffic-log-1440x900.png`
  - `/tmp/mockserver-traffic-log-390x844.png`
- `git diff --check`
