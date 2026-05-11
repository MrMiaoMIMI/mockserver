# Runtime debug queue simplification

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-intent-based-console-simplification/PRD.md

## What to build

Make Runtime Diagnostics read as a debug queue first. Hide traffic metric cards behind Insights and remove the duplicate Details row button, while keeping row click and request drawer actions intact.

## Acceptance criteria

- [x] Runtime page first screen starts with the request queue rather than traffic metric cards.
- [x] Traffic analytics remain available through Insights.
- [x] Request rows rely on row click and keyboard selection for details instead of an extra Details button.
- [x] Request rows stay compact while preserving method, path, status, outcome, fallback, namespace, host, trace, and timing.
- [x] Replay, open-rules, use-as-simulation, and create-rule remain available in the request drawer.

## Blocked by

- .scratch/frontend-intent-based-console-simplification/issues/01-global-and-rulesets-noise-reduction.md

## Completion notes

Implemented in `Dashboard.vue`. Runtime Diagnostics now reads as a request queue first. Traffic metrics moved behind Insights, and request detail access is row-driven with advanced actions retained in the drawer.
