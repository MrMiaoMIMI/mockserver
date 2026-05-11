# Ruleset list and selector values become readable

Status: done
Type: AFK

## Parent

.scratch/frontend-readable-typography/PRD.md

## What to build

Improve the readability of the ruleset list page, especially code-like operational values such as ruleset IDs, namespaces, selector hosts, selector paths, status chips, and card statistics.

This slice should make ruleset comparison and selector inspection comfortable while preserving the compact list/card workflow and existing copy-to-inspect behavior for long selector values.

## Acceptance criteria

- [x] Ruleset names, ruleset IDs, namespace values, status chips, selector labels, selector values, and card stats are readable without opening dialogs.
- [x] Selector value chips and hidden-value popovers use readable code text and preserve copy behavior.
- [x] Long host and path values truncate, wrap, or expose full text predictably without resizing cards unexpectedly on hover.
- [x] The ruleset list remains scannable at desktop size and usable at narrower widths.
- [x] A style audit confirms remaining smallest-token uses on this page and selector components are intentional.
- [x] Frontend build/type checks pass.

## Blocked by

- .scratch/frontend-readable-typography/issues/01-readable-typography-baseline-shell-controls-dashboard.md

## Comments

- Completed in frontend readability pass. Verified with `npm run build`, `git diff --check`, small-token style audit, and Chrome headless desktop/narrow screenshots.
