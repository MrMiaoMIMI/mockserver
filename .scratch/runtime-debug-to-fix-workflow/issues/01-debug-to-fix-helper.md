# Debug-to-Fix helper module

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/runtime-debug-to-fix-workflow/PRD.md

## What to build

Create the tested frontend helper behavior that turns a Runtime Diagnostics row into a temporary debug payload, chooses a target ruleset when possible, creates a readable diagnosis summary, and derives a safe rule seed from the captured runtime event.

## Acceptance criteria

- [x] Runtime rows can be converted into debug payloads with action, target ruleset, request facts, and captured event.
- [x] Target ruleset inference prefers an explicit runtime ruleset ID and falls back to same namespace/protocol drafts.
- [x] A generated rule seed includes useful method, host, path, and query predicates without overfitting to trace ID.
- [x] Diagnosis copy distinguishes matched, fallback, unmatched, and error outcomes.
- [x] Unit tests cover target inference, payload creation, diagnosis, and generated rule seed behavior.

## Completion notes

- Added `debugToFix` helper behavior and 4 unit tests.
- Verified with `npm test` and `npm run type-check`.

## Blocked by

None - can start immediately.
