# Issue 01: Backend SPEX ResponseSpec Uses Direct JSON Resp

Status: Completed

## Scope

Update SPEX protocol response metadata so `resp` is a direct JSON field, not a JSON-encoded string.

## Changes

- Changed SPEX default action response from `resp: "{}"` to `resp: {}`.
- Changed SPEX `ProtocolSpec.response.fields[].resp` type from `string` with `json_string` format to `json`.
- Updated backend tests that construct and assert SPEX responses.

## Verification

- `go test ./mockprotocol ./internal/controller`
- `go test ./...`
