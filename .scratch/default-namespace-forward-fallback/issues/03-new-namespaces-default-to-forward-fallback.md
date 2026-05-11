# New namespaces default to forward fallback

Status: done
Type: AFK

## Parent

.scratch/default-namespace-forward-fallback/PRD.md

## What to build

When a user creates a new namespace, both ruleset miss and rule miss should default to forward fallback. This default should be consistent across the admin API and frontend namespace creation flow, while preserving the ability to explicitly choose response fallback.

The completed slice should make newly created namespaces pass-through by default, prevent the UI from accidentally saving the old response defaults, and keep existing response fallback configuration editable.

## Acceptance criteria

- [ ] Creating a namespace without explicit fallback actions, if supported by the API contract, stores forward fallback for both ruleset miss and rule miss.
- [ ] Creating a namespace from the frontend starts both fallback editors in forward mode.
- [ ] The namespace list and namespace detail responses show forward fallback for newly created namespaces.
- [ ] A newly created namespace can still be changed to explicit response fallback for either miss reason.
- [ ] Frontend ruleset namespace options no longer synthesize old response fallback defaults for typed namespace values.
- [ ] Focused backend tests and frontend build or type checks validate the new default behavior.

## Blocked by

- .scratch/default-namespace-forward-fallback/issues/01-default-namespace-ruleset-miss-forwards.md

## Comments
