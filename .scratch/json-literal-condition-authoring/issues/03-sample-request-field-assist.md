# Sample request field assist

Status: completed
Type: AFK

## Parent

.scratch/json-literal-condition-authoring/PRD.md

## What to build

Add optional sample request JSON support to the condition authoring flow. The sample should generate dynamic JSON field path suggestions and advisory type warnings without persisting schema data into the ruleset.

## Acceptance criteria

- [x] HTTP rules use the sample as `request.body`.
- [x] SPEX rules use the sample as `request.req`.
- [x] Nested fields are available from the field dropdown.
- [x] Array item fields are available as wildcard paths such as `items[*].sku`.
- [x] The editor warns when a JSON literal value type differs from the inferred sample field type.
- [x] Invalid sample JSON is reported without breaking normal condition editing.
- [x] Unit tests cover field inference and warning behavior.

## Completion notes

- Added optional sample request JSON authoring in the Condition tab.
- Sample-derived fields are merged into the predicate field dropdown and warnings remain non-blocking.

## Blocked by

- .scratch/json-literal-condition-authoring/issues/02-json-literal-value-authoring.md
