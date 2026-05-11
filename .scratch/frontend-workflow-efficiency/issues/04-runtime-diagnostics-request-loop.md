# Runtime diagnostics request loop

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-workflow-efficiency/PRD.md

## What to build

Turn Runtime Diagnostics into a clearer diagnosis loop. Runtime request rows should be inspectable, replay results should stay connected to the source request, and users should be able to jump from a request to the owning ruleset/rule context.

## Acceptance criteria

- [x] Runtime metrics and filters use shared primitives where appropriate.
- [x] A selected runtime request exposes a readable detail drawer or side panel.
- [x] Request detail includes outcome, status, duration, namespace, host/path, trace, ruleset, rule, fallback, and message when available.
- [x] Replay result clearly references the source request.
- [x] Open-rules navigation preserves ruleset and selected rule context.
- [x] Runtime Diagnostics browser QA shows a clear diagnosis flow.

## Completion notes

- Added request row selection, request detail drawer, shared metrics/filter primitives, and route/replay facts.
- Generated a local runtime request with trace `uiux-runtime-qa-1` to verify the request drawer and replay result source label `#1`.
- Browser QA covered the runtime dashboard, request detail drawer, and replay output.

## Blocked by

- .scratch/frontend-workflow-efficiency/issues/01-shared-ui-primitives.md
