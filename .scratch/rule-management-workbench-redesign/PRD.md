# PRD: Rule Management Workbench Redesign

Status: done
Created: 2026-05-08

## Problem Statement

The latest frontend optimization improved the visual theme and made the ruleset rules page look more polished, but the core product problem is still the rule management experience. Users visit `/rulesets/:id/rules` primarily to manage, understand, edit, validate, simulate, and publish rules. The page should therefore be designed around rule operations and rule diagnosis, not around theme presentation or generic page decoration.

The current experience still has several structural problems:

- RuleManager and Workbench are both important, but the task boundary between them is not clear enough.
- The rule list is not yet optimized as the primary management surface for scanning, comparing, filtering, selecting, reordering, duplicating, enabling, disabling, and deleting rules.
- The Workbench is still too crowded for serious rule editing and diagnosis because several complex jobs compete for the same space.
- Ruleset-level information is useful, but it can consume attention and layout space that should belong to rule management.
- Some actions are visible as buttons before their meaning or workflow is obvious, so users need to infer how validation, simulation, snapshots, publishing, and rule edits relate to each other.
- The previous refactor improved the shell of the ruleset workspace, but this follow-up should redesign the information architecture and interaction model more deeply.

This PRD is for a second-phase refactor. The goal is not another theme pass. The goal is to make RuleManager and Workbench feel like a purpose-built rule management product surface.

## Solution

Redesign the ruleset rules page around a clear rule management task model:

- The page should first help users understand which ruleset they are editing.
- The main workspace should then help users browse and manage rules efficiently.
- The selected rule should drive the Workbench context.
- The Workbench should expand into task-oriented modes for editing, inspecting, simulating, validating, publishing, and reviewing snapshots.
- Ruleset metadata should be compact by default, with richer detail available through deliberate interaction.

RuleManager should become the durable management surface. It should support dense but readable rule scanning, fast operations, stable selection, rule comparison signals, and safe destructive actions. It should not feel like a small list attached below a large header.

Workbench should become a contextual task surface. Instead of forcing all complex tools into a cramped panel, it should expose the right task at the right time: selected rule overview, structured rule editor, event simulation, validation diagnostics, publish readiness, snapshot history, and rollback preview.

The design should preserve the current frontend stack and the recently improved visual direction where it helps readability. However, theme is not a success criterion for this PRD. The success criterion is whether users can manage rules with less effort, less ambiguity, and less context switching.

## User Stories

