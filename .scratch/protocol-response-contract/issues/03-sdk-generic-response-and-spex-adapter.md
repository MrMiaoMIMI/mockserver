# 03 - SDK Generic Response And SPEX Adapter

Status: Completed

## Scope

Refactor `mocksdk` decision types and protocol adapters to consume generic protocol responses.

## Tasks

- Change SDK `ResponseDecision` to carry `protocol` and `payload`.
- Update HTTP adapter to parse and apply HTTP payload fields.
- Add SPEX adapter helper for decoding `{code, resp}`.
- Update SDK tests and README examples.

## Acceptance

- HTTP adapter still writes status, headers, and body correctly.
- SPEX adapter decodes `code` and `resp` from a generic response decision.
- SDK tests pass under the module Go version.
