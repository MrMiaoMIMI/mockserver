# Name-first rule creation

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-workspace-interaction-simplification/PRD.md

## What to build

Make create-rule authoring user-facing by asking for a rule name first and generating Rule ID automatically. Existing edit behavior should preserve stable IDs, and advanced raw JSON should still expose the full payload for power users.

## Acceptance criteria

- [x] Create mode does not require the user to manually fill Rule ID in the primary identity form.
- [x] Create mode generates a stable Rule ID from the rule name.
- [x] Generated Rule IDs avoid duplicates in the current ruleset.
- [x] Create mode shows the generated Rule ID as secondary metadata.
- [x] Edit mode keeps the existing Rule ID read-only.
- [x] Raw JSON mode can still represent and validate the full rule payload.
- [x] Rule ID generation and locked-id behavior are covered by unit tests.

## Blocked by

- .scratch/rule-workspace-interaction-simplification/issues/03-rule-card-action-reduction.md

## Completion notes

Implemented in the rule editor and form adapter. Create mode now asks for Name first and displays Generated Rule ID as secondary metadata. ID generation is slug-based, duplicate-safe, and covered by unit tests alongside locked edit-ID behavior.
