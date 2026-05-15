# 01 - Protocol Response Spec And Core Action

Status: Completed

## Scope

Refactor the protocol and BO model layer so response payloads are protocol-native and no longer HTTP-shaped.

## Tasks

- Extend `mockprotocol.ProtocolSpec` with response payload metadata and defaults.
- Add response specs for HTTP, SPEX, and cache.
- Replace rule action response fields with `type`, `renderer`, and `ProtocolResponse`.
- Replace namespace top-level fallback fields with `policies[protocol]`.
- Replace runtime/simulation response models with `ProtocolResponse`.
- Replace old action constants with `respond`, `forward`, and renderer constants.

## Acceptance

- Core models no longer expose `Action.Status`, `Action.Headers`, `Action.Body`, `NamespaceResponseFallback`, or HTTP-only `ActionExecution`.
- `ProtocolSpec` can validate and default HTTP and SPEX response payloads.
- Unit tests cover response spec validation for HTTP and SPEX.
