# Ruleset workspace context baseline

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Establish the redesigned Ruleset Workspace foundation for `/rulesets/:id/rules`. The page should orient the user around the current ruleset before deeper editing work starts: draft identity, namespace, selector, version, published state, snapshot availability, and primary actions should be visible in a compact workspace header.

This slice should keep using the existing admin APIs first. It should introduce a focused workspace state boundary for loading draft context, published state, snapshots, active workbench mode, operation loading state, and errors. If the slice proves that frontend aggregation is too noisy to keep maintainable, document that finding in the issue comments and create a follow-up backend workspace summary issue rather than mixing the API design into this slice.

## Acceptance criteria

- [x] The ruleset workspace shows draft identity, namespace, selector values, version, enabled state, rule count, published state, and snapshot availability in a clear header/context area.
- [x] Primary actions for back, refresh, settings, validate, publish, simulate, and snapshots are organized around the current ruleset instead of being scattered across unrelated panels.
- [x] The page has an explicit active workbench mode state that can later host rule detail, simulation, diagnostics, snapshots, and rollback views.
- [x] Loading, empty, not-found, and API error states are visible in the workspace context.
- [x] Existing admin APIs are used for draft, published, and snapshot data unless a follow-up backend summary issue is explicitly created from implementation evidence.
- [x] The page remains usable at desktop and narrower viewport widths.
- [x] Frontend build/type checks pass.

## Blocked by

None - can start immediately.

## Comments

- Completed with existing admin APIs; no backend workspace summary endpoint was needed for this baseline slice.
- Verified with `npm run build`, `git diff --check`, and Chrome headless desktop/narrow/not-found screenshots for the target ruleset workspace.
