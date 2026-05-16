# PRD: JSON Literal Condition Authoring

Status: done
Created: 2026-05-17

## Problem Statement

Rule authors need to match JSON request fields for HTTP and SPEX traffic, but dynamic JSON fields such as request body fields do not have a fixed schema. A user can type `123`, `true`, or `["a", "b"]`, but the current editor can still save many dynamic JSON child values as plain text. That makes it unclear whether a rule is matching a number, string, boolean, array, object, or null.

The ambiguity is especially painful for SPEX and HTTP JSON bodies because one business interface can contain many field types. Users should not need to choose a separate value type for every field, but they also should not have to guess whether the system will silently convert values behind their back.

## Solution

Treat condition values for dynamic JSON fields as JSON literals. For dynamic JSON fields, the value input is a JSON value editor: `123` saves as a number, `"123"` saves as a string, `true` saves as a boolean, `["a", "b"]` saves as an array, objects save as objects, and `null` saves as null. Invalid JSON values are blocked with direct guidance, especially the common mistake of typing an unquoted string.

Backend equality and membership matching should follow the same strict JSON semantics. Numbers compare as JSON numbers, strings compare only to strings, booleans compare only to booleans, arrays and objects compare structurally, and no fallback string conversion should make `"123"` equal to `123`.

The editor should optionally accept a sample request JSON for the current rule authoring session. For HTTP rulesets the sample represents `request.body`; for SPEX rulesets the sample represents `request.req`; other dynamic JSON roots can use the same helper model. The sample is authoring metadata only. It should provide selectable field paths, inferred field types, and warning messages when the literal value looks inconsistent with the sample.

## User Stories

1. As a MockServer user, I want JSON body condition values to be typed from the value I enter, so that I do not have to choose a separate type control for every field.
2. As a MockServer user, I want `123` to mean number `123`, so that numeric matching is explicit.
3. As a MockServer user, I want `"123"` to mean string `"123"`, so that string matching is explicit.
4. As a MockServer user, I want `true` and `false` to mean booleans, so that boolean matching is explicit.
5. As a MockServer user, I want arrays and objects to be valid condition values, so that list and object comparisons can be expressed without raw rule JSON.
6. As a MockServer user, I want invalid JSON values to show a clear error, so that I can fix unquoted strings and malformed arrays quickly.
7. As a MockServer user, I want a live saved-condition preview, so that I can see the exact JSON rule payload before saving.
8. As a MockServer user, I want backend matching to use the same JSON value semantics as the editor, so that runtime behavior matches what the UI shows.
9. As a MockServer user, I want `"123"` and `123` to be different values, so that type mistakes do not accidentally match.
10. As a MockServer user, I want to paste an optional sample request body, so that the editor can help me find nested fields.
11. As a SPEX rule author, I want the sample request to map to `request.req`, so that cmd plus req authoring feels natural.
12. As an HTTP rule author, I want the sample request to map to `request.body`, so that JSON body matching is natural.
13. As a MockServer user, I want nested object fields to appear as selectable paths, so that I do not have to manually type long paths.
14. As a MockServer user, I want array item fields to appear with `[*]` paths, so that matching any list item is discoverable.
15. As a MockServer user, I want warnings when the literal value type differs from the sample field type, so that mistakes are visible before save.
16. As a MockServer user, I want warnings to remain non-blocking, so that a sample request does not become a hard schema.
17. As a MockServer user, I want the UI to stay compact and readable, so that the new helpers do not crowd the Condition tab.
18. As a MockServer developer, I want JSON literal parsing and sample field inference to be testable outside the Vue component, so that the behavior stays stable.

## Implementation Decisions

- Dynamic JSON condition fields use JSON literal value editing. This applies to protocol fields marked as dynamic JSON roots and their child paths, including HTTP request body, SPEX request req, cache request value, and meta extra.
- Non-dynamic string fields such as method, path, host, header values, query values, cmd, and param keep their existing friendly text/list/regex inputs.
- JSON literal value editing is blocking for invalid JSON when the operator requires a value.
- `exists`, `not_exists`, `is_null`, and `is_not_null` continue to require no value.
- Strict backend equality treats JSON number values as equal across Go numeric representations, but it does not compare numbers to strings.
- Numeric comparison operators only match numeric values. Strings that look numeric do not match numeric comparisons.
- Sample request JSON is stored in the editor form only and is not persisted in the ruleset or sent to runtime APIs.
- Sample request field inference should emit object child paths, array wildcard paths, and scalar type hints.
- Sample request type warnings are advisory. They should not block save or simulation.
- The editor should show a concise sample request helper above the condition tree, not as a separate page.
- Advanced raw JSON editing remains available.

## Testing Decisions

- Backend tests should verify strict JSON equality, strict membership, number comparison behavior, arrays, objects, and string-versus-number mismatches.
- Frontend utility tests should verify JSON literal parsing, invalid JSON errors, sample field inference, wildcard array paths, and warning generation.
- Rule form adapter tests should verify dynamic JSON literals are preserved in the final rule payload.
- Rule authoring helper tests should verify sample-derived field options and warnings appear for HTTP and SPEX dynamic JSON roots.
- Frontend build and type checks should verify the Vue integration.
- UI smoke checks should load the SPEX rules page and verify the rule editor page is still served after the change.

## Out of Scope

- Persisting sample request JSON as a ruleset schema.
- Building a full schema registry or OpenAPI/protobuf import flow.
- Changing selector authoring.
- Changing CEL expression semantics.
- Adding JSONPath filters or advanced query syntax.
- Supporting object keys containing dots in predicate field paths.

## Further Notes

This feature intentionally does not preserve older permissive equality behavior. The project is allowed to break compatibility, and strict JSON semantics make the rule authoring model easier to reason about.

## Completion Notes

- Implemented strict backend JSON equality and numeric comparison behavior.
- Implemented JSON literal value editing for dynamic JSON request fields.
- Implemented optional sample request JSON field inference and advisory type warnings.
- Verification passed: `go test ./...`, `npm test`, `npm run build`, `git diff --check`, and local SPEX rules route smoke.
