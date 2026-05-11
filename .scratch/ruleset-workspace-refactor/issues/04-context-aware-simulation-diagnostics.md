# Context-aware simulation and diagnostics

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Turn simulation into a context-aware diagnostic workflow. The initial event payload should be derived from the current ruleset and selected rule instead of using a static event that may point to an unrelated namespace. Simulation results should be presented as structured diagnostics first, with raw JSON still available for advanced inspection.

This slice should extract helpers for simulation event defaults and diagnostic view models so the behavior is testable and reusable.

## Acceptance criteria

- [x] The default simulation event uses the current ruleset namespace and protocol.
- [x] The default simulation event prefers selector host and path hints when available.
- [x] The default simulation event can use simple selected-rule condition hints where safe.
- [x] Users can switch simulation target between draft and published without losing the event payload.
- [x] Simulation diagnostics show matched state, matched ruleset, matched rule, fallback state, fallback reason, response status, and candidate rules when available.
- [x] Selector checks and condition explanations are rendered as readable diagnostic traces when present.
- [x] Raw JSON remains available after structured diagnostics.
- [x] The diagnostic result persists until superseded by the next validation, simulation, publish, rollback preview, or rollback action.
- [x] Frontend build/type checks pass.

## Blocked by

- .scratch/ruleset-workspace-refactor/issues/01-ruleset-workspace-context-baseline.md
- .scratch/ruleset-workspace-refactor/issues/02-rule-stack-selection-readable-summaries.md

## Comments

- Added reusable simulation helpers for context-derived event defaults and structured diagnostic view models.
- Simulation defaults now use current ruleset namespace/protocol, selector hints, and safe selected-rule condition hints without overwriting user-customized payloads.
- Result inspector now renders matched/fallback/status/candidate summaries plus selector checks and condition traces before raw JSON.
- Verified with `npm run build`, `git diff --check`, live admin simulate API response, and Chrome headless screenshot for the simulation workbench.