1. As a MockServer user, I want the rules page to prioritize rule management, so that the most important screen area is used for the work I came to do.
2. As a MockServer user, I want ruleset identity to be visible but compact, so that I know my context without losing vertical space.
3. As a MockServer user, I want ruleset metadata to be collapsed or summarized by default, so that secondary information does not compete with rule operations.
4. As a MockServer user, I want namespace, protocol, selector, draft version, and publish state to be shown in one compact context row, so that I can orient myself quickly.
5. As a MockServer user, I want detailed ruleset metadata available on demand, so that I can inspect it when needed without always seeing it.
6. As a MockServer user, I want the difference between ruleset context and rule management to be visually obvious, so that I understand the page structure immediately.
7. As a MockServer user, I want RuleManager to be the main management area, so that browsing and operating on rules feels central rather than secondary.
8. As a MockServer user, I want rule rows to be readable at normal browser zoom, so that I do not need to strain to inspect IDs, names, conditions, or actions.
9. As a MockServer user, I want each rule row to show priority, enabled state, name, rule ID, condition summary, action summary, and key status signals, so that I can understand rule behavior before opening it.
10. As a MockServer user, I want long rule IDs, selector values, paths, and headers to be inspectable without breaking layout, so that dense data stays usable.
11. As a MockServer user, I want to search rules by name, ID, condition text, action text, and status, so that I can find relevant rules quickly.
12. As a MockServer user, I want to filter rules by enabled state, validation state, match result, and action type, so that I can narrow the list to the rules I care about.
13. As a MockServer user, I want rule ordering to be visible and easy to adjust, so that I can reason about match priority.
14. As a MockServer user, I want to select a rule without opening a disruptive modal, so that I keep the list context while inspecting details.
15. As a MockServer user, I want the selected rule to clearly drive the Workbench, so that I know which rule I am editing, simulating, or validating.
16. As a MockServer user, I want create, duplicate, edit, enable, disable, reorder, and delete operations to be close to the rule list, so that rule management actions are easy to discover.
17. As a MockServer user, I want destructive actions to include precise target names and IDs, so that I do not delete or rollback the wrong rule behavior.
18. As a MockServer user, I want bulk or repeated operations to avoid unnecessary page movement, so that managing many rules remains efficient.
19. As a MockServer user, I want Workbench modes to be task-oriented instead of a generic collection of tabs, so that the page guides me through the actual workflow.
20. As a MockServer user, I want the Workbench to show a selected rule overview first, so that I can inspect behavior before deciding to edit or simulate.
21. As a MockServer user, I want rule editing to happen in a spacious Workbench mode, so that condition and action fields are not cramped.
22. As a MockServer user, I want rule editing sections to separate identity, matching condition, response action, and raw JSON, so that complex rule data is easier to reason about.
23. As a MockServer user, I want condition editing to support structured predicate, ALL, ANY, NOT, CEL, and raw JSON forms, so that simple and advanced rules are both manageable.
24. As a MockServer user, I want action editing to show controls specific to static, template, CEL, sequence, and webhook responses, so that irrelevant fields do not add noise.
25. As a MockServer user, I want raw JSON editing to remain available for advanced cases, so that the UI does not block uncommon rule definitions.
26. As a MockServer user, I want validation errors to link back to the relevant rule, condition, action, or JSON section, so that failures are actionable.
27. As a MockServer user, I want simulation to use the current namespace, selector, and selected rule as context, so that default test events are meaningful.
28. As a MockServer user, I want to simulate against draft and published behavior without losing the current event input, so that I can compare changes before publishing.
29. As a MockServer user, I want simulation results to show selector match, rule match, fallback state, response result, and miss reason as structured diagnostics, so that I can understand outcomes quickly.
30. As a MockServer user, I want rules that matched, missed, or were skipped during simulation to be reflected back in RuleManager, so that diagnosis connects to the rule list.
31. As a MockServer user, I want validation, simulation, publish, and rollback outputs to be organized by workflow, so that I do not have to interpret unrelated raw JSON blocks.
32. As a MockServer user, I want raw response payloads to remain available after structured summaries, so that deep debugging is still possible.
33. As a MockServer user, I want publish readiness to show whether the current draft is valid and different from published state, so that I know whether publishing is necessary.
34. As a MockServer user, I want snapshot history and rollback preview to be accessible without taking over the main rule list, so that I can review history while keeping current context.
35. As a MockServer user, I want smaller screens to preserve the same task order, so that rule management remains usable even when the layout stacks vertically.
36. As a MockServer user, I want keyboard focus, tab order, and visible focus states to be reliable, so that dense rule operations remain accessible.
37. As a MockServer developer, I want RuleManager state to be separated from page rendering, so that rule selection, filtering, sorting, and operations are easier to test.
38. As a MockServer developer, I want Workbench task state to be explicit, so that transitions between inspect, edit, simulate, validate, publish, snapshot, and rollback modes are predictable.
39. As a MockServer developer, I want rule summary, condition summary, action summary, and diagnostic summary generation to be centralized, so that the list and Workbench speak the same language.
40. As a MockServer developer, I want API payload mapping to remain isolated in adapters, so that UI improvements do not accidentally change backend rule contracts.
41. As a MockServer maintainer, I want this refactor split into independently shippable issues, so that we can verify one vertical workflow at a time.

## Implementation Decisions

