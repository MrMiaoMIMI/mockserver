# Mock SDK Quickstart

Mock SDK is the runtime decision layer for generated or injected code. It is intended for `mockinject` implementations that intercept an original HTTP request, ask MockServer for a decision, and then either return a mock response or continue with protocol-specific forwarding outside the SDK.

The SDK is maintained as a separate Go module:

```text
github.com/MrMiaoMIMI/mocksdk
```

It supports Go 1.17+ and uses only the Go standard library so injected business repositories do not inherit MockServer's higher Go toolchain requirement.

## Tool Chain Roles

- `mockserver` manages namespaces, draft rulesets, published snapshots, matching, fallback policy, and the SDK decision endpoint.
- `mocksdk` converts protocol-specific calls into MockServer events, calls the SDK decision endpoint, and exposes typed decisions plus protocol payload decoders. It does not execute upstream forwarding.
- `mockinject` is external to this repository. It injects code before business compilation, calls `mocksdk`, and owns protocol-specific execution such as forwarding the original request.

## Environment Variables

Environment variables have higher priority than explicit SDK options.

| Variable | Meaning |
| --- | --- |
| `MOCKSERVER_HOST` | Base URL of MockServer, for example `http://127.0.0.1:8080`. It may include the deployment prefix, for example `http://127.0.0.1:8080/tenant-a`. Overrides `Config.MockServerURL`. |
| `MOCKSERVER_API_PREFIX` | Optional MockServer deployment path prefix, for example `/tenant-a`. Overrides `Config.APIPrefix`. The SDK appends it before `/mockserver/api/v1/sdk/decision` unless `MOCKSERVER_HOST` already ends with the same prefix. |
| `MOCKSERVER_NAMESPACE_ID` | Namespace used by adapter helpers when the caller does not pass a namespace. Overrides `Config.Namespace`. |
| `MOCKSERVER_TIMEOUT_MS` | Optional decision-call timeout in milliseconds. Overrides `Config.Timeout` when set to a positive integer. |

If neither `MOCKSERVER_NAMESPACE_ID` nor `Config.Namespace` is set, the SDK uses `default`.

## Scenario Context

Short-lived agent scenarios are passed through request context instead of
process-wide environment variables:

```go
ctx = mocksdk.WithScenarioID(ctx, "scn_checkout_timeout")
```

The SDK validates scenario ids before sending the decision request. Valid ids
start with `scn_`, are 8 to 128 characters long, and only contain letters,
digits, underscores, and hyphens.

If an event already contains `event.Meta.ScenarioID`, that explicit value wins
over the context value. This lets mockinject decide how to propagate the
scenario id from each test request without constraining the whole business
process to one scenario.

## Decision Endpoint

The SDK calls:

```text
POST /mockserver/api/v1/sdk/decision
```

When the MockServer backend is configured with `server.api_prefix: /tenant-a`
or `MOCKSERVER_API_PREFIX=/tenant-a`, configure the SDK with either
`MockServerURL: "http://127.0.0.1:8080/tenant-a"` or
`MockServerURL: "http://127.0.0.1:8080", APIPrefix: "/tenant-a"`. The final
request becomes:

```text
POST /tenant-a/mockserver/api/v1/sdk/decision
```

By default this endpoint evaluates published rulesets. When the event carries
`meta.scenario_id`, MockServer first evaluates active scenario overlay rules for
that scenario. While `meta.scenario_id` is present, a missing, inactive, expired,
or unmatched scenario resolves through the event namespace fallback policy and
does not fall through to published rulesets. Draft rules do not affect SDK
decisions until they are published.
Rulesets must declare at least one selector condition, so unrelated SDK traffic that does not enter a declared ruleset boundary becomes `ruleset_miss`.

The endpoint returns MockServer's standard response envelope. The payload is `data.decision`.
MockServer records raw SDK traffic for matched decisions, `rule_miss`, and decision errors. `ruleset_miss` is counted in runtime metrics but is not persisted as raw traffic.

### Response Decision

`kind=response` means injected code should return the decision response to the business process.

This can happen when:

- A published rule matches and produces a mock response.
- A namespace fallback policy is configured as `respond`.

Important fields:

- `matched`: true for a rule hit, false for a fallback response.
- `fallback`: true when the response came from namespace fallback.
- `trace.ruleset_id`: matched ruleset ID when available.
- `trace.rule_id`: matched rule ID when available.
- `trace.snapshot_id`: matched published snapshot ID when available.
- `trace.fallback_reason`: `ruleset_miss` or `rule_miss` for fallback responses.
- `protocol`: selected protocol for this decision.
- `response.protocol`: protocol of the response payload.
- `response.payload`: protocol-native response payload. The SDK stores it as one raw JSON payload and protocol adapters decode it into typed payloads. HTTP adapters decode `status`, `headers`, and `body`; SPEX adapters decode `code` and raw JSON `resp`; cache adapters decode `hit` and `value`.
- `meta.trace_id`: trace ID propagated from the original request event.
- `meta.scenario_id`: scenario id used for the decision when the request carried one.

