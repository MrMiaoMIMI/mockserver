# Issue 01: Backend Traffic Summary And Detail Contract

Status: done
Blocked by: none

## Scope

- Add a traffic event detail service/repository/DAO path.
- Add `GET /mockserver/api/v1/admin/traffic/events/:traffic_event_id`.
- Make list response use summary rows by default.
- Keep detail response full fidelity with event, decision, explain, and indexes.

## Acceptance

- List endpoint omits raw payloads and indexes by default.
- Detail endpoint returns one event with payloads and indexes.
- Missing detail id returns a business not found error.
- Go tests cover the new service/controller contract.

## Result

- Added `GET /mockserver/api/v1/admin/traffic/events/:traffic_event_id`.
- List response now returns summary rows only; detail response returns event, decision, explain, and indexes.
- API smoke confirmed list omits payloads/indexes and detail includes payloads/indexes.
