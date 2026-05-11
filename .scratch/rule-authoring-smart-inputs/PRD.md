# PRD: Rule Authoring Smart Inputs

Status: completed
Created: 2026-05-09

## Problem Statement

MockServer now supports multiple protocols through code-registered `ProtocolSpec` definitions, and the frontend can render protocol fields and operators. However, users still need to understand low-level rule payload details when configuring conditions and selectors. They must know when `value` should be a string, number, boolean, array, JSON object, or omitted entirely. They also need to understand field-path conventions such as `request.headers.x-env[0]`, `request.query.tag[*]`, `request.body.status`, and `request.value.user.id`.

This makes protocol-generalized rule authoring technically correct but still hard to use. The frontend should guide users as they select a field and operator, so the value editor, examples, helper copy, and resulting payload shape change automatically.

## Solution

Introduce a protocol-aware smart input layer for condition and selector authoring. The UI should derive an input specification from the selected protocol field and operator, then render the correct value editor:

- no value input for existence and null checks
- text input for string comparison and string matching
- numeric input for numeric comparison
- boolean choices for boolean equality
- list/chip input for `in` and `not_in`
- regex input with immediate regex validation
- JSON textarea/editor for object, root JSON, and array comparisons

Dynamic protocol fields should become easier to fill. Instead of requiring users to manually type full paths, the editor should support field root selection plus a focused suffix/key builder:

- HTTP query and header fields expose a key input and first/any value mode.
- HTTP body JSON fields expose a body path input.
- Cache value JSON fields expose a value path input.

The editor should also show concise helper text, example chips, and a payload preview so users can understand exactly what will be sent to MockServer without memorizing the rule JSON shape.

## User Stories

1. As a rule author, I want the value input to change when I choose a different operator, so that I do not need to know the JSON payload shape.
2. As a rule author, I want `exists`, `not_exists`, `is_null`, and `is_not_null` to hide the value input, so that I do not accidentally fill meaningless values.
3. As a rule author, I want `in` and `not_in` to use a list editor, so that I can build arrays without writing JSON by hand.
4. As a rule author, I want numeric operators to use numeric input, so that `123` is saved as a number instead of the string `"123"`.
5. As a rule author, I want boolean fields to use true/false choices, so that I do not mistype boolean values.
6. As a rule author, I want regex values to be validated immediately, so that invalid regex rules are caught before saving.
7. As a rule author, I want JSON object and array fields to use JSON input, so that complex values remain explicit and readable.
8. As an HTTP user, I want request method fields to offer method examples, so that common method rules are fast to create.
9. As an HTTP user, I want header fields to expose a header key input, so that I do not need to write `request.headers.<key>[0]` manually.
10. As an HTTP user, I want query fields to expose a query key input, so that query rules are easier to create.
11. As an HTTP user, I want query and header builders to choose first value or any value, so that I can control `[0]` versus `[*]` matching.
12. As an HTTP user, I want body JSON fields to expose a body path input, so that `request.body.status` can be created from `status`.
13. As a cache user, I want cache value fields to expose a value path input, so that `request.value.user.id` can be created from `user.id`.
14. As a cache user, I want operation and key examples, so that get/set/del and key-prefix rules are quick to configure.
15. As a selector author, I want selectors to use the same smart value behavior as rule conditions, so that settings and rules feel consistent.
16. As a selector author, I want selector rows to show examples and payload preview, so that I can understand the resulting selector.
17. As a frontend user, I want helper text to explain the selected operator, so that I know how the matcher will interpret my value.
18. As a frontend user, I want examples to be clickable, so that common values can be inserted quickly.
19. As a frontend user, I want the generated condition JSON preview to be visible, so that advanced users can verify exact payloads.
20. As a maintainer, I want the input-selection logic isolated in a deep utility module, so that future protocols can reuse it without duplicating UI code.
21. As a maintainer, I want smart input behavior covered by unit tests, so that value serialization does not regress.
22. As a maintainer, I want browser QA on desktop and narrow viewport, so that the denser authoring UI remains usable.

## Implementation Decisions

- Build a protocol-aware value input specification module that derives editor type, requiredness, helper text, placeholders, examples, and value normalization from field metadata and operator.
- Keep the backend API and `ProtocolSpec` schema unchanged in this slice; derive first-version authoring hints from existing field paths, field types, dynamic-path flags, and operators.
- Introduce a reusable smart value input component for text, number, boolean, list, regex, JSON, and no-value editors.
- Introduce dynamic path builder behavior inside condition authoring so dynamic roots can produce concrete paths without forcing manual full-path entry.
- Reuse the same smart value input component for selector rows.
- Preserve existing presets as shortcuts, but make the selected protocol field and operator drive the value editor.
- Show a payload preview for predicate conditions and selector rows.
- Do not add new routes or backend endpoints.
- Do not add database schema changes.

## Testing Decisions

- Good tests should assert externally visible authoring behavior: selected editor kind, value serialization, dynamic path generation, helper/example availability, and resulting condition/selector payloads.
- Unit tests should cover the value input specification module because it is the deep, reusable module for the feature.
- Existing rule authoring tests should be extended to prove smart defaults for HTTP and cache conditions.
- Existing diagnostics and form adapter tests remain useful prior art for protocol-aware frontend helpers.
- Final verification should run `npm test`, `npm run type-check`, `npm run build`, and `go test ./...`.
- Browser QA should cover at least creating a ruleset selector row and inspecting a condition editor at desktop and narrow viewport widths.

## Out of Scope

- Backend changes to `ProtocolSpec` authoring hints.
- Dynamic user-configured protocol specs.
- SPEX, gRPC, or MQ protocol implementation.
- Full CEL expression builder.
- Full JSONPath autocomplete from real captured traffic.
- Persisted user-specific field favorites.
- A full redesign of the application shell.

## Further Notes

- This feature improves usability on top of the existing protocol-generalized architecture.
- Future protocol-specific hints can later move into `ProtocolSpec` once SPEX/gRPC/MQ reveal which hints belong in the backend contract.
- The first version should prioritize correct value typing and clear guidance over advanced autocomplete.
