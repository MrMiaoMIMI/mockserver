# Issue 02: Frontend Response Editor Model

Status: Completed

## Scope

Introduce protocol-driven response helpers and change rule/fallback form state to structured response payloads.

## Changes

- Added `web/src/utils/responseSpec.ts`.
- Added helpers for defaults, JSON drafts, validation, nested path access, and build serialization.
- Changed response form state from raw JSON strings to structured `responsePayload` plus `responseFieldDrafts`.
- Extended frontend protocol field types with response metadata such as `required`, `default`, `min`, `max`, and `format`.

## Verification

- `npm run test -- src/utils/__tests__/responseSpec.test.ts`
- `npm run test`
- `npm run type-check`
