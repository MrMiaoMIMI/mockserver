# PRD: Protocol End-to-End Adaptation

Status: completed
Created: 2026-05-09

## Problem Statement

MockServer has a protocol-neutral event foundation, but the end-to-end product experience is not fully protocol-neutral yet. The backend matcher accepts a generic request document and exposes a protocol catalog, while the remaining authoring surface still carries HTTP assumptions:

- ruleset selectors are modeled as HTTP `hosts` and `path_prefixes`
- frontend selector editing and list/detail summaries render only HTTP selector fields
- frontend condition presets still behave as the primary field source instead of consuming `ProtocolSpec`
- simulation defaults are optimized for HTTP and do not create useful cache events
- Mock SDK has generic decisions and cache event support, but the protocol adapter shape is not explicit enough for future SPEX, gRPC, cache, and MQ adapters

The user wants MockServer BE, MockServer FE, and Mock SDK to all complete the protocol-generalization work. The immediate goal is not to add every future protocol, but to make HTTP and cache both first-class end-to-end examples so future protocols follow the same pattern.

## Solution

Complete the protocol-generalization vertical slice by making `ProtocolSpec` the shared authoring contract across backend, frontend, and SDK docs.

Backend ruleset selectors should become protocol field predicates instead of HTTP-specific `hosts/path_prefixes`. A selector should be expressed as protocol-approved field conditions, for example:

```json
{
  "selector": {
    "all": [
      {"field": "request.operation", "op": "eq", "value": "get"},
      {"field": "request.key", "op": "prefix", "value": "user:"}
    ]
  }
}
```

HTTP uses the same shape:

```json
{
  "selector": {
    "all": [
      {"field": "request.host", "op": "eq", "value": "demo.com"},
      {"field": "request.path", "op": "prefix", "value": "/api/"}
    ]
  }
}
```

The protocol catalog should return effective field and selector operators so the frontend does not duplicate backend operator derivation. The frontend should load protocol specs, render protocol choices, selector rows, condition fields, and simulation defaults from the selected protocol. HTTP remains supported as a first-class protocol; cache becomes the first complete non-HTTP authoring and SDK decision path.

Mock SDK should make the core/adapter separation explicit: core decision client and event/decision types stay generic, while HTTP and cache adapters own protocol-specific event construction and response application helpers.

## User Stories

1. As a MockServer user, I want selectors to use protocol fields, so that non-HTTP protocols do not depend on HTTP host or path concepts.
2. As a MockServer user, I want HTTP selectors to continue supporting host and path-prefix matching, so that existing HTTP mock workflows remain ergonomic.
3. As a MockServer user, I want cache selectors to support operation and key matching, so that cache rulesets can be narrowed before rule evaluation.
4. As a MockServer user, I want selector matching to appear in simulation explain output, so that I can understand why a ruleset was or was not selected.
5. As a MockServer user, I want selector validation to reject fields not declared as selectors by the protocol, so that rulesets fail early instead of behaving unexpectedly.
6. As a MockServer user, I want selector validation to reject unsupported operators for selector fields, so that the UI and backend remain consistent.
7. As a frontend user, I want to choose a protocol from MockServer's protocol catalog, so that the rule editor reflects what the backend actually supports.
8. As a frontend user, I want selector field dropdowns to come from the selected protocol, so that HTTP and cache both have correct selector options.
9. As a frontend user, I want condition field dropdowns to come from the selected protocol, so that I do not need to memorize protocol-specific paths.
10. As a frontend user, I want dynamic fields to allow path suffix entry, so that I can write rules such as `request.body.status` or `request.value.user.id`.
11. As a frontend user, I want operator dropdowns to follow the selected field, so that I do not configure operators the backend will reject.
12. As a frontend user, I want HTTP simulation defaults to continue using method, host, path, headers, query, and body, so that HTTP debugging remains fast.
13. As a frontend user, I want cache simulation defaults to include operation, key, ttl, and value, so that cache rules can be tested from the UI.
14. As a frontend user, I want rule list and detail pages to summarize generic selectors, so that HTTP and cache rulesets are both readable.
15. As a mockinject developer, I want Mock SDK core to stay protocol-neutral, so that future protocol adapters can share the same decision client.
16. As a mockinject developer, I want HTTP adapter helpers to own HTTP request normalization and response application, so that HTTP-specific behavior is not confused with the SDK core.
17. As a mockinject developer, I want cache adapter helpers to own cache event construction, so that cache mockinject code has a stable integration point.
18. As a maintainer, I want `ProtocolSpec` to return effective operators, so that frontend code does not duplicate backend operator logic.
19. As a maintainer, I want HTTP and cache end-to-end tests, so that both protocols prove the same authoring and decision model.
20. As a maintainer, I want documentation to state that HTTP runtime is a debug convenience and SDK decision is the generic path, so that future protocol work is placed correctly.

## Implementation Decisions

- Replace the HTTP-specific selector shape with a generic selector condition list.
- Selector predicates should use only fields declared in `ProtocolSpec.Selectors`.
- Selector operator validation should use the effective operators derived from `ProtocolSpec`.
- Empty selectors remain valid and match all rulesets with the same protocol and namespace.
- Selector explain output should continue to use `selector_checks`, but each check should identify the protocol field path.
- Selector specificity should be generic: exact matches are more specific than prefixes, longer prefix values are more specific than shorter ones, and more matched selector predicates increase specificity.
- `/mockserver/api/v1/admin/protocols` should return effective operators for fields and selectors.
- The frontend store should cache the protocol catalog and expose helpers for finding specs, fields, selectors, and operators.
- The ruleset settings dialog should render protocol-driven selector rows rather than hardcoded hosts/path prefixes.
- The condition editor should use protocol fields as its primary field source and keep presets only as shortcuts.
- Simulation default event construction should be protocol-aware.
- Rule list and rule detail selector summaries should render generic selector predicates.
- Mock SDK core should retain generic `Client`, `Event`, `Decision`, and error handling.
- HTTP adapter helpers should own `NormalizeHTTPRequest`, `Decide`, and `ApplyResponse`.
- Cache adapter helpers should own `Event` and cache request projection.
- Existing HTTP runtime server-side forwarding remains only for HTTP runtime debugging.
- Non-HTTP protocols should use the SDK decision endpoint and mockinject-owned forwarding.

## Testing Decisions

- Good tests should assert externally visible behavior: API validation, selector matching, simulation explain output, frontend view model output, and SDK adapter behavior.
- Backend tests should cover HTTP selector migration, cache selector matching, invalid selector fields, invalid selector operators, and protocol catalog effective operators.
- Frontend tests should cover protocol catalog helpers, generic selector summaries, protocol-aware simulation defaults, and condition/operator options.
- SDK tests should cover HTTP and cache adapter construction without requiring server-side forwarding.
- Existing controller end-to-end tests should be updated to prove HTTP and cache rulesets work through the new selector shape.
- Final verification should run Go tests and frontend build.

## Out of Scope

- Full SPEX adapter implementation.
- Full gRPC adapter implementation.
- Full MQ adapter implementation.
- Runtime proxy routes for non-HTTP protocols.
- Dynamic user-defined protocol specs stored in the database.
- Protocol plugin loading.
- A full visual redesign of the frontend shell.
- Database migrations for selector compatibility with old persisted JSON.

## Further Notes

- This PRD intentionally treats cache as the complete non-HTTP tracer bullet before adding heavier RPC protocols.
- The project is greenfield, so old HTTP selector compatibility adapters are not required.
- Future protocol additions should register a protocol spec, add a protocol adapter normalizer, and gain frontend authoring support through the same `ProtocolSpec` contract.
