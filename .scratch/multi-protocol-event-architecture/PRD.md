# PRD: Multi-Protocol Event Architecture

Status: completed
Created: 2026-05-09

## Problem Statement

MockServer is intended to become a generic mock tool rather than an HTTP-only runtime proxy. The current implementation already has useful protocol concepts such as ruleset protocol, event protocol, namespace fallback, and decision responses. However, the request model, selector model, SDK normalizer, frontend field helpers, and HTTP runtime route still mostly assume HTTP request shape.

This makes future protocol support risky. RPC protocols such as SPEX or gRPC, cache operations, and MQ traffic all need different original-request capture, metadata, payload, response application, and forwarding behavior. If MockServer keeps expanding the current HTTP request model, the generic rule engine will accumulate protocol-specific fields and every new protocol will require edits across matching, validation, SDK conversion, frontend fields, and runtime forwarding.

The user wants a design where MockServer remains a generic decision engine, protocol-specific request conversion is explicit and testable, and mockinject remains responsible for forwarding or executing original protocol calls.

## Solution

Introduce a code-registered protocol contract and migrate events to a protocol-neutral request document.

MockServer should evaluate rules against:

```go
type Event struct {
    Protocol  string
    Namespace string
    Request   map[string]any
    Meta      EventMeta
}
```

Each protocol registers a `ProtocolSpec` in code. The spec defines stable rule field roots, selectors, dynamic-path behavior, field types, and optional field-level operator overrides. Users do not dynamically define protocol specs in the database. Users continue to configure and persist rulesets, rules, namespaces, and fallback policy.

Mock SDK provides protocol normalizers that convert original protocol calls into Events. It does not fetch protocol specs at runtime to perform conversion, because conversion requires concrete protocol SDK knowledge. `ProtocolSpec` defines what normalized fields exist; SDK normalizers implement how to produce them.

The first non-HTTP tracer bullet is cache. Cache support proves that the engine can match a protocol that does not use HTTP host/path/method as its core shape.

## User Stories

1. As a MockServer user, I want one ruleset model for HTTP and non-HTTP protocols, so that I can reuse the same mental model across mock scenarios.
2. As a MockServer user, I want protocol-specific field paths to be discoverable, so that I can create valid rules without memorizing hidden field names.
3. As a MockServer user, I want HTTP JSON bodies to be matched with paths like `request.body.status`, so that I can write readable payload-based rules.
4. As a MockServer user, I want raw HTTP body matching through `request.raw_body`, so that I can match non-JSON payloads or debug original text.
5. As a MockServer user, I want cache rules to match `request.operation`, `request.key`, `request.ttl_ms`, and `request.value`, so that cache operations can be mocked without HTTP concepts.
6. As a MockServer user, I want `is_null` to mean missing or null, so that absent values and JSON null values can be treated as the same no-value state.
7. As a MockServer user, I want `exists` to mean the path exists even if its value is null, so that I can distinguish presence from value.
8. As a MockServer user, I want dynamic JSON fields to support all first-version operators, so that the first version does not block useful payload matching.
9. As a mockinject developer, I want mocksdk HTTP normalization to output the new request document, so that injected HTTP code matches the same field paths that the UI exposes.
10. As a mockinject developer, I want mocksdk cache normalization to output cache Events, so that a cache mockinject prototype can call MockServer without inventing its own event schema.
11. As a mockinject developer, I want MockServer to return response or forward decisions only, so that forwarding remains inside the protocol-specific injected code.
12. As a frontend user, I want the rule editor to receive protocol field and selector metadata from MockServer, so that field dropdowns can become protocol-aware.
13. As a frontend user, I want selectors to be protocol-defined, so that HTTP uses path/host-like selectors while cache can use operation/key-like selectors.
14. As a maintainer, I want protocol specs code-registered and versionable, so that MockServer, frontend metadata, and SDK normalizers do not drift.
15. As a maintainer, I want protocol specs to avoid labels and advanced flags in the first version, so that the contract remains small.
16. As a maintainer, I want field-level operators to default from field type, so that specs only override operators when a field needs special constraints.
17. As a maintainer, I want HTTP old event fields migrated directly, so that the greenfield project converges on the new shape without compatibility adapters.
18. As a maintainer, I want matcher indexing to remain an implementation detail, so that the first `ProtocolSpec` contract does not expose premature `IndexHints`.
19. As a maintainer, I want tests proving HTTP and cache decisions through the same endpoint, so that HTTP is no longer the only practical protocol.
20. As a maintainer, I want documentation to explain ProtocolSpec, SDK normalizers, and mockinject ownership, so that contributors add future protocols at the right layer.

