# Contextual rule editor workbench

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Replace the large disconnected rule modal with a contextual rule editor inside the Ruleset Workspace. The selected rule should stay connected to the rule stack while identity, condition, action, and advanced JSON editing are organized into clear editing sections.

This slice should extract a rule form adapter that maps API rule payloads to editable form state and back to valid API payloads. The adapter should make rule editing testable and reduce coupling between UI layout and rule serialization.

## Acceptance criteria

- [x] Selecting create or edit opens a contextual editor in the workspace rather than a disconnected large modal.
- [x] Rule identity, condition, action, and advanced/raw editing are separated into clear sections or modes.
- [x] Condition editing supports predicate, ALL, ANY, NOT, CEL expression, and raw JSON paths.
- [x] Action editing shows relevant fields for static response, template response, CEL response, sequence response, and webhook response.
- [x] Sequence response editing supports ordered steps without requiring users to hand-edit the full JSON array for common changes.
- [x] Webhook editing clearly exposes URL, method, timeout, and headers.
- [x] Form validation errors point to the relevant field or section before submitting to the backend.
- [x] Existing add-rule and update-rule API behavior is preserved.
- [x] Frontend build/type checks pass.

## Blocked by

- .scratch/ruleset-workspace-refactor/issues/02-rule-stack-selection-readable-summaries.md

## Comments

- Replaced the rule modal with a contextual Editor workbench panel wired to the selected rule stack row.
- Added a pure rule form adapter for default form state, API payload mapping, raw JSON mode, sequence step mapping, and section-aware validation errors.
- Verified with `npm run build`, `git diff --check`, and Chrome headless desktop/narrow screenshots for the target ruleset workspace.
