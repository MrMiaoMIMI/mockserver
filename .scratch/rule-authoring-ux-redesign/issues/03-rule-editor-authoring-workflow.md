# Rule Editor Authoring Workflow

Status: done
Type: AFK

## Parent

.scratch/rule-authoring-ux-redesign/PRD.md

## What to build

Redesign the rule editor workflow around a guided authoring overview, condition builder, action builder, live validation, advanced JSON, and draft simulation. The editor should make the common static response flow clear while keeping advanced action types available.

## Acceptance criteria

- [x] The editor shows a compact authoring overview with rule identity, condition summary, action summary, and readiness.
- [x] The condition section uses the preset-aware condition builder.
- [x] The action section makes static response authoring clear and describes each action type.
- [x] Headers and body inputs surface JSON validation problems before save.
- [x] The editor can run a draft simulation from the current generated or edited event.
- [x] Simulation output uses the existing structured result inspector.
- [x] Save and raw JSON synchronization still produce the existing rule payload shape.
- [x] Frontend tests, type-check, and build pass.

## Blocked by

- .scratch/rule-authoring-ux-redesign/issues/01-condition-preset-builder.md
- .scratch/rule-authoring-ux-redesign/issues/02-authoring-view-model-and-readiness.md

## Verification

- Updated `RuleEditorWorkbench.vue` with authoring readiness cards, condition/action summaries, guided action cards, raw JSON sync, and a preview tab backed by `ResultInspector`.
- Added backend draft override simulation support across request model, view, service interface, and service implementation.
- Verified with `npm run test`, `npm run type-check`, `npm run build`, and `go test ./...`.
