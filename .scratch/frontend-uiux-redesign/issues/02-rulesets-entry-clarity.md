# Rulesets entry clarity

Status: done
Type: AFK

## Parent

.scratch/frontend-uiux-redesign/PRD.md

## What to build

Redesign the Rulesets entry page so users can scan and act on rulesets quickly. The page should prioritize search, core filters, summary metrics, ruleset identity, namespace, selectors, rule count, publish state, and primary navigation to rule management. Secondary actions should remain available without visually competing with the main workflow.

## Acceptance criteria

- [x] Rulesets search, namespace filter, sort, and status filters are easy to scan and have clear selected states.
- [x] Summary cards are useful but visually secondary to the list content.
- [x] Ruleset cards or rows have stronger boundaries, clearer hierarchy, and better spacing than the previous dark-card layout.
- [x] Primary action to manage rules is obvious; settings and publish remain available without overwhelming the card.
- [x] Selector values remain compact and inspectable.
- [x] Desktop and narrow viewport layouts do not overlap or hide core actions.

## Blocked by

- .scratch/frontend-uiux-redesign/issues/01-theme-shell-clarity-baseline.md

## Comments

- Completed with the light console visual system and Rulesets browser QA. Captured desktop and 390px screenshots under `output/playwright/`.
