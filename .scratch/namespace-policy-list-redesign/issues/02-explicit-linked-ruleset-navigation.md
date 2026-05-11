# Explicit linked ruleset navigation

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/namespace-policy-list-redesign/PRD.md

## What to build

Remove the ambiguous namespace-level `View rules` action and make every ruleset navigation explicit. Users should only navigate to a rules page by choosing a specific linked ruleset.

## Acceptance criteria

- [x] The Namespaces page no longer renders a `View rules` button.
- [x] Linked rulesets are shown as explicit clickable targets.
- [x] A namespace with multiple linked rulesets does not default to opening the first one.
- [x] Overflow or hidden linked rulesets remain discoverable through a compact interaction.
- [x] Empty linked-ruleset state is clear for unused namespaces.

## Completion notes

- Removed the namespace-level `View rules` button.
- Kept only per-ruleset navigation controls for ruleset pages.
- Added compact overflow handling for linked rulesets beyond the visible set.
- Added view-model test coverage ensuring all linked rulesets remain available.

## Blocked by

- .scratch/namespace-policy-list-redesign/issues/01-namespace-policy-list-surface.md
