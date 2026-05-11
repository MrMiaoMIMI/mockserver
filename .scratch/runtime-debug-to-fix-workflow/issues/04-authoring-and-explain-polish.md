# Authoring and explain polish

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/runtime-debug-to-fix-workflow/PRD.md

## What to build

Refine the seeded rule authoring and simulation explain presentation so the debug-to-fix workflow feels intentional rather than a hidden prefill. The editor should label runtime-seeded drafts clearly, keep normal validation intact, and make simulation output readable as a diagnosis path.

## Acceptance criteria

- [x] Rule Editor clearly labels when a create form was seeded from a runtime request.
- [x] Seeded forms still use existing section readiness, inline validation, preview, raw JSON, and save behavior.
- [x] Simulation output highlights selector checks and condition explanations as a diagnosis path.
- [x] The debug source context remains visible while using simulation and editor tasks.
- [x] No existing rule editing or preview behavior regresses.

## Completion notes

- Rule Editor create mode accepts a runtime seed rule and displays the runtime source label.
- ResultInspector now includes a compact diagnosis path before detailed explain sections.
- Existing unit tests and browser QA passed.

## Blocked by

- .scratch/runtime-debug-to-fix-workflow/issues/03-workspace-debug-intake.md
