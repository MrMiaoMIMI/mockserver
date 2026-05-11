# Theme and shell clarity baseline

Status: done
Type: AFK

## Parent

.scratch/frontend-uiux-redesign/PRD.md

## What to build

Create the redesigned visual foundation for MockServer: a light content canvas, restrained dark sidebar, clear surface hierarchy, blue primary actions, semantic state colors, and aligned Element Plus component styling. This slice should make the app shell, shared page container, common controls, dialogs, popovers, tags, and utility classes feel coherent before page-specific work begins.

## Acceptance criteria

- [x] The app uses a light main canvas with clear surface, panel, control, and border contrast.
- [x] The sidebar remains dark and readable while the header/content areas become visually lighter and cleaner.
- [x] Blue is used for primary actions and selected UI states; green is reserved for success/enabled/running states.
- [x] Element Plus buttons, inputs, selects, dialogs, dropdowns, popovers, tags, tabs, tables, empty states, and alerts align with the new tokens.
- [x] Shared chips, code chips, icon buttons, panels, focus rings, and scrollbars remain readable and consistent.
- [x] The frontend build passes.

## Blocked by

None - can start immediately.

## Comments

- Completed in the frontend UI/UX redesign pass. Verified with `npm run build`, browser screenshots, and `git diff --check`.