- Treat this as an information architecture and interaction refactor, not a visual theme refactor.
- Keep the current Vue 3, TypeScript, Vite, Element Plus, Pinia, and SCSS stack.
- Keep the existing ruleset rules route contract.
- Keep the current backend rule, ruleset, namespace, validation, simulation, publish, snapshot, and rollback contracts unless a narrow backend improvement is justified.
- Do not introduce compatibility layers for old frontend layout behavior. This project is still new, so implement the best current design directly.
- Preserve the recent readable console/workbench visual direction where it supports the task, but do not spend the PRD on color or theme changes.
- Redesign the page structure into compact ruleset context, primary rule management surface, and contextual Workbench.
- Make ruleset context compact by default. Full details should be available through a details drawer, popover, disclosure panel, or equivalent inspect-on-demand interaction.
- RuleManager should own browse and manage workflows: list density, search, filter, sorting, selection, priority movement, create, duplicate, enable, disable, delete, and row-level status signals.
- Workbench should own selected-rule and workflow-specific tasks: rule overview, rule editing, simulation, validation diagnostics, publish readiness, snapshot history, and rollback preview.
- Workbench mode should be explicit in frontend state. Avoid hidden coupling where a result panel changes meaning based on unrelated previous actions.
- The selected rule should be a first-class state concept. Selection should remain stable after filtering, saving, reordering, simulation, validation, and refresh where possible.
- Search and filtering should not mutate the underlying rule order. Priority order remains the source of truth.
- Reordering should make priority consequences visible before or immediately after the operation.
- Duplicate rule should be supported as a rule management workflow. Use existing create semantics unless backend support is needed for safer ID generation.
- Rule delete should clearly show the target rule name and ID.
- Rule enabled or disabled state should be visible in both RuleManager and selected-rule Workbench overview.
- RuleManager rows should avoid overloaded tags. Use tags only for high-signal state such as disabled, invalid, matched, miss, draft-only, or action type.
- Long technical values should be truncated only with inspect-on-demand behavior, not silently hidden.
- Workbench should avoid duplicate page titles and repeated headers. Every header in the Workbench should explain the current task or selected target.
- Rule editing should be spacious enough for structured condition and action configuration. If a side panel is too cramped, use a full Workbench mode rather than a modal or narrow drawer.
- Keep raw JSON editors available but position them as advanced tools inside the relevant task.
- Simulation defaults should be generated from current ruleset and selected rule context. Prefer namespace, protocol, selector host/path, method, headers, and simple condition hints when safe.
- Simulation should be able to update RuleManager status signals after a run, such as matched rule, missed rule, selector miss, rule miss, or fallback.
- Validation should be able to update RuleManager status signals after a run, such as invalid rule count or selected rule error state.
- Publish readiness should summarize validation state, draft-vs-published state, and publish outcome before raw JSON.
- Snapshot and rollback workflows should remain present, but they should not dominate the default rule management view.
- Introduce or strengthen a rule workspace state module that owns loading, selected rule, active Workbench mode, search, filters, sort state, operation loading states, diagnostics, and refresh behavior.
- Introduce or strengthen rule collection view helpers that derive visible rows, grouped status counts, empty states, and selected-row stability from rule data.
- Continue using rule form adapter boundaries for mapping API rule payloads into editable UI state and back.
- Introduce or strengthen Workbench task model helpers that define allowed transitions between inspect, edit, create, duplicate, simulate, validate, publish, snapshots, and rollback.
- Introduce or strengthen diagnostic view-model helpers that transform simulation, validation, publish, and rollback responses into structured UI panels.
- Backend changes are optional. Add a backend workspace summary endpoint only if frontend composition remains materially complex after modularizing the client.
- If added, a backend workspace summary endpoint should expose stable domain summaries rather than visual layout data.
- Do not change rule matching semantics, fallback semantics, namespace default behavior, publish semantics, rollback semantics, or storage schema as part of this PRD.

## Testing Decisions

- Good tests should verify task behavior, state transitions, data transformations, and API contract preservation. They should not assert fragile theme details.
- Add unit tests for rule collection view helpers: search, filtering, selected-row stability, ordering, empty states, and status count derivation.
- Add unit tests for rule summary helpers: condition summary, action summary, disabled state, invalid state, matched state, miss state, and long-value handling.
- Add unit tests for Workbench task model helpers: inspect-to-edit, inspect-to-simulate, create-to-save, duplicate-to-save, validate-to-diagnostics, publish-to-result, snapshot-to-rollback-preview, and cancel flows.
- Add or extend tests for rule form adapters so that UI form changes preserve API rule payload shape.
- Add tests for simulation default event generation from namespace, selector, selected rule, and common condition structures.
- Add tests for diagnostic view-model helpers so that selector miss, rule miss, fallback, matched rule, validation error, publish success, publish failure, rollback preview, and rollback result are rendered as distinct structured states.
- If backend workspace summary work is introduced, add backend tests that prove the endpoint returns stable domain summaries and does not change existing admin API behavior.
- Use Playwright or equivalent browser QA to verify the main workflows on desktop and narrow viewport sizes.
- Browser QA should cover at least: no rules, one rule, many rules, long rule IDs, long selector values, disabled rule, invalid rule, create rule, duplicate rule, edit rule, reorder rule, delete rule, simulation match, selector miss, rule miss, fallback, validate, publish, snapshot preview, and rollback preview.
- Build verification should include frontend type checking or production build.
- Regression verification should confirm that existing namespace fallback behavior and ruleset publish or rollback behavior are unchanged.

## Out of Scope

- A full visual theme redesign.
- Replacing the current frontend framework or UI library.
- Redesigning the dashboard, namespace list, or ruleset list except where navigation into the rules page requires minor consistency updates.
- Changing rule engine matching semantics.
- Changing default namespace fallback behavior.
- Changing backend storage schema.
- Adding multi-user collaboration or conflict resolution.
- Adding authentication, permission, or audit systems beyond existing behavior.
- Building the SDK package workstream.
- Rewriting all frontend components unrelated to rule management.

## Further Notes

This PRD follows the completed `ruleset-workspace-refactor` work. That earlier work made the page more coherent and introduced a stronger workspace shell. This PRD is intentionally narrower and deeper: it treats RuleManager and Workbench as the product core and asks for a task-centered redesign of rule management.

The expected outcome is not that the page looks like a different theme. The expected outcome is that users can scan, edit, simulate, validate, publish, and review rules with less layout friction and less uncertainty about what to do next.

## Comments

- Implemented through the eight local issues under `.scratch/rule-management-workbench-redesign/issues/`.
- Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for the target rules page across desktop and narrow viewports.
