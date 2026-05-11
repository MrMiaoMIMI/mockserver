# Runtime debug actions

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/runtime-debug-to-fix-workflow/PRD.md

## What to build

Extend Runtime Diagnostics request detail so a selected runtime request explains the likely diagnosis and offers direct actions to use the request as a simulation event or create a rule from the request when a target ruleset can be inferred.

## Acceptance criteria

- [x] Request detail shows a diagnosis summary and suggested next action.
- [x] Request detail exposes a Use as simulation action when the request has a replayable event and a target ruleset.
- [x] Request detail exposes a Create rule action when the request has a replayable event and a target ruleset.
- [x] Debug actions store a temporary payload and navigate to the correct Ruleset Workspace route.
- [x] Existing replay, copy trace, and open-rules actions continue to work.

## Completion notes

- Runtime request detail now shows diagnosis and target context plus debug actions.
- Browser QA covered the request drawer with trace `debug-to-fix-qa-1`.

## Blocked by

- .scratch/runtime-debug-to-fix-workflow/issues/01-debug-to-fix-helper.md
