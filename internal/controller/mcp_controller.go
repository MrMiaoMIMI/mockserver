package controller

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
)

type AgentMCPController struct {
	handler http.Handler
}

type agentMCPTools struct {
	scenarios service.ScenarioService
	traffic   service.TrafficService
}

type mcpCreateScenarioArgs struct {
	ScenarioID  string `json:"scenario_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TTLSeconds  uint64 `json:"ttl_seconds,omitempty"`
}

type mcpSetStaticRuleArgs struct {
	ScenarioID      string       `json:"scenario_id"`
	Protocol        string       `json:"protocol"`
	RuleID          string       `json:"rule_id,omitempty"`
	Name            string       `json:"name,omitempty"`
	Enabled         *bool        `json:"enabled,omitempty"`
	Priority        int          `json:"priority,omitempty"`
	Condition       bo.Condition `json:"condition"`
	ResponsePayload any          `json:"response_payload,omitempty"`
}

type mcpUpsertRuleArgs struct {
	ScenarioID string  `json:"scenario_id"`
	Protocol   string  `json:"protocol"`
	Rule       bo.Rule `json:"rule"`
}

type mcpSimulateEventArgs struct {
	ScenarioID      string   `json:"scenario_id"`
	Event           bo.Event `json:"event"`
	ExplainOnly     bool     `json:"explain_only,omitempty"`
	ExplainMaxDepth int      `json:"explain_max_depth,omitempty"`
	ExplainCompact  bool     `json:"explain_compact,omitempty"`
	ExplainSummary  bool     `json:"explain_summary,omitempty"`
}

type mcpListTrafficArgs struct {
	ScenarioID   string `json:"scenario_id"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
	StartTime    uint64 `json:"start_time,omitempty"`
	EndTime      uint64 `json:"end_time,omitempty"`
	ProtocolName string `json:"protocol_name,omitempty"`
	NamespaceID  string `json:"namespace_id,omitempty"`
	Outcome      string `json:"outcome,omitempty"`
	DecisionKind string `json:"decision_kind,omitempty"`
	RuleID       string `json:"rule_id,omitempty"`
}

type mcpDeleteScenarioArgs struct {
	ScenarioID string `json:"scenario_id"`
}

const scenarioIDPattern = `^scn_[A-Za-z0-9_-]{4,124}$`

const mcpScenarioInstructions = `Use these tools to create short-lived mock scenarios for AI test execution. Preferred flow: 1) mockserver_create_scenario, 2) mockserver_set_static_rule for each dependency response, 3) run the business test with the returned scenario_id injected into mocksdk context, 4) mockserver_list_scenario_traffic if assertions need recorded calls, 5) mockserver_delete_scenario. Scenario rules do not take namespace; namespace belongs to the SDK event and only selects fallback behavior when no scenario rule matches.`

const createScenarioDescription = `Create one isolated, short-lived scenario for one test run. Return value contains scenario_id; pass it to mocksdk context. Do not create namespaces or published rulesets for AI tests.`

const scenarioIDDescription = `Scenario id. Must match ^scn_[A-Za-z0-9_-]{4,124}$ and be at most 128 chars. Use the id returned by mockserver_create_scenario. Only provide a custom id when it helps debugging, for example scn_checkout_timeout_case1; otherwise omit it on create so MockServer generates one.`

const setStaticRuleDescription = `Preferred rule tool for AI agents. Create or update one scenario rule that returns a static response for http, spex, or cache. Use this instead of mockserver_upsert_rule unless you need advanced action types such as template, CEL, sequence, or webhook.`

const staticConditionDescription = `Required rule condition. Use exactly the request fields produced by the SDK protocol adapter. Common examples: HTTP {"field":"request.path","op":"eq","value":"/api/orders"} or {"all":[{"field":"request.method","op":"eq","value":"POST"},{"field":"request.path","op":"prefix","value":"/api/payment"}]}; SPEX {"field":"request.cmd","op":"eq","value":"shop.GetOrder"}; cache {"field":"request.key","op":"eq","value":"user:10001"}. Common ops: eq, ne, prefix, contains, regex, exists.`

