# Issue 04: Wire Response Editor Into Rule and Fallback Surfaces

Status: Completed

## Scope

Replace the old raw response textarea in all mock response editing surfaces.

## Changes

- Updated `RuleEditorWorkbench.vue` to use `ProtocolResponseEditor` for static actions and sequence steps.
- Updated `FallbackEditor.vue` and `NamespaceList.vue` to use the same model and build path for fallback policies.
- Updated rule form adapters so defaults and saved payloads are derived from the active protocol spec.

## Verification

- `npm run test -- src/utils/__tests__/ruleFormAdapter.test.ts src/utils/__tests__/ruleAuthoring.test.ts src/utils/__tests__/entryLists.test.ts`
- `npm run test`
- `npm run type-check`
