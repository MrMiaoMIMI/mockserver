# Agent Scenario Overlay

Agent Scenario Overlay is the short-lived mock layer for AI-driven tests and
normal business tests that need isolated dependency responses. It does not
publish rulesets and does not change the business process environment.

## Scenario ID Contract

Every scenario is addressed by `scenario_id`.

- Must start with `scn_`
- Total length must be between 8 and 128 characters
- Allowed characters are ASCII letters, digits, underscore, and hyphen
- Must be unique while active

Callers may provide a meaningful id such as `scn_checkout_timeout`; when omitted,
MockServer generates a compact random id.

## Runtime Flow

1. Create a scenario.
2. Add one or more scenario rules.
3. Run the business process with `mocksdk.WithScenarioID(ctx, scenarioID)`.
4. The SDK sends `event.meta.scenario_id` to MockServer.
5. MockServer evaluates the scenario id first. Scenario rules are scoped by
   scenario id and protocol only; they do not configure namespace.
6. If the scenario does not exist, is inactive, or no scenario rule matches,
   MockServer applies the fallback policy from the event namespace. It does not
   fall through to published rulesets while `scenario_id` is present.
7. Query scenario traffic, then delete the scenario when the test is done.

The business process can still be used for normal tests. `scenario_id` is
carried in request context by mockinject or the caller, not by a process-wide
environment variable.

## HTTP Agent API

Agent APIs use the same JWT auth as admin APIs:

```text
POST   /mockserver/api/v1/agent/scenarios
GET    /mockserver/api/v1/agent/scenarios
GET    /mockserver/api/v1/agent/scenarios/:scenario_id
PATCH  /mockserver/api/v1/agent/scenarios/:scenario_id
GET    /mockserver/api/v1/agent/scenarios/:scenario_id/rules
POST   /mockserver/api/v1/agent/scenarios/:scenario_id/rules/quick
PUT    /mockserver/api/v1/agent/scenarios/:scenario_id/rules/:rule_id
DELETE /mockserver/api/v1/agent/scenarios/:scenario_id/rules/:rule_id
POST   /mockserver/api/v1/agent/scenarios/:scenario_id/simulate
GET    /mockserver/api/v1/agent/scenarios/:scenario_id/traffic
DELETE /mockserver/api/v1/agent/scenarios/:scenario_id
```

The console uses the list/detail/rules endpoints to show short-lived scenarios,
extend TTL, inspect scenario rules, and drill into SDK traffic generated with
the scenario id. Scenario `expire_time` values in API responses and database
rows are Unix milliseconds; `ttl_seconds` is only the creation/update input
unit.

Create a scenario:

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/agent/scenarios \
  -H "Authorization: Bearer $JWT" \
  -H 'Content-Type: application/json' \
  -d '{"scenario_id":"scn_checkout_timeout","ttl_seconds":3600}'
```

Add a quick HTTP rule:

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/agent/scenarios/scn_checkout_timeout/rules/quick \
  -H "Authorization: Bearer $JWT" \
  -H 'Content-Type: application/json' \
  -d '{
    "rule_id": "rule_payment_timeout",
    "priority": 100,
    "match": {
      "method": "POST",
      "path": "/api/payment/charge"
    },
    "respond": {
      "status": 504,
      "body": {"error": "payment timeout"}
    }
  }'
```

## MCP Server

MockServer exposes a Streamable HTTP MCP endpoint at:

```text
/mockserver/api/v1/agent/mcp
```

The endpoint is registered inside the authenticated agent route group, so MCP
clients must send the same `Authorization: Bearer <jwt>` header.

Available tools:

- `mockserver_create_scenario`
- `mockserver_set_static_rule`
- `mockserver_upsert_rule`
- `mockserver_simulate_event`
- `mockserver_list_scenario_traffic`
- `mockserver_delete_scenario`

The MCP server is a protocol adapter over the existing scenario services.
`mockserver_set_static_rule` is the preferred AI-agent path: provide
`scenario_id`, `protocol`, a condition object, and a static protocol response
payload. Use `mockserver_upsert_rule` only when the full rule object is needed.

## SDK Injection

`mocksdk` validates and propagates `scenario_id` from context:

```go
ctx = mocksdk.WithScenarioID(ctx, "scn_checkout_timeout")

decision, err := client.Decide(ctx, mocksdk.Event{
    Protocol:  "http",
    Operation: "request",
    Namespace: "default",
    Request: mocksdk.EventRequest{
        "method": "POST",
        "path":   "/api/payment/charge",
    },
})
```

If the caller also sets `event.Meta.ScenarioID`, the explicit event value wins.
Invalid scenario ids fail fast as SDK configuration errors before the decision
request is sent.

Direct MockServer HTTP runtime calls can also pass:

```text
X-Mockserver-Scenario-ID: scn_checkout_timeout
```

That header is only for the direct runtime adapter. Mockinject-based business
tests should prefer the SDK context path.

## Storage

Scenario state is stored in:

- `mockserver_scenario_tab`
- `mockserver_scenario_rule_tab`

`mockserver_scenario_tab.expire_time`, `mockserver_scenario_tab.scenario_json`
`expire_time`, and `mockserver_scenario_rule_tab.expire_time` use Unix
milliseconds.

SDK traffic stores the matched scenario in
`mockserver_traffic_event_tab.scenario_code`. Scenario rules no longer use
`namespace_code`; event namespace remains the fallback-policy selector. See
`docs/db_schema.sql` for DDL.
