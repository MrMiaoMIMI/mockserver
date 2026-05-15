# Protocol Response Contract

Status: Completed

## Background

Mockserver has become a multi-protocol decision engine, but the response model is still designed around HTTP. Rule actions, namespace fallback actions, runtime decisions, SDK decisions, and the UI all expose `status`, `headers`, and `body` as first-class response fields. This blocks native responses for protocols such as SPEX, where a response is normally represented as:

```json
{
  "code": 0,
  "resp": "{\"ok\":true}"
}
```

The current namespace model also stores one fallback policy per namespace. Once different protocols share the same namespace, the fallback can only describe an HTTP-style response or a forward target, which is not a valid contract for SPEX, cache, MQ, or future protocols.

This project does not need backward compatibility. The optimal design is to make protocol response payloads part of `ProtocolSpec`, make all response decisions protocol-native, and make namespace fallback policies protocol-scoped.

## Goals

- Move response shape definition into `ProtocolSpec`.
- Replace HTTP-specific mock response fields with a generic protocol response action.
- Support native SPEX response payloads with `code` and `resp`.
- Support HTTP response payloads through the same generic contract, without hardcoded core model fields.
- Make namespace miss policies protocol-scoped so a namespace can serve HTTP and SPEX at the same time.
- Keep mockserver as a protocol-agnostic decision engine. Protocol adapters apply or decode payloads at the edge.
- Update SDK, frontend, docs, and tests to use the new protocol response contract.

## Non-Goals

- No compatibility bridge for old `static_response`, `status`, `headers`, or `body` action payloads.
- No database migration for old JSON data.
- No full implementation for future protocols such as gRPC, MQ, or cache write-back semantics beyond defining protocol-native response contracts where current examples need them.

## Product Requirements

### ProtocolSpec Owns Response Shape

Each registered protocol must declare both request fields/selectors and response payload fields.

Example HTTP response contract:

```json
{
  "status": 200,
  "headers": {
    "content-type": ["application/json"]
  },
  "body": {"ok": true}
}
```

Example SPEX response contract:

```json
{
  "code": 0,
  "resp": "{\"ok\":true}"
}
```

### Generic Rule Action

Rules produce actions with a protocol-native response payload:

```json
{
  "type": "respond",
  "renderer": "static",
  "response": {
    "payload": {
      "code": 0,
      "resp": "{\"ok\":true}"
    }
  }
}
```

Supported renderers:

- `static`: stores a literal response payload.
- `template`: renders a full response payload JSON object.
- `cel`: evaluates to a full response payload object.
- `sequence`: rotates or indexes across response payloads.
- `webhook`: delegates response payload creation to an external webhook.

### Protocol-Scoped Namespace Policies

Namespace config must store miss behavior by protocol:

```json
{
  "id": "default",
  "policies": {
    "http": {
      "ruleset_miss_action": {"type": "forward", "forward": {"timeout_ms": 5000}},
      "rule_miss_action": {"type": "forward", "forward": {"timeout_ms": 5000}}
    },
    "spex": {
      "ruleset_miss_action": {
        "type": "respond",
        "renderer": "static",
        "response": {"payload": {"code": 404, "resp": "{\"message\":\"ruleset miss\"}"}}
      },
      "rule_miss_action": {
        "type": "respond",
        "renderer": "static",
        "response": {"payload": {"code": 404, "resp": "{\"message\":\"rule miss\"}"}}
      }
    }
  }
}
```

### Runtime Decision Shape

Runtime decisions return protocol-native responses:

```json
{
  "kind": "response",
  "matched": true,
  "protocol": "spex",
  "response": {
    "protocol": "spex",
    "payload": {
      "code": 0,
      "resp": "{\"source\":\"mock\"}"
    }
  }
}
```

### SDK Boundary

SDK core stores decisions generically. Protocol adapters decode/apply payloads:

- HTTP adapter applies `status`, `headers`, and `body` to `http.ResponseWriter`.
- SPEX adapter decodes `code` and `resp`.
- Future protocol adapters own their protocol-specific response projection.

## Acceptance Criteria

- HTTP mock rules still produce correct runtime HTTP status, headers, and body through the HTTP adapter.
- SPEX mock rules can produce `{code, resp}` payloads through the generic decision API and SDK SPEX adapter.
- Namespace policies can define different miss actions for HTTP and SPEX under the same namespace.
- `ProtocolSpec` exposes response field metadata for HTTP, SPEX, and cache.
- Backend validation rejects response payloads that do not match the selected protocol response spec.
- Frontend authoring no longer exposes HTTP-only response fields for every protocol; it edits protocol response payload JSON.
- Go backend tests, Go SDK tests, and frontend unit/type checks pass.
