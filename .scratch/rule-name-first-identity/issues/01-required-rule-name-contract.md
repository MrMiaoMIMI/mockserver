# Required rule name contract

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-name-first-identity/PRD.md

## What to build

Make Rule Name a required part of the rule contract across backend validation, frontend form validation, examples, and tests while preserving generated stable Rule IDs.

## Acceptance criteria

- [x] Backend validation rejects rules with a blank name.
- [x] Frontend form validation rejects rules with a blank name.
- [x] Raw JSON rule validation rejects missing or blank `name`.
- [x] Existing Rule ID required and duplicate checks continue to work.
- [x] Example and test rule payloads include meaningful names.
- [x] Unit tests cover blank-name validation.

## Completion notes

- Added backend rule-name validation in the compiler validation path.
- Made frontend `Rule.name` required and mirrored validation in normal and raw JSON form paths.
- Updated examples and backend/frontend test fixtures to include names.

## Blocked by

None - can start immediately.
