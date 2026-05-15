# 04 - Frontend Protocol Response Authoring

Status: Completed

## Scope

Update frontend types, rule authoring, namespace policy editing, summaries, diagnostics, and tests to use protocol response payloads.

## Tasks

- Update TypeScript API models for protocol response specs, actions, decisions, and namespace policies.
- Replace HTTP-only rule action fields with protocol response payload JSON editing.
- Update namespace fallback editing to configure per-protocol policies.
- Update summaries and diagnostics to display protocol-native response labels.
- Update frontend unit tests.

## Acceptance

- Rule authoring works for HTTP and SPEX payload JSON.
- Namespace editor can represent separate HTTP and SPEX policies.
- Frontend type check and unit tests pass.
