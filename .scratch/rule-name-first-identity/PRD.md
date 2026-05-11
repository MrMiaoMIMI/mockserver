# PRD: Rule Name First Identity

Status: done
Created: 2026-05-11
Completed: 2026-05-11

## Problem Statement

Rules now have both a system-generated Rule ID and a user-defined Rule Name, but the interface still gives Rule ID the strongest visual weight in the Rule Manager and workspace titles. This creates a mismatch: Rule ID is stable and useful for machines, diagnostics, diffs, snapshots, and API references, but it is not the most meaningful identifier for a user trying to understand what a mock rule does.

The current create flow already generates Rule ID from the name, but Rule Name remains optional and existing rule cards still read as ID-first. Users see values such as `rule-002` more prominently than the actual business purpose of the rule. The product should treat Rule Name as the user-facing identity and Rule ID as secondary technical metadata.

## Solution

Make Rule Name the primary identity for authoring and display:

- Rule Name is required for user-created and edited rules.
- Rule ID remains generated, stable, unique, and machine-owned.
- Rule lists, workspace titles, editor titles, and inspect panels present Rule Name first.
- Rule ID remains visible as secondary metadata for copying, debugging, raw JSON, API routes, simulation traces, and snapshot diffs.
- Editing Rule Name after creation does not mutate Rule ID.
- Duplicate rules get a copied Rule Name and a new generated Rule ID.
- Existing unnamed data is handled gracefully in display but must get a Rule Name before save/publish.

## User Stories

1. As a MockServer user, I want each rule to have a required name, so that every rule has a human-readable purpose.
2. As a MockServer user, I want Rule Manager cards to show the rule name first, so that I can scan rules by meaning rather than generated IDs.
3. As a MockServer user, I want generated Rule IDs to be visually secondary, so that technical identifiers do not dominate the workspace.
4. As a MockServer user, I want to still see the Rule ID, so that I can debug API traces and simulation output when needed.
5. As a MockServer user, I want create mode to require Name, so that I cannot save an unnamed rule.
6. As a MockServer user, I want the generated Rule ID to update before save while I edit the Name, so that I can inspect what will be persisted.
7. As a MockServer user, I want the generated Rule ID to avoid duplicates, so that saving succeeds without manual ID management.
8. As a MockServer user, I want editing an existing rule's Name to keep the existing Rule ID, so that references remain stable.
9. As a MockServer user, I want duplicate rules to receive a copied Name and unique Rule ID, so that copies are understandable and safe.
10. As a MockServer user, I want the selected-rule context to use Rule Name first, so that the right workspace feels user-centered.
11. As a MockServer user, I want inspect and edit titles to use Rule Name first, so that the active work is clear.
12. As a MockServer user, I want fallback text for old unnamed data, so that old records do not render as blank cards.
13. As a MockServer user, I want save/publish validation to block unnamed rules, so that the product remains clean after migration.
14. As a MockServer developer, I want backend validation to enforce Rule Name, so that API callers and UI callers share the same rule contract.
15. As a MockServer developer, I want frontend form validation to mirror backend validation, so that users get immediate feedback.
16. As a MockServer developer, I want rule display labels centralized, so that list, workspace, and diagnostics do not drift in naming behavior.
17. As a MockServer developer, I want tests to cover required names and stable generated IDs, so that the name-first contract does not regress.
18. As a MockServer developer, I want examples and bootstrap data to include Rule Names, so that demo data shows the intended product language.

## Implementation Decisions

- Treat Rule ID as a required machine identifier that is generated on create and stable after save.
- Treat Rule Name as a required human identifier for the normal form path and backend validation.
- Keep Rule ID in JSON payloads and route paths; do not remove or rename backend API fields.
- Remove optional typing for Rule Name on the frontend domain type where practical.
- Backend rule-set validation should fail when a rule name is blank.
- Frontend form adapter should fail when a rule name is blank.
- The editor's identity readiness text should say Rule Name, generated Rule ID, priority, and enabled state are valid.
- Rule Manager should use Rule Name as the card title and show Rule ID as secondary metadata.
- Rule Workspace titles should use Rule Name first and only fall back to Rule ID when necessary for old data.
- Duplicate behavior should copy the rule name and regenerate the ID from the copied name.
- Raw JSON remains available for power users, but raw JSON validation still requires `name`.

## Testing Decisions

- Unit tests should cover Rule Name required validation in the frontend form adapter.
- Unit tests should cover generated ID behavior from Rule Name and edit-mode locked ID behavior.
- Backend tests should cover validation failure for blank rule names.
- Existing engine, service, controller, and frontend tests should keep passing with sample data updated to include rule names.
- Browser QA should verify the target Ruleset Workspace page shows name-first cards and name-first create/edit behavior.
- Good tests should assert user-visible behavior and validation contracts, not CSS implementation details.

## Out of Scope

- Removing Rule ID from backend APIs.
- Changing simulation trace payload shape.
- Changing snapshot diff payload shape.
- Database schema changes.
- Full migration scripts for existing persisted local database rows.
- Drag-and-drop sorting.
- Bulk rename tools.

## Further Notes

This PRD builds on the Rule Workspace interaction simplification pass. The important product boundary is simple: Rule Name is for users; Rule ID is for stable system references.

## Completion Notes

- Backend validation now rejects blank rule names while preserving required generated rule IDs.
- Frontend domain types, form adapter, raw JSON validation, and editor UX now treat Rule Name as required.
- Rule Manager, workspace titles, inspect panels, create form, duplicate flow, and ruleset drawer inventory now display Rule Name first and Rule ID as technical metadata.
- Existing unnamed local data displays `Untitled rule` safely, but save/publish validation now requires a real name.
- Verified with `go test ./...`, `npm run type-check`, `npm test`, `npm run build`, `git diff --check`, and desktop Playwright QA on `/rulesets/http-default-test1-b03f4c4f/rules`.
