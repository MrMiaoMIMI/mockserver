# Snapshot history and rollback workbench

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Redesign snapshot history and rollback into a clear rollback workbench. Users should understand which snapshot they are selecting, how it differs from the current published version, and what impact rollback has before executing it.

Rollback preview should reuse the same event payload used by simulation where possible, so users can evaluate rollback behavior against a meaningful request example.

## Acceptance criteria

- [x] Snapshot history shows snapshot ID, version, published time, operator, reason, and source snapshot when available.
- [x] Selecting a snapshot opens a rollback preview flow without losing ruleset workspace context.
- [x] Rollback preview shows current versus target diff summary, changed state, ruleset field diffs, and rule diffs in readable structured panels.
- [x] Rollback preview can run with the current simulation event payload when available.
- [x] Rollback confirmation clearly states the target snapshot and ruleset.
- [x] Rollback success updates published state, snapshot history, and diagnostic result.
- [x] Rollback errors appear in the rollback workflow area.
- [x] Existing rollback preview and rollback backend behavior is preserved.
- [x] Frontend build/type checks pass.

## Implementation notes

- Added `SnapshotRollbackWorkbench.vue` for snapshot history, rollback preview, rollback action, inline errors, and success state.
- Added `snapshotRollback.ts` view helpers to normalize snapshot history and rollback diff summaries.
- Typed rollback diff payloads in `mockserver.ts` so frontend panels use the backend structured diff contract directly.
- Kept existing rollback preview and rollback API calls; only moved their results and errors into the new workbench flow.

## Verification

- `npm run build` in `web/` passes.
- `git diff --check` passes.
- Screenshot QA completed for desktop and narrow local views:
  - `/tmp/mockserver-workspace-issue6/snapshots-desktop.png`
  - `/tmp/mockserver-workspace-issue6/snapshots-narrow-full.png`

## Blocked by

- .scratch/ruleset-workspace-refactor/issues/04-context-aware-simulation-diagnostics.md
- .scratch/ruleset-workspace-refactor/issues/05-validation-publish-readiness-workflow.md
