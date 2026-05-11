# Workspace debug intake

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/runtime-debug-to-fix-workflow/PRD.md

## What to build

Teach Ruleset Workspace to consume a temporary debug payload from Runtime Diagnostics, show the source request context, prefill simulation with the captured event, and open the correct workbench task based on the requested action.

## Acceptance criteria

- [x] Workspace loads a debug payload from route query/session storage after the target ruleset loads.
- [x] Use-as-simulation payloads open the simulation task and populate Event JSON with the runtime event.
- [x] Create-rule payloads open the editor task in create mode with a request-derived seed rule.
- [x] The workspace shows a source banner with request ID, trace, method, host/path, outcome, and source action.
- [x] The loaded debug payload can be cleared without breaking normal workspace behavior.

## Completion notes

- Workspace debug intake supports both simulation and seeded create-rule flows.
- Browser QA verified the runtime source banner, prefilled Event JSON, and seeded rule form.

## Blocked by

- .scratch/runtime-debug-to-fix-workflow/issues/01-debug-to-fix-helper.md
- .scratch/runtime-debug-to-fix-workflow/issues/02-runtime-debug-actions.md
