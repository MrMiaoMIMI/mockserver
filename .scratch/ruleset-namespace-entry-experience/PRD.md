# PRD: Ruleset And Namespace Entry Experience

Status: done
Created: 2026-05-08

## Problem Statement

The ruleset rules workspace is now much stronger as the core rule management surface, but users still reach it through weaker entry pages. The Ruleset List and Namespace List currently expose the necessary objects and actions, but they do not yet guide users to the right target quickly enough.

From the user's perspective, the entry experience should answer operational questions before opening a detail workspace:

- Which rulesets are published, draft-only, changed, disabled, or potentially stale?
- Which rulesets belong to which namespace?
- Which selector hosts and paths identify a ruleset?
- Which rulesets have no rules or too many rules?
- Which namespace controls the fallback behavior when a ruleset misses or a rule misses?
- Which namespaces forward original requests by default and which return synthetic responses?
- Which namespaces are used by existing rulesets?

Without this entry-level information architecture, users can still manage rules, but they spend too much effort locating the right ruleset and understanding miss behavior.

## Solution

Redesign the Ruleset List and Namespace List as decision-oriented entry surfaces for rule management.

Ruleset List should help users find the right ruleset, understand its publish and draft state, inspect selector summaries, see namespace and rule counts, and choose the next action: open rules, edit settings, or publish.

Namespace List should explain fallback behavior clearly. Users should be able to see ruleset miss and rule miss policies at a glance, understand whether a namespace forwards original requests or returns a response, and see which namespaces are actually used by rulesets.

The implementation should keep the current console/workbench visual direction and avoid another theme pass. The key improvement is information architecture: stronger summaries, better filters, clearer sorting, tighter actions, and reusable view-model helpers that are testable outside Vue components.

## User Stories

1. As a MockServer user, I want the Ruleset List to show operational state clearly, so that I can choose the right ruleset without opening every workspace.
2. As a MockServer user, I want to see whether a ruleset is published, draft-only, changed, enabled, or disabled, so that I understand release state before acting.
3. As a MockServer user, I want to see namespace, protocol, rule count, selector host count, and selector path count, so that I can compare rulesets quickly.
4. As a MockServer user, I want long ruleset IDs and selector values to remain inspectable, so that dense technical identifiers do not become unreadable.
5. As a MockServer user, I want to search rulesets by name, ID, namespace, host, path, publish state, and rule IDs, so that I can find a target quickly.
6. As a MockServer user, I want to filter rulesets by enabled state, publish state, namespace, and empty/non-empty rules, so that I can narrow the list to the relevant set.
7. As a MockServer user, I want to sort rulesets by name, namespace, rule count, publish state, and version, so that I can scan the list in the order that matches my task.
8. As a MockServer user, I want the primary action for a ruleset to open the rules workspace, so that managing rules is the main path.
9. As a MockServer user, I want secondary actions such as settings and publish to be clear but not dominant, so that I do not trigger the wrong workflow accidentally.
10. As a MockServer user, I want empty rulesets to be visible as a special state, so that I can identify unfinished or placeholder configuration.
11. As a MockServer user, I want the Namespace List to explain fallback behavior, so that I know what happens when request matching misses.
12. As a MockServer user, I want to see ruleset miss and rule miss policies side by side, so that I can distinguish selector miss from rule miss.
13. As a MockServer user, I want forward fallback to be visually distinct from response fallback, so that I can understand whether traffic passes through.
14. As a MockServer user, I want to see fallback response status and forward timeout in the list, so that I can audit miss behavior without editing.
15. As a MockServer user, I want to see how many rulesets use a namespace, so that I understand whether editing a namespace affects active rulesets.
16. As a MockServer user, I want to filter namespaces by ruleset miss policy, rule miss policy, and usage, so that I can find fallback policies quickly.
17. As a MockServer user, I want to search namespaces by ID, name, description, fallback type, and related ruleset IDs, so that I can locate policy ownership.
18. As a MockServer user, I want namespace editing to preserve the existing fallback editor behavior, so that the list redesign does not change saved policy semantics.
19. As a MockServer user, I want list pages to remain readable on narrower viewports, so that operational triage remains usable outside a wide desktop.
20. As a MockServer developer, I want ruleset list derivation in a testable helper, so that filtering, sorting, and publish-state labels are not embedded in the page component.
21. As a MockServer developer, I want namespace list derivation in a testable helper, so that fallback summaries and usage counts are consistent.
22. As a MockServer developer, I want tests to cover external view behavior, so that future UI changes can safely preserve search, filter, sort, and state semantics.
23. As a MockServer maintainer, I want this work split into small vertical slices, so that each entry page improvement is independently demoable.

## Implementation Decisions

- Treat this as entry-page information architecture work, not a theme redesign.
- Keep Vue 3, TypeScript, Vite, Element Plus, Pinia, and SCSS.
- Keep existing admin API contracts. Use existing draft ruleset, published ruleset, and namespace list endpoints.
- Do not add backend endpoints unless the frontend cannot derive a stable entry summary from existing data. This is not expected.
- Add a ruleset entry view-model helper that derives search text, publish state, status labels, selector counts, namespace grouping, sort keys, and filtered rows.
- Add a namespace entry view-model helper that derives fallback labels, fallback tones, usage counts, related ruleset IDs, search text, sort keys, and filtered rows.
- Ruleset List should emphasize opening the rules workspace as the primary action.
- Ruleset List should preserve settings and publish workflows as secondary actions.
- Ruleset List should compare draft and current published data while ignoring version noise, so that "changed" means meaningful content difference.
- Namespace List should fetch rulesets as context so namespace usage can be shown without backend changes.
- Namespace List should keep the existing fallback create/edit contract and default forward behavior.
- Both entry pages should use compact controls, stable row/card dimensions, readable metadata, and inspectable long values.
- Both entry pages should avoid nested decorative cards. Cards or rows are acceptable for repeated business items.

## Testing Decisions

- Good tests should verify list behavior and derived view-model semantics, not exact CSS or component internals.
- Add unit tests for ruleset entry helpers covering publish state, changed state, search, filters, sorting, selector counts, and empty-rule state.
- Add unit tests for namespace entry helpers covering fallback labels, forward/response tones, ruleset usage counts, search, filters, and sorting.
- Reuse the existing Vitest setup and pure TypeScript helper test style already used for rule collection, diagnostics, and rule form adapter logic.
- Use Playwright browser checks for Ruleset List and Namespace List at desktop and narrow viewport widths.
- Use frontend build/type checks and backend Go tests as regression verification.

## Out of Scope

- Reworking the ruleset rules workspace again.
- Changing rule engine matching semantics.
- Changing namespace fallback runtime semantics.
- Changing backend storage schema.
- Adding authentication or permission behavior.
- Building a full E2E test suite for every frontend route.
- Replacing Element Plus or the current frontend stack.

## Further Notes

This work follows the Rule Management Workbench redesign. The goal is to make the upstream entry pages match the improved depth of the rules workspace, so users can move from triage to management with less uncertainty.

## Completion Notes

- Implemented shared entry view-model helpers for ruleset and namespace list derivation.
- Redesigned Ruleset List around publish state, rule count, selector summary, filters, sorting, and rules workspace as the primary action.
- Redesigned Namespace List around fallback policy, namespace usage, related rulesets, filters, sorting, and an edit dialog with policy preview.
- Hardened responsive behavior for desktop and narrow browser widths, including long selector values and dialog layering over the app shell.
- Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks on `/rulesets` and `/namespaces`.
