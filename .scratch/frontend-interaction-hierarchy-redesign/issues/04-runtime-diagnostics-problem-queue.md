# Runtime diagnostics problem queue

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-interaction-hierarchy-redesign/PRD.md

## What to build

Refine Runtime Diagnostics into a problem queue that prioritizes recent requests and keeps analytics secondary. Request rows should be easier to scan, and advanced diagnosis actions should live in the request detail drawer.

## Acceptance criteria

- [x] Runtime first screen prioritizes recent request rows over secondary analytics panels.
- [x] Request rows clearly show method, path, outcome, status, fallback reason, and detail action.
- [x] Replay, open-rules, use-as-simulation, and create-rule actions remain available from request detail.
- [x] Empty or low-value analytics panels do not dominate the page.
- [x] Runtime filters and sort controls remain available without excessive vertical space.
- [x] Existing replay and runtime-to-workspace debug flows still work.

## Completion notes

- Made Insights optional and hidden by default.
- Kept request rows focused on problem facts and a Details action.
- Preserved replay, open rules, use-as-simulation, and create-rule actions inside the request detail drawer.

## Blocked by

- .scratch/frontend-interaction-hierarchy-redesign/issues/01-rulesets-entry-information-hierarchy.md
