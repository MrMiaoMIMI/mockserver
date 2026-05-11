# Readable typography baseline for shell, controls, and Dashboard

Status: done
Type: AFK

## Parent

.scratch/frontend-readable-typography/PRD.md

## What to build

Improve the shared frontend readability baseline so the app shell, common controls, metadata chips, filter controls, Element Plus components, and Dashboard stop relying on text that is too small for daily operation.

This slice should establish the shared type scale and baseline visual rhythm that later page-specific slices build on. The completed slice should be visible immediately in the shell and Dashboard without changing backend behavior or the product's existing console/workbench direction.

## Acceptance criteria

- [x] The shared frontend type scale makes body text, form fields, labels, chips, table text, and code-like utility text readable at normal viewing distance.
- [x] Element Plus overrides for buttons, inputs, textareas, form labels, dialogs, tables, and empty states align with the improved type scale.
- [x] App shell, page headers, metadata chips, filter controls, and Dashboard metric/detail rows remain compact but visibly more readable.
- [x] Dashboard keys, labels, captions, and metric support text no longer use unreadably small text.
- [x] Desktop and narrower viewport checks show no text overlap, awkward clipping, or control resizing caused by the larger type.
- [x] Frontend build/type checks pass.

## Blocked by

None - can start immediately.

## Comments

- Completed in frontend readability pass. Verified with `npm run build`, `git diff --check`, small-token style audit, and Chrome headless desktop/narrow screenshots.
