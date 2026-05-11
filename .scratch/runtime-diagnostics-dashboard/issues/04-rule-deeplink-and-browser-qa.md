# Rule Deep Link And Browser QA

Status: done
Type: AFK

## Parent

.scratch/runtime-diagnostics-dashboard/PRD.md

## What to build

Harden the diagnostics workflow end to end. Runtime rows should be able to focus a rule in the rules workspace when a rule ID is present, and the completed dashboard should be checked in a real browser across desktop and narrow viewport sizes.

## Acceptance criteria

- [x] Rules workspace honors a rule deep link query when the referenced rule exists.
- [x] Runtime dashboard drill-down opens the relevant ruleset workspace and carries the rule ID when present.
- [x] Browser QA covers dashboard load, request stream, filters, sort, drill-down, replay entry, and empty or sparse metrics state.
- [x] Desktop and narrow viewport screenshots show no obvious overlap, clipping, broken controls, or inaccessible primary actions.
- [x] Frontend tests, type-check, production build, backend Go tests, and `git diff --check` pass.
- [x] Issues and PRD are updated to done with verification notes.

## Blocked by

- .scratch/runtime-diagnostics-dashboard/issues/03-dashboard-diagnostics-workflow.md

## Verification

- Playwright desktop dashboard check at `http://127.0.0.1:6174/dashboard`.
- Playwright search/filter/sort check confirmed dashboard controls update the request stream.
- Playwright replay check confirmed structured simulation output without the unrelated validation panel.
- Playwright drill-down check confirmed `/rulesets/http-my-namespace-15a80b12-my-ruleset-1-397258df/rules?workbench=inspect&rule=rule-003` selects `rule-003`.
- Playwright narrow viewport check at 390x844 reported no horizontal overflow.
- `npm run test`
- `npm run type-check`
- `npm run build`
- `go test ./...`
- `git diff --check`
