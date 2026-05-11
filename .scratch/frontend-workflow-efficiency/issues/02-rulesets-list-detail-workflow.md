# Rulesets list and detail workflow

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-workflow-efficiency/PRD.md

## What to build

Redesign the Rulesets entry page into a list-first workflow with a detail drawer. Users should be able to scan many rulesets, filter consistently, open a detail view, inspect selector and metadata details, and jump to rule management or secondary actions without visual clutter.

## Acceptance criteria

- [x] Rulesets are presented in a list-first layout that scans better than the previous card wall.
- [x] Core fields are visible in each row: identity, enabled state, publish state, namespace, protocol, version, rule count, selector summary, and primary action.
- [x] A detail drawer exposes secondary metadata, full selector summary, stats, settings, publish, and manage-rules actions.
- [x] Existing create, settings, publish, and manage-rules workflows remain accessible.
- [x] Filter and summary controls use shared primitives.
- [x] Desktop and narrow viewport layouts remain usable.

## Completion notes

- Rebuilt `RuleSetList.vue` around a dense inventory list, shared summary metrics, shared filters, and an inspectable detail drawer.
- Browser QA covered the desktop list, detail drawer, and 390px narrow viewport.

## Blocked by

- .scratch/frontend-workflow-efficiency/issues/01-shared-ui-primitives.md
