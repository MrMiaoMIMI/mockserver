# Name-first Rule Manager display

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-name-first-identity/PRD.md

## What to build

Update Rule Manager and rule display helpers so cards and selected-rule context present Rule Name first and Rule ID as secondary technical metadata.

## Acceptance criteria

- [x] Rule Manager card title uses Rule Name as the strongest text.
- [x] Rule ID is still visible as secondary metadata.
- [x] Search continues to match Rule Name and Rule ID.
- [x] Selected-rule summary uses Rule Name where space allows.
- [x] Old unnamed data displays a safe fallback without making the ID the primary user-facing label.

## Completion notes

- Added centralized rule display helpers for name-first and ID metadata labels.
- Updated Rule Manager selected summary and cards to use Rule Name first.
- Kept existing search coverage over Rule Name and Rule ID.

## Blocked by

- .scratch/rule-name-first-identity/issues/01-required-rule-name-contract.md
