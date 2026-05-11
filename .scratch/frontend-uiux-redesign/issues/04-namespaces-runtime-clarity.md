# Namespaces and Runtime clarity

Status: done
Type: AFK

## Parent

.scratch/frontend-uiux-redesign/PRD.md

## What to build

Apply the new UI/UX direction to Namespaces and Runtime Diagnostics. Namespaces should make fallback policy easy to understand at a glance while keeping detailed policy editing available. Runtime Diagnostics should emphasize traffic health first and present request/fallback/ruleset details in a calm, readable diagnostic layout.

## Acceptance criteria

- [x] Namespaces filters, summary cards, namespace entries, fallback summaries, and policy editor dialog align with the redesigned visual system.
- [x] Namespace cards clearly distinguish usage state from ruleset-miss and rule-miss fallback policy.
- [x] Runtime Diagnostics metric band, request stream, side panels, replay result, and empty states use clear hierarchy and lighter surfaces.
- [x] Runtime outcome/status colors use semantic state colors rather than generic primary color.
- [x] Existing actions such as edit policy, view rules, open rules, replay, and refresh remain accessible.
- [x] Desktop screenshots show clear boundaries and reduced visual noise.

## Blocked by

- .scratch/frontend-uiux-redesign/issues/01-theme-shell-clarity-baseline.md

## Comments

- Completed with the new shared tokens and page-specific hover/selected-state cleanup. Verified through Namespaces and Runtime Diagnostics browser screenshots.
