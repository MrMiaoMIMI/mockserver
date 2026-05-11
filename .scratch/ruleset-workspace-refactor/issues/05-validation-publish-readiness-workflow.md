# Validation and publish readiness workflow

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Make validation and publish readiness a first-class workflow in the Ruleset Workspace. Users should be able to see whether the draft is valid, whether it differs from the current published version, what will happen when publishing, and what happened after publish completes.

This slice should reuse existing validate and publish endpoints unless implementation evidence shows a stable backend summary contract is required.

## Acceptance criteria

- [x] The workspace shows current validation status and validation issues in a structured, actionable panel.
- [x] Backend validation issues are mapped back to ruleset or rule sections where possible.
- [x] Publish action clearly indicates required validation state, in-progress state, success state, and failure state.
- [x] Publish result shows the published version or snapshot information returned by the backend.
- [x] The workspace indicates whether the draft is published, draft-only, or likely changed relative to current published data using available API data.
- [x] Errors from validate and publish appear in the relevant workflow area rather than only as generic notifications.
- [x] Existing publish audit behavior and reason prompt behavior are preserved.
- [x] Frontend build/type checks pass.

## Blocked by

- .scratch/ruleset-workspace-refactor/issues/01-ruleset-workspace-context-baseline.md
- .scratch/ruleset-workspace-refactor/issues/04-context-aware-simulation-diagnostics.md

## Comments

- Added a dedicated Ready workbench panel for validation state, backend issue mapping, publication state, publish errors, and publish success snapshot details.
- Validation and publish now report workflow-specific errors in the Ready panel; publish still uses the existing reason prompt before calling the existing publish endpoint.
- Draft publication state is inferred from current published data as draft-only, published-current, or draft-changed.
- Verified with `npm run build`, `git diff --check`, live admin validate API, and Chrome headless desktop/narrow screenshots for the Ready workbench.
