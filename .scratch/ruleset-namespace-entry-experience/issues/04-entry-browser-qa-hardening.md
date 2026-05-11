# Entry Browser QA Hardening

Status: done
Type: AFK

## Parent

.scratch/ruleset-namespace-entry-experience/PRD.md

## What to build

Verify and harden the redesigned Ruleset List and Namespace List entry pages in a real browser. The pages should remain readable and task-oriented on desktop and narrow viewport sizes.

## Acceptance criteria

- [x] Ruleset List browser QA covers search, filters, sorting, primary workspace navigation, settings entry, publish entry, long selectors, empty rulesets, and changed/draft/published states.
- [x] Namespace List browser QA covers search, filters, sorting, usage count, forward fallback, response fallback, and edit dialog entry.
- [x] Desktop and narrow viewport screenshots show no obvious overlap, clipping, broken controls, or inaccessible primary actions.
- [x] Frontend tests, type-check, production build, and backend Go tests pass.
- [x] Issues and PRD are updated to done with verification notes.

## Blocked by

- .scratch/ruleset-namespace-entry-experience/issues/02-ruleset-list-entry-redesign.md
- .scratch/ruleset-namespace-entry-experience/issues/03-namespace-fallback-entry-redesign.md

## Comments

- Browser-checked `/rulesets` and `/namespaces` with Playwright at desktop width and narrow `390x844` viewport.
- Verified console warnings/errors are clean for the checked pages and admin list endpoints return 200.
- Fixed QA findings during the pass: compacted desktop summary/toolbars, kept mobile segmented filters horizontal, hid horizontal list overflow from long selector values, and fixed entry dialog overlay clipping over the sidebar.
- Verified with `npm run test`, `npm run type-check`, `npm run build`, and `go test ./...`.