### Forward Decision

`kind=forward` means injected code should forward the original request itself.

This happens when a ruleset miss or rule miss resolves to namespace fallback policy `forward`.

Important fields:

- `fallback`: true.
- `trace.fallback_reason`: `ruleset_miss` or `rule_miss`.
- `forward.timeout_ms`: effective forward timeout policy from MockServer.
- `meta.trace_id`: trace ID propagated from the original request event.

The SDK decision endpoint does not forward upstream traffic. It only returns the decision. Direct MockServer HTTP runtime URLs may still execute server-side forward fallback, but that path is a runtime debugging convenience for HTTP rules, not the generic forwarding architecture.

## Minimal Example

```go
client, err := mocksdk.NewClient(mocksdk.Config{
    MockServerURL: "http://127.0.0.1:8080",
    APIPrefix:     "/tenant-a",
    Namespace:     "default",
})
if err != nil {
    return err
}

decision, err := httpadapter.Decide(ctx, client, originalRequest, "")
if err != nil {
    return err
}

switch decision.Kind {
case mocksdk.DecisionKindResponse:
    return httpadapter.ApplyResponse(responseWriter, decision)
case mocksdk.DecisionKindForward:
    // mockinject owns protocol-specific forwarding. For HTTP, use the
    // injected application's own client/transport path here.
    return continueOriginalRequest(ctx, originalRequest, decision)
default:
    return fmt.Errorf("unsupported mock decision kind %q", decision.Kind)
}
```

If mockinject needs the decoded response payload instead of directly applying
it, use the protocol adapter helper:

```go
payload, err := httpadapter.PayloadFromDecision(decision)
if err != nil {
    return err
}
status := payload.Status
headers := payload.Headers
body := payload.Body
```

The same pattern is available for SPEX and cache:

```go
spexPayload, err := spexadapter.PayloadFromDecision(decision)
cachePayload, err := cacheadapter.PayloadFromDecision(decision)
```

## Event Projection

Events use a protocol-neutral request document:

```go
type Event struct {
    Protocol  string
    Operation string
    Namespace string
    Request   map[string]interface{}
    Meta      EventMeta
}
```

Protocol field names come from code-registered `ProtocolSpec` definitions in MockServer. The SDK does not fetch those specs to discover how to convert original requests. Each protocol normalizer is code because it must understand the concrete source API, such as `http.Request` or a cache client call.

HTTP normalization provides fields such as `request.method`, `request.host`, `request.path`, `request.query`, `request.headers`, `request.body`, and `request.raw_body`. JSON bodies set both `request.body` and `request.raw_body`; non-JSON bodies set only `request.raw_body`; empty bodies set neither.

Cache events can be built with:

```go
event := cacheadapter.Event("default", cacheadapter.Request{
    Operation: "get",
    Key:       "user:123",
})

decision, err := client.Decide(ctx, event)
```

Cache projection provides `request.operation`, `request.key`, optional `request.ttl_ms`, and optional `request.value`.

SPEX normalization is available from `github.com/MrMiaoMIMI/mocksdk/spexadapter`:

```go
event := spexadapter.Event("default", spexadapter.Request{
    Cmd: "shop.GetOrder",
    Req: map[string]interface{}{
        "order_id": "1001",
    },
    Param: "region=sg",
})

decision, err := client.Decide(ctx, event)
```

SPEX projection provides `request.cmd`, optional JSON `request.req`, and optional `request.param`.

## Errors

SDK errors are classified with `mocksdk.ErrorKind`:

- `config`: invalid SDK configuration, invalid namespace, invalid URL, or invalid request construction.
- `transport`: network failure, context cancellation, timeout, or response body read failure.
- `server`: non-2xx HTTP status or non-zero MockServer response envelope code.
- `decode`: malformed response envelope or missing required decision fields.

Use `mocksdk.IsErrorKind(err, mocksdk.ErrorKindTransport)` to branch on a class of failures.

## Scope

The SDK core contains only decision-client primitives and shared decision/event types. Protocol-specific request projection lives in adapter packages such as `github.com/MrMiaoMIMI/mocksdk/httpadapter`, `github.com/MrMiaoMIMI/mocksdk/cacheadapter`, and `github.com/MrMiaoMIMI/mocksdk/spexadapter`. gRPC, MQ, mockinject code generation, and protocol-specific forwarding implementations remain out of scope for the SDK module.
