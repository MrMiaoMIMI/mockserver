# Mock SDK Quickstart

Mock SDK is the runtime decision layer for generated or injected code. It is intended for `mockinject` implementations that intercept an original HTTP request, ask MockServer for a decision, and then either return a mock response or continue with protocol-specific forwarding outside the SDK.

The SDK is maintained as a separate Go module:

```text
github.com/MrMiaoMIMI/mocksdk
```

It supports Go 1.17+ and uses only the Go standard library so injected business repositories do not inherit MockServer's higher Go toolchain requirement.

## Tool Chain Roles

- `mockserver` manages namespaces, draft rulesets, published snapshots, matching, fallback policy, and the SDK decision endpoint.
- `mocksdk` converts protocol-specific calls into MockServer events, calls the SDK decision endpoint, and exposes typed decisions. It does not execute upstream forwarding.
- `mockinject` is external to this repository. It injects code before business compilation, calls `mocksdk`, and owns protocol-specific execution such as forwarding the original request.

## Environment Variables

Environment variables have higher priority than explicit SDK options.

| Variable | Meaning |
| --- | --- |
| `MOCKSERVER_HOST` | Base URL of MockServer, for example `http://127.0.0.1:8080`. Overrides `Config.MockServerURL`. |
| `MOCKSERVER_NAMESPACE_ID` | Namespace used by adapter helpers when the caller does not pass a namespace. Overrides `Config.Namespace`. |
| `MOCKSERVER_TIMEOUT_MS` | Optional decision-call timeout in milliseconds. Overrides `Config.Timeout` when set to a positive integer. |

If neither `MOCKSERVER_NAMESPACE_ID` nor `Config.Namespace` is set, the SDK uses `default`.

## Decision Endpoint

The SDK calls:

```text
POST /mockserver/api/v1/sdk/decision
```

This endpoint evaluates published rulesets only. Draft rules do not affect SDK decisions until they are published.

The endpoint returns MockServer's standard response envelope. The payload is `data.decision`.

### Response Decision

`kind=response` means injected code should return the decision response to the business process.

This can happen when:

- A published rule matches and produces a mock response.
- A namespace fallback policy is configured as `response`.

Important fields:

- `matched`: true for a rule hit, false for a fallback response.
- `fallback`: true when the response came from namespace fallback.
- `trace.ruleset_id`: matched ruleset ID when available.
- `trace.rule_id`: matched rule ID when available.
- `trace.fallback_reason`: `ruleset_miss` or `rule_miss` for fallback responses.
- `response.status`, `response.headers`, `response.body`: HTTP response to apply.
- `meta.trace_id`: trace ID propagated from the original request event.

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