const staticResponseDescription = `Static protocol response payload. Required for useful mocks. HTTP shape: {"status":200,"headers":{"X-Mock":["yes"]},"body":{"ok":true}}. SPEX shape: {"code":0,"resp":{"ok":true}}. Cache shape: {"hit":true,"value":{"ok":true}}. Do not wrap this in {"protocol":...,"payload":...}; this field is the payload itself.`

const rawRuleDescription = `Advanced escape hatch: create or update a full mockserver Rule object for any supported protocol. Prefer mockserver_set_static_rule for normal AI tests. Use this only when the rule needs non-static renderer/action fields. Rule must include id, name, enabled, priority, when, and action.`

const simulateEventDescription = `Simulate an event against scenario rules without recording traffic. Use this to debug whether a newly-created rule matches before running the real business test. The event.namespace is required because it selects fallback policy when no scenario rule matches; scenario rules themselves do not store namespace.`

const eventDescription = `Mockserver event object. Required fields: protocol, namespace, request. protocol must be http, spex, or cache. HTTP example: {"protocol":"http","namespace":"default","request":{"method":"GET","path":"/api/orders"}}. SPEX example: {"protocol":"spex","namespace":"default","request":{"cmd":"shop.GetOrder","req":{"order_id":1}}}. Cache example: {"protocol":"cache","namespace":"default","request":{"operation":"get","key":"user:10001"}}.`

const trafficDescription = `List SDK decision traffic recorded for a scenario after the business test runs. Use filters only when narrowing assertions; otherwise pass only scenario_id and optionally limit. Times are Unix milliseconds.`

const deleteScenarioDescription = `Delete one scenario and all of its scenario rules after the test run. Call this during cleanup when the scenario is no longer needed.`

func NewAgentMCPController(scenarioService service.ScenarioService, trafficService service.TrafficService) *AgentMCPController {
	tools := &agentMCPTools{scenarios: scenarioService, traffic: trafficService}
	server := mcpserver.NewMCPServer(
		"mockserver-agent",
		"0.1.0",
		mcpserver.WithRecovery(),
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithInstructions(mcpScenarioInstructions),
	)
	tools.register(server)
	return &AgentMCPController{
		handler: mcpserver.NewStreamableHTTPServer(server, mcpserver.WithStateLess(true)),
	}
}

func (c *AgentMCPController) Handle(ctx *gin.Context) {
	c.handler.ServeHTTP(ctx.Writer, ctx.Request)
}

