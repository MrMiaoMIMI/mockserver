# Responsive workspace hardening and final QA

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Harden the redesigned Ruleset Workspace after the functional slices are complete. This slice should verify responsive layout, long-value handling, operation states, accessibility of core controls, and final build hygiene across the full workspace.

The completed slice should make the page feel coherent as one workspace across desktop and narrower viewport sizes.

## Acceptance criteria

- [x] Narrower viewport layout stacks in this order: ruleset context, primary actions, rule stack, active workbench.
- [x] Long rule IDs, selector values, condition summaries, action summaries, snapshot IDs, and diagnostic paths truncate, wrap, or expose full text predictably.
- [x] Loading, saving, validating, simulating, publishing, rollback preview, and rollback states are visible and do not conflict with each other.
- [x] Empty draft, one-rule draft, multi-rule draft, disabled rule, invalid rule, simulation match, simulation miss, snapshot list, rollback preview, and rollback confirmation have been manually checked.
- [x] Raw JSON remains available in diagnostics and advanced editors.
- [x] Existing console/workbench visual direction is preserved.
- [x] Chrome or browser-based screenshots are checked at desktop and narrower viewport sizes.
- [x] Frontend build/type checks pass.
- [x] `git diff --check` passes.

## Implementation notes

- Hardened small-viewport layout so ruleset context appears before primary actions, then rule stack, then the active workbench.
- Kept rule stack internally scrollable in stacked layouts to avoid rule rows bleeding into the workbench.
- Added safer wrapping/truncation for long page titles, meta pills, rule IDs, snapshot IDs, rollback diff paths, validation paths, and diagnostic paths.
- Added `aria-label`, `aria-selected`, focus-visible, and keyboard selection coverage for core rule and snapshot controls.

## Verification

- Final user acceptance pass completed after additional layout feedback for the ruleset header and active workbench area.
- Latest Playwright QA artifacts are kept under ignored local output storage at `.artifacts/output/playwright/`.
- Created temporary QA rulesets covering empty draft, multi-rule draft, disabled rule, invalid rule, long values, two snapshots, rollback preview, simulation match, and simulation miss; cleaned them from local MySQL after verification.
- API checks confirmed rollback preview diff with simulation payload, simulation match/miss, and invalid validation state.
- Chrome screenshots checked:
  - `/tmp/mockserver-workspace-issue7/qa-snapshots-desktop.png`
  - `/tmp/mockserver-workspace-issue7/qa-editor-430-fixed.png`
  - `/tmp/mockserver-workspace-issue7/qa-readiness-820-fixed.png`
  - `/tmp/mockserver-workspace-issue7/qa-empty-820.png`
  - `/tmp/mockserver-workspace-issue7/rollback-preview-scrolled-cdp.png`
  - `/tmp/mockserver-workspace-issue7/rollback-confirm-cdp.png`
- `npm run build` in `web/` passes.
- `git diff --check` passes.

## Blocked by

- .scratch/ruleset-workspace-refactor/issues/01-ruleset-workspace-context-baseline.md
- .scratch/ruleset-workspace-refactor/issues/02-rule-stack-selection-readable-summaries.md
- .scratch/ruleset-workspace-refactor/issues/03-contextual-rule-editor-workbench.md
- .scratch/ruleset-workspace-refactor/issues/04-context-aware-simulation-diagnostics.md
- .scratch/ruleset-workspace-refactor/issues/05-validation-publish-readiness-workflow.md
- .scratch/ruleset-workspace-refactor/issues/06-snapshot-history-rollback-workbench.md