## Implementation Decisions

- Define `ProtocolSpec` as a code-registered protocol contract, not a user-configured DB record.
- First-version `FieldSpec` includes only `path`, `type`, `dynamic_path`, and optional `operators`.
- First-version `SelectorSpec` is included because selectors are protocol-specific field choices that the frontend and backend must share.
- Do not include `label` or `advanced` in the first field contract.
- Do not include `IndexHints` in the first public ProtocolSpec. Indexing should remain an internal matcher optimization until a second protocol proves which hints are stable enough to expose.
- If a field does not define `operators`, derive operators from its field type.
- Add `is_null` and `is_not_null`.
- Define `exists` as path exists, even if the value is null.
- Define `not_exists` as path does not exist.
- Define `is_null` as path does not exist, or path exists but its value is null or Go nil.
- Define `is_not_null` as path exists and its value is not null or Go nil.
- JSON dynamic subfields may use all first-version operators.
- HTTP request projection sets `request.raw_body` whenever a body exists.
- HTTP request projection also sets `request.body` only when the body parses as JSON.
- Empty HTTP bodies set neither `request.body` nor `request.raw_body`.
- HTTP spec fields are `request.method`, `request.host`, `request.path`, `request.query`, `request.headers`, `request.body`, and `request.raw_body`.
- Cache spec fields are `request.operation`, `request.key`, `request.ttl_ms`, and `request.value`.
- HTTP selector fields are `request.host` and `request.path`.
- Cache selector fields are `request.operation` and `request.key`.
- Backend validation should use ProtocolSpec to validate field roots, dynamic paths, and operators.
- The engine should resolve fields against the new `request` document rather than a fixed HTTP request struct.
- Mock SDK should expose generic `Decide` plus protocol normalizers such as HTTP and cache event builders.
- Mockinject remains responsible for applying protocol response decisions and forwarding original requests.

## Testing Decisions

- Good tests should assert public behavior and stable contracts, not internal helper implementation details.
- ProtocolSpec tests should cover default operators, field-level operator overrides, selectors, and validation of dynamic paths.
- Path resolver tests should cover `exists`, `not_exists`, `is_null`, and `is_not_null`.
- HTTP normalizer tests should cover JSON body, raw body, non-JSON body, and empty body under the new request document.
- Backend decision tests should prove existing HTTP runtime and SDK decision behavior still work after event migration.
- Cache tracer bullet tests should prove a cache ruleset can be created, published, and matched through the SDK decision endpoint.
- SDK tests should prove cache Events can be built and passed through the decision client.
- Frontend type or build verification should prove the TypeScript event shape accepts dynamic request fields.
- Final verification should run `go test ./...` and a frontend build if frontend type changes require it.

## Out of Scope

- Full SPEX implementation.
- Full gRPC implementation.
- Full MQ implementation.
- Real cache client forwarding inside MockServer or Mock SDK.
- Dynamic user-defined ProtocolSpec stored in the database.
- Runtime proxy routes for non-HTTP protocols.
- Protocol plugin loading.
- A full frontend redesign for protocol-specific rule authoring.
- Compatibility adapters for the old HTTP event request shape.

## Further Notes

- This PRD follows the earlier decision-only boundary: MockServer and Mock SDK return decisions; mockinject performs protocol-specific execution.
- Cache is intentionally selected as the first non-HTTP tracer bullet because it has a small request shape and clearly demonstrates operation/key/value matching without HTTP host/path semantics.
- `IndexHints` should stay internal for now. The engine can still build fast indexes for selected fields, but the public protocol contract should not expose optimization knobs until the system has more protocol evidence.