func (t *agentMCPTools) register(server *mcpserver.MCPServer) {
	server.AddTool(mcp.NewTool(
		"mockserver_create_scenario",
		mcp.WithDescription(createScenarioDescription),
		mcp.WithString("scenario_id", mcp.Description(scenarioIDDescription), mcp.Pattern(scenarioIDPattern), mcp.MinLength(8), mcp.MaxLength(128)),
		mcp.WithString("name", mcp.Description("Optional human-readable scenario name for UI display. Omit for AI-created scenarios unless a readable label helps the user distinguish multiple active scenarios."), mcp.MaxLength(128)),
		mcp.WithString("description", mcp.Description("Optional short note shown in the scenarios page. Omit unless it adds useful debugging context for humans."), mcp.MaxLength(512)),
		mcp.WithNumber("ttl_seconds", mcp.Description("Scenario lifetime in seconds. Omit to use 3600. Use 300-3600 for normal AI tests; use a larger value only for long-running manual debugging."), mcp.Min(1), mcp.DefaultNumber(3600)),
	), t.createScenario)

	server.AddTool(mcp.NewTool(
		"mockserver_set_static_rule",
		mcp.WithDescription(setStaticRuleDescription),
		mcp.WithString("scenario_id", mcp.Description("Required. Existing scenario id returned by mockserver_create_scenario. This scopes the rule to one short-lived test scenario."), mcp.Required(), mcp.Pattern(scenarioIDPattern), mcp.MinLength(8), mcp.MaxLength(128)),
		mcp.WithString("protocol", mcp.Description("Required protocol enum. Use http for HTTP dependency APIs, spex for SPEX/RPC dependency calls, cache for cache get/set style dependencies."), mcp.Required(), mcp.Enum(eo.ProtocolHTTP, eo.ProtocolSPEX, eo.ProtocolCache)),
		mcp.WithString("rule_id", mcp.Description("Optional stable rule id. Omit when creating a new one-off rule. Provide only when updating the same rule later or when test assertions need a predictable rule_id. Allowed characters: letters, numbers, underscore, hyphen; max 64."), mcp.Pattern(`^[A-Za-z0-9_-]{1,64}$`), mcp.MaxLength(64)),
		mcp.WithString("name", mcp.Description("Optional display name. Omit for normal AI-created rules; MockServer will use rule_id. Provide a concise label only when humans will inspect the scenarios page."), mcp.MaxLength(128)),
		mcp.WithBoolean("enabled", mcp.Description("Optional. Omit for normal mocks; default is true. Set false only to temporarily keep a rule stored but inactive."), mcp.DefaultBool(true)),
		mcp.WithNumber("priority", mcp.Description("Optional match order. Higher number wins when multiple rules match the same event. Omit or use 0 for a single rule. Use 100, 90, 80... when creating overlapping rules."), mcp.DefaultNumber(0)),
		mcp.WithObject("condition", mcp.Description(staticConditionDescription), mcp.Required()),
		mcp.WithAny("response_payload", mcp.Description(staticResponseDescription), mcp.Required()),
	), t.setStaticRule)

	server.AddTool(mcp.NewTool(
		"mockserver_upsert_rule",
		mcp.WithDescription(rawRuleDescription),
		mcp.WithString("scenario_id", mcp.Description("Required. Existing scenario id returned by mockserver_create_scenario. Rule is stored only inside this scenario."), mcp.Required(), mcp.Pattern(scenarioIDPattern), mcp.MinLength(8), mcp.MaxLength(128)),
		mcp.WithString("protocol", mcp.Description("Required protocol enum for validating rule.when fields and action.response protocol. Allowed values: http, spex, cache."), mcp.Required(), mcp.Enum(eo.ProtocolHTTP, eo.ProtocolSPEX, eo.ProtocolCache)),
		mcp.WithObject("rule", mcp.Description("Required full Rule object. Example static SPEX rule: {\"id\":\"rule_get_order\",\"name\":\"Get order\",\"enabled\":true,\"priority\":100,\"when\":{\"field\":\"request.cmd\",\"op\":\"eq\",\"value\":\"shop.GetOrder\"},\"action\":{\"type\":\"respond\",\"renderer\":\"static\",\"response\":{\"protocol\":\"spex\",\"payload\":{\"code\":0,\"resp\":{\"order_id\":1}}}}}. Do not include namespace."), mcp.Required()),
	), t.upsertRule)

	server.AddTool(mcp.NewTool(
		"mockserver_simulate_event",
		mcp.WithDescription(simulateEventDescription),
		mcp.WithString("scenario_id", mcp.Description("Required. Existing scenario id whose rules should be tested."), mcp.Required(), mcp.Pattern(scenarioIDPattern), mcp.MinLength(8), mcp.MaxLength(128)),
		mcp.WithObject("event", mcp.Description(eventDescription), mcp.Required()),
		mcp.WithBoolean("explain_only", mcp.Description("Optional. Default false. Set true only when you want diagnostics without returning/applying the matched response."), mcp.DefaultBool(false)),
		mcp.WithNumber("explain_max_depth", mcp.Description("Optional diagnostics depth limit. Omit for normal simulation. Use 1-3 for compact debugging output; 0 means no explicit depth limit."), mcp.Min(0)),
		mcp.WithBoolean("explain_compact", mcp.Description("Optional. Default false. Set true to remove verbose expected/actual details from explain output."), mcp.DefaultBool(false)),
		mcp.WithBoolean("explain_summary", mcp.Description("Optional. Default false. Set true for most AI debugging; it returns concise rule-selection and match summaries."), mcp.DefaultBool(false)),
	), t.simulateEvent)

	server.AddTool(mcp.NewTool(
		"mockserver_list_scenario_traffic",
		mcp.WithDescription(trafficDescription),
		mcp.WithString("scenario_id", mcp.Description("Required. Scenario id injected into the SDK during the business test."), mcp.Required(), mcp.Pattern(scenarioIDPattern), mcp.MinLength(8), mcp.MaxLength(128)),
		mcp.WithNumber("limit", mcp.Description("Optional page size. Omit for default 50. Use 1-200; the service caps overly large values."), mcp.Min(1), mcp.Max(200), mcp.DefaultNumber(50)),
		mcp.WithNumber("offset", mcp.Description("Optional zero-based page offset. Omit for the first page."), mcp.Min(0), mcp.DefaultNumber(0)),
		mcp.WithNumber("start_time", mcp.Description("Optional inclusive start time in Unix milliseconds, not seconds. Omit unless narrowing traffic by time window."), mcp.Min(0)),
		mcp.WithNumber("end_time", mcp.Description("Optional exclusive end time in Unix milliseconds, not seconds. Omit unless narrowing traffic by time window."), mcp.Min(0)),
		mcp.WithString("protocol_name", mcp.Description("Optional protocol filter enum. Use only when asserting one protocol's traffic. Allowed values: http, spex, cache."), mcp.Enum(eo.ProtocolHTTP, eo.ProtocolSPEX, eo.ProtocolCache)),
		mcp.WithString("namespace_id", mcp.Description("Optional namespace filter from the SDK event, for example default or shop-sg. This is for traffic filtering only; scenario rules do not store namespace."), mcp.Pattern(`^[A-Za-z0-9_-]{1,64}$`), mcp.MaxLength(64)),
		mcp.WithString("outcome", mcp.Description("Optional outcome filter enum. Use matched for scenario/rule hits, fallback for namespace fallback, unmatched when no decision matched and no fallback response was produced, error for decision errors. Omit to see all outcomes."), mcp.Enum(bo.TrafficOutcomeMatched, bo.TrafficOutcomeFallback, bo.TrafficOutcomeUnmatched, bo.TrafficOutcomeError)),
		mcp.WithString("decision_kind", mcp.Description("Optional decision kind filter enum. Use response for mocked responses, forward for pass-through decisions. Omit to see both."), mcp.Enum(eo.DecisionKindResponse, eo.DecisionKindForward)),
		mcp.WithString("rule_id", mcp.Description("Optional rule id filter. Use only when checking whether a specific scenario rule was hit."), mcp.Pattern(`^[A-Za-z0-9_-]{1,64}$`), mcp.MaxLength(64)),
	), t.listScenarioTraffic)

	server.AddTool(mcp.NewTool(
		"mockserver_delete_scenario",
		mcp.WithDescription(deleteScenarioDescription),
		mcp.WithString("scenario_id", mcp.Description("Required. Scenario id to delete. This removes all scenario rules too; do not pass a namespace or rule id."), mcp.Required(), mcp.Pattern(scenarioIDPattern), mcp.MinLength(8), mcp.MaxLength(128)),
	), t.deleteScenario)
}

