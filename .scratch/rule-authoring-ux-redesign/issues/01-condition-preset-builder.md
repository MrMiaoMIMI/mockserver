# Condition Preset Builder

Status: done
Type: AFK

## Parent

.scratch/rule-authoring-ux-redesign/PRD.md

## What to build

Add a guided condition preset layer to the rule authoring experience. Users should be able to add common request predicates such as method, path, query, header, body field, host, and trace ID without memorizing internal field names, while preserving ALL, ANY, NOT, CEL, and raw JSON paths.

## Acceptance criteria

- [x] The condition builder exposes common predicate presets for method, path, query, header, body field, host, and trace ID.
- [x] Presets generate the same condition payload shape currently accepted by the backend.
- [x] Predicate rows show readable preset labels and operator labels.
- [x] Predicate rows support JSON-safe value editing for strings, numbers, booleans, arrays, and objects.
- [x] Incomplete predicates show inline validation without waiting for save.
- [x] Existing ALL, ANY, NOT, CEL, and raw JSON condition flows remain available.
- [x] Unit tests cover preset payload generation, readable labels, and validation.

## Blocked by

None - can start immediately.

## Verification

- Added and covered preset payload generation, key resolution, JSON-safe value parsing, and validation in `web/src/utils/__tests__/ruleAuthoring.test.ts`.
- Confirmed through Playwright that the editor condition tab exposes preset controls and remains readable in the live ruleset workspace.
