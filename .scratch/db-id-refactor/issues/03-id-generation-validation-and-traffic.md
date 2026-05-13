# Issue 03: ID Generation, Validation, And Traffic Querying

Status: done

Blocked by: issues/02-backend-repository-id-resolution.md

## Scope

Shorten generated codes, validate manual codes, and make traffic storage/querying use numeric references where possible.

## Tasks

- Add centralized code length and character validation.
- Cap ruleset codes at 96 characters and namespace/rule codes at 64 characters.
- Generate compact snapshot codes.
- Resolve traffic namespace/ruleset/snapshot references to numeric IDs before insert and query.
- Keep short denormalized codes for frontend display and debugging.

## Acceptance

- Invalid or overlong external codes fail before database write.
- Traffic list filters by namespace/ruleset code resolve to numeric indexed fields.
- Traffic index rows are linked by `traffic_event_id`.
