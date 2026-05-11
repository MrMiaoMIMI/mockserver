# Authoring View Model And Readiness

Status: done
Type: AFK

## Parent

.scratch/rule-authoring-ux-redesign/PRD.md

## What to build

Create a testable authoring view model that turns the current rule form into section readiness, concise summaries, validation hints, action descriptions, and generated simulation events. The Vue editor should use this helper rather than embedding all interpretation inside the component.

## Acceptance criteria

- [x] The view model reports readiness for identity, condition, action, and advanced sections.
- [x] The view model produces a concise authoring overview for the current rule.
- [x] The view model explains the selected action type in user-facing terms.
- [x] The view model produces simulation event suggestions from the current rule and ruleset selector.
- [x] The generated event can be edited by the user before simulation.
- [x] Unit tests cover readiness, summaries, action descriptions, and simulation event generation.

## Blocked by

- .scratch/rule-authoring-ux-redesign/issues/01-condition-preset-builder.md

## Verification

- Added `web/src/utils/ruleAuthoring.ts` as the testable authoring view model for readiness, summaries, action descriptions, simulation event suggestions, and draft override construction.
- Covered readiness, event generation, and draft override behavior in `web/src/utils/__tests__/ruleAuthoring.test.ts`.
