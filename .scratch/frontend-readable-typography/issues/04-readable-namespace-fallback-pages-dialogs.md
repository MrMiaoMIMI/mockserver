# Namespace fallback pages and dialogs become readable

Status: done
Type: AFK

## Parent

.scratch/frontend-readable-typography/PRD.md

## What to build

Improve the readability of namespace management surfaces, including namespace cards, fallback summaries, fallback editors, namespace dialogs, and related form fields.

This slice should make response versus forward fallback policies easy to understand at a glance while preserving the existing namespace workflow and explicit fallback editing behavior.

## Acceptance criteria

- [x] Namespace names, IDs, descriptions, fallback labels, fallback summaries, and metadata chips are readable in the namespace list.
- [x] Fallback summary values and forward/response details are readable without requiring browser zoom.
- [x] Fallback editor controls, labels, timeout fields, response body fields, and helper text remain compact but legible in the namespace dialog.
- [x] Long fallback details truncate, wrap, or expose full text predictably without layout jumps.
- [x] The namespace page and dialogs remain usable at desktop and narrower widths.
- [x] Frontend build/type checks pass.

## Blocked by

- .scratch/frontend-readable-typography/issues/01-readable-typography-baseline-shell-controls-dashboard.md

## Comments

- Completed in frontend readability pass. Verified with `npm run build`, `git diff --check`, small-token style audit, and Chrome headless desktop/narrow screenshots.