func (t *agentMCPTools) createScenario(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args mcpCreateScenarioArgs
	if err := req.BindArguments(&args); err != nil {
		return mcpToolError(err), nil
	}
	scenario := bo.Scenario{
		ID:          strings.TrimSpace(args.ScenarioID),
		Name:        strings.TrimSpace(args.Name),
		Description: strings.TrimSpace(args.Description),
	}
	if args.TTLSeconds > 0 {
		scenario.ExpireTime = scenarioExpireTimeFromTTLSeconds(args.TTLSeconds)
	}
	created, err := t.scenarios.CreateScenario(ctx, scenario)
	if err != nil {
		return mcpToolError(err), nil
	}
	return mcpToolJSON(response.NewScenarioResponse(created))
}

func (t *agentMCPTools) setStaticRule(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args mcpSetStaticRuleArgs
	if err := req.BindArguments(&args); err != nil {
		return mcpToolError(err), nil
	}
	protocol := strings.ToLower(strings.TrimSpace(args.Protocol))
	ruleID := strings.TrimSpace(args.RuleID)
	if ruleID == "" {
		ruleID = protocol + "-" + mcpRandomIDToken()
	}
	name := strings.TrimSpace(args.Name)
	if name == "" {
		name = ruleID
	}
	enabled := true
	if args.Enabled != nil {
		enabled = *args.Enabled
	}
	rule, err := t.scenarios.UpsertScenarioRule(ctx, bo.ScenarioRule{
		ScenarioID: args.ScenarioID,
		RuleID:     ruleID,
		Name:       name,
		Protocol:   protocol,
		Enabled:    enabled,
		Priority:   args.Priority,
		Rule: bo.Rule{
			ID:       ruleID,
			Name:     name,
			Enabled:  enabled,
			Priority: args.Priority,
			When:     args.Condition,
			Action: bo.Action{
				Type:     eo.ActionTypeRespond,
				Renderer: eo.ActionRendererStatic,
				Response: &bo.ProtocolResponse{
					Protocol: protocol,
					Payload:  mcpResponsePayload(args.ResponsePayload),
				},
			},
		},
	})
	if err != nil {
		return mcpToolError(err), nil
	}
	return mcpToolJSON(response.NewScenarioRuleResponse(rule))
}

