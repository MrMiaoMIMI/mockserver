# 02 - Backend Engine Runtime And Traffic

Status: Completed

## Scope

Update backend compilation, simulation, runtime decisions, HTTP runtime application, namespace miss handling, and traffic indexing to use protocol responses.

## Tasks

- Validate actions against the ruleset protocol response spec.
- Execute static, template, CEL, sequence, and webhook renderers into protocol response payloads.
- Resolve namespace miss policies by `event.protocol`.
- Keep HTTP forward fallback as runtime HTTP debug behavior, returning HTTP response payloads.
- Apply HTTP protocol payloads at the HTTP runtime controller edge.
- Store and index response protocol plus protocol-specific response fields in traffic records.

## Acceptance

- HTTP runtime endpoints apply `response.payload.status`, `response.payload.headers`, and `response.payload.body`.
- SDK decision API returns SPEX `{code, resp}` without coercing it into HTTP fields.
- Namespace miss behavior can differ for HTTP and SPEX in the same namespace.
- Backend Go tests cover HTTP response, SPEX response decision, and namespace protocol policy behavior.
