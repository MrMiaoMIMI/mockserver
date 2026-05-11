# Namespace policy list surface

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/namespace-policy-list-redesign/PRD.md

## What to build

Replace the namespace card grid with a compact policy-list surface that shows namespace identity, usage, fallback policies, linked rulesets, and row actions in a scan-friendly layout. The existing filters, sorting, create flow, and edit policy flow should remain available.

## Acceptance criteria

- [x] Namespaces render as a list rather than a card grid.
- [x] Each row shows namespace name, namespace ID, usage count, ruleset-miss policy, rule-miss policy, linked rulesets, and row actions.
- [x] Page-level summary information is visually secondary and does not duplicate large card blocks.
- [x] Existing search, usage filter, fallback filters, sorting, create, and edit flows still work.
- [x] The list remains readable on the desktop viewport used by MockServer.

## Completion notes

- Replaced the namespace grid/card surface with a policy-list surface.
- Kept toolbar filters, sorting, create namespace, and edit policy behavior in place.
- Converted the summary strip from large cards into compact metric pills.

## Blocked by

None - can start immediately.