func (t *agentMCPTools) upsertRule(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args mcpUpsertRuleArgs
	if err := req.BindArguments(&args); err != nil {
		return mcpToolError(err), nil
	}
	rule, err := t.scenarios.UpsertScenarioRule(ctx, bo.ScenarioRule{
		ScenarioID: args.ScenarioID,
		RuleID:     args.Rule.ID,
		Name:       args.Rule.Name,
		Protocol:   args.Protocol,
		Enabled:    args.Rule.Enabled,
		Priority:   args.Rule.Priority,
		Rule:       args.Rule,
	})
	if err != nil {
		return mcpToolError(err), nil
	}
	return mcpToolJSON(response.NewScenarioRuleResponse(rule))
}

func (t *agentMCPTools) simulateEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args mcpSimulateEventArgs
	if err := req.BindArguments(&args); err != nil {
		return mcpToolError(err), nil
	}
	if strings.TrimSpace(args.Event.Protocol) == "" {
		args.Event.Protocol = eo.ProtocolHTTP
	}
	result, err := t.scenarios.SimulateScenario(ctx, args.ScenarioID, args.Event, args.ExplainOnly, args.ExplainMaxDepth, args.ExplainCompact, args.ExplainSummary)
	if err != nil {
		return mcpToolError(err), nil
	}
	return mcpToolJSON(response.SimulateScenarioResponse{Result: result})
}

func (t *agentMCPTools) listScenarioTraffic(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args mcpListTrafficArgs
	if err := req.BindArguments(&args); err != nil {
		return mcpToolError(err), nil
	}
	list, err := t.traffic.ListTrafficEvents(ctx, bo.TrafficQuery{
		Limit:        args.Limit,
		Offset:       args.Offset,
		StartTime:    args.StartTime,
		EndTime:      args.EndTime,
		ScenarioID:   strings.TrimSpace(args.ScenarioID),
		ProtocolName: strings.TrimSpace(args.ProtocolName),
		NamespaceID:  strings.TrimSpace(args.NamespaceID),
		Outcome:      strings.TrimSpace(args.Outcome),
		DecisionKind: strings.TrimSpace(args.DecisionKind),
		RuleID:       strings.TrimSpace(args.RuleID),
	})
	if err != nil {
		return mcpToolError(err), nil
	}
	return mcpToolJSON(response.NewListTrafficEventsResponse(list))
}

func (t *agentMCPTools) deleteScenario(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args mcpDeleteScenarioArgs
	if err := req.BindArguments(&args); err != nil {
		return mcpToolError(err), nil
	}
	if err := t.scenarios.DeleteScenario(ctx, args.ScenarioID); err != nil {
		return mcpToolError(err), nil
	}
	return mcpToolJSON(map[string]any{"deleted": true, "scenario_id": strings.TrimSpace(args.ScenarioID)})
}

func mcpToolJSON(value any) (*mcp.CallToolResult, error) {
	result, err := mcp.NewToolResultJSON(value)
	if err != nil {
		return mcpToolError(err), nil
	}
	return result, nil
}

func mcpToolError(err error) *mcp.CallToolResult {
	if err == nil {
		return mcp.NewToolResultError("unknown error")
	}
	return mcp.NewToolResultError(err.Error())
}

func mcpResponsePayload(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{"value": value}
}

func mcpRandomIDToken() string {
	var data [4]byte
	if _, err := rand.Read(data[:]); err == nil {
		return hex.EncodeToString(data[:])
	}
	return strings.ToLower(time.Now().UTC().Format("150405000"))
}
