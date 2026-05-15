# PRD: Protocol-Driven Mock Response Editor

Status: Completed

## Background

The rule action editor previously exposed a single `Response Payload JSON` textarea. Users had to know the internal action payload shape for each protocol, such as HTTP `status`/`headers`/`body` or SPEX `code`/`resp`, which made rule authoring error-prone and protocol-specific.

The desired direction is to let `ProtocolSpec.response` declare the top-level response shape and let the frontend render first-class controls from that declaration. HTTP response `body` and SPEX `resp` should remain raw JSON editor fields; their internal business fields are user-owned and should not be inferred from protocol metadata.

Backward compatibility is intentionally out of scope.

## Goals

- Replace the raw action response textarea with a protocol-driven response editor.
- Keep the editor protocol-agnostic: frontend behavior comes from `ProtocolSpec.response.fields`.
- Render top-level protocol response fields with suitable controls:
  - number fields as numeric inputs
  - bool fields as switches
  - string fields as text inputs
  - header objects as editable key/value rows
  - JSON/object/array fields as JSON editors
- Model HTTP body and SPEX resp directly as JSON values, not JSON-encoded strings.
- Reuse the same response editor for static actions, sequence actions, and namespace fallback policies.

## Non-Goals

- No introspection of HTTP body or SPEX resp internal fields.
- No migration layer for old saved rules that stored SPEX `resp` as a JSON string.
- No per-protocol hardcoded frontend response schema.

## Requirements

1. Backend `ProtocolSpec` must declare SPEX `resp` as a JSON field with an object default.
2. Frontend form state must store protocol response payloads as structured objects.
3. Frontend must maintain JSON field drafts separately so invalid JSON can be surfaced before save.
4. Save/build logic must validate and serialize according to `ProtocolSpec.response.fields`.
5. Rule editor and namespace fallback editor must use the same response editor component.
6. Verification must cover Go tests, frontend tests, type-check, build, and browser QA for SPEX editing.

## Acceptance Criteria

- Editing a SPEX rule action shows `code` plus a `resp JSON` editor, not a generic `Response Payload JSON` textarea.
- SPEX saved action payload uses direct JSON: `{ "code": 0, "resp": { ... } }`.
- HTTP body remains a JSON editor and does not expose internal body fields.
- Header editing is key/value based instead of requiring manual whole-object JSON edits.
- Static action, sequence action, and fallback policy response editing share the same behavior.
- Existing automated test suites pass.
- Browser QA confirms the new SPEX action editor renders against the updated protocol spec.
