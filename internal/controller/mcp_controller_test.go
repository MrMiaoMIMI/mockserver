package controller

import (
	"context"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
)

type testScenarioServiceForMCP struct {
	createdScenario bo.Scenario
	staticRule      bo.ScenarioRule
}

func (s *testScenarioServiceForMCP) CreateScenario(ctx context.Context, scenario bo.Scenario) (bo.Scenario, error) {
	_ = ctx
	scenario.Status = bo.ScenarioStatusActive
	scenario.Version = 1
	s.createdScenario = scenario
	return scenario, nil
}

func (s *testScenarioServiceForMCP) GetScenario(ctx context.Context, id string) (bo.Scenario, error) {
	_, _ = ctx, id
	return bo.Scenario{}, nil
}

func (s *testScenarioServiceForMCP) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]bo.Scenario, error) {
	_, _, _ = s, ctx, query
	return nil, nil
}

func (s *testScenarioServiceForMCP) UpdateScenario(ctx context.Context, id string, update bo.ScenarioUpdate) (bo.Scenario, error) {
	_, _, _, _ = s, ctx, id, update
	return bo.Scenario{}, nil
}

func (s *testScenarioServiceForMCP) DeleteScenario(ctx context.Context, id string) error {
	_, _ = ctx, id
	return nil
}

func (s *testScenarioServiceForMCP) UpsertScenarioRule(ctx context.Context, rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	_ = ctx
	s.staticRule = rule
	return rule, nil
}

func (s *testScenarioServiceForMCP) UpsertHTTPQuickRule(ctx context.Context, scenarioID string, quick bo.HTTPQuickRule) (bo.ScenarioRule, error) {
	_, _ = ctx, quick
	return bo.ScenarioRule{
		ScenarioID: scenarioID,
		RuleID:     "rule_api_orders",
		Protocol:   eo.ProtocolHTTP,
		Enabled:    true,
		Rule:       bo.Rule{ID: "rule_api_orders", Enabled: true},
		Version:    1,
	}, nil
}

func (s *testScenarioServiceForMCP) ListScenarioRules(ctx context.Context, scenarioID string) ([]bo.ScenarioRule, error) {
	_, _, _ = s, ctx, scenarioID
	return nil, nil
}

func (s *testScenarioServiceForMCP) DeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	_, _, _, _ = s, ctx, scenarioID, ruleID
	return nil
}

func (s *testScenarioServiceForMCP) ListActiveScenarioRules(ctx context.Context, scenarioID, protocol, namespace string) ([]bo.ScenarioRule, error) {
	_, _, _, _ = ctx, scenarioID, protocol, namespace
	return nil, nil
}

func (s *testScenarioServiceForMCP) SimulateScenario(ctx context.Context, scenarioID string, event bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error) {
	_, _, _, _, _, _ = ctx, scenarioID, explainOnly, explainMaxDepth, explainCompact, explainSummary
	_ = event
	return bo.SimulationResult{Matched: true}, nil
}

func TestAgentMCPSetStaticRuleBindsScenarioRule(t *testing.T) {
	scenarioService := &testScenarioServiceForMCP{}
	tools := &agentMCPTools{scenarios: scenarioService, traffic: &testTrafficServiceForController{}}

	result, err := tools.setStaticRule(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"scenario_id":      "scn_agent_orders",
				"protocol":         eo.ProtocolSPEX,
				"rule_id":          "rule_get_order",
				"condition":        map[string]any{"field": "request.cmd", "op": "eq", "value": "order.Get"},
				"response_payload": map[string]any{"order_id": 10001},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatalf("expected successful tool result: %+v", result)
	}
	if scenarioService.staticRule.ScenarioID != "scn_agent_orders" {
		t.Fatalf("unexpected scenario id: %s", scenarioService.staticRule.ScenarioID)
	}
	if scenarioService.staticRule.Protocol != eo.ProtocolSPEX || scenarioService.staticRule.Namespace != "" {
		t.Fatalf("unexpected static rule scope: %+v", scenarioService.staticRule)
	}
	if scenarioService.staticRule.Rule.Action.Response.Protocol != eo.ProtocolSPEX {
		t.Fatalf("unexpected response: %+v", scenarioService.staticRule.Rule.Action.Response)
	}
}

type mcpScenarioRepository struct {
	mu        sync.RWMutex
	scenarios map[string]bo.Scenario
	rules     map[string]map[string]bo.ScenarioRule
	nextDBID  uint64
}

func newMCPScenarioRepository() *mcpScenarioRepository {
	return &mcpScenarioRepository{
		scenarios: map[string]bo.Scenario{},
		rules:     map[string]map[string]bo.ScenarioRule{},
	}
}

func (r *mcpScenarioRepository) CreateScenario(ctx context.Context, scenario bo.Scenario) (bo.Scenario, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.scenarios[scenario.ID]; exists {
		return bo.Scenario{}, dao.ErrConflict
	}
	r.nextDBID++
	scenario.DBID = r.nextDBID
	r.scenarios[scenario.ID] = scenario
	return scenario, nil
}

func (r *mcpScenarioRepository) GetScenario(ctx context.Context, id string) (bo.Scenario, bool, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	scenario, ok := r.scenarios[id]
	return scenario, ok, nil
}

func (r *mcpScenarioRepository) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]bo.Scenario, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]bo.Scenario, 0, len(r.scenarios))
	for _, scenario := range r.scenarios {
		if query.Status != "" && scenario.Status != query.Status {
			continue
		}
		if !query.IncludeExpired && query.Now > 0 && scenario.ExpireTime <= query.Now {
			continue
		}
		items = append(items, scenario)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ExpireTime != items[j].ExpireTime {
			return items[i].ExpireTime < items[j].ExpireTime
		}
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (r *mcpScenarioRepository) UpdateScenario(ctx context.Context, scenario bo.Scenario, expectedVersion int) (bo.Scenario, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.scenarios[scenario.ID]
	if !ok {
		return bo.Scenario{}, dao.ErrNotFound
	}
	if current.Version != expectedVersion {
		return bo.Scenario{}, dao.ErrConflict
	}
	r.scenarios[scenario.ID] = scenario
	return scenario, nil
}

func (r *mcpScenarioRepository) DeleteScenario(ctx context.Context, id string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.scenarios, id)
	delete(r.rules, id)
	return nil
}

func (r *mcpScenarioRepository) UpsertScenarioRule(ctx context.Context, rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	scenario, ok := r.scenarios[rule.ScenarioID]
	if !ok {
		return bo.ScenarioRule{}, dao.ErrNotFound
	}
	if r.rules[rule.ScenarioID] == nil {
		r.rules[rule.ScenarioID] = map[string]bo.ScenarioRule{}
	}
	if current, exists := r.rules[rule.ScenarioID][rule.RuleID]; exists {
		rule.Version = current.Version + 1
	} else {
		rule.Version = 1
	}
	r.nextDBID++
	rule.DBID = r.nextDBID
	rule.ScenarioDBID = scenario.DBID
	r.rules[rule.ScenarioID][rule.RuleID] = rule
	return rule, nil
}

func (r *mcpScenarioRepository) ListScenarioRules(ctx context.Context, query bo.ScenarioRuleQuery) ([]bo.ScenarioRule, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]bo.ScenarioRule, 0, len(r.rules[query.ScenarioID]))
	for _, rule := range r.rules[query.ScenarioID] {
		if query.Protocol != "" && rule.Protocol != query.Protocol {
			continue
		}
		if query.Namespace != "" && rule.Namespace != query.Namespace {
			continue
		}
		if query.EnabledOnly && !rule.Enabled {
			continue
		}
		if query.Now > 0 && rule.ExpireTime <= query.Now {
			continue
		}
		items = append(items, rule)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority > items[j].Priority
		}
		return items[i].RuleID < items[j].RuleID
	})
	return items, nil
}

func (r *mcpScenarioRepository) DeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules[scenarioID], ruleID)
	return nil
}

func TestAgentMCPStreamableHTTPTools(t *testing.T) {
	ctx := context.Background()
	scenarioService := service.NewScenarioService(newMCPScenarioRepository())
	trafficService := &testTrafficServiceForController{}
	controller := NewAgentMCPController(scenarioService, trafficService)
	server := httptest.NewServer(controller.handler)
	defer server.Close()

	client, err := mcpclient.NewStreamableHttpClient(server.URL)
	if err != nil {
		t.Fatalf("NewStreamableHttpClient() error = %v", err)
	}
	defer client.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mockserver-mcp-test", Version: "1.0.0"}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}
	initResp, err := client.Initialize(ctx, initReq)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if initResp.ServerInfo.Name != "mockserver-agent" {
		t.Fatalf("unexpected server info: %+v", initResp.ServerInfo)
	}

	listResp, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	expectedTools := []string{
		"mockserver_create_scenario",
		"mockserver_delete_scenario",
		"mockserver_list_scenario_traffic",
		"mockserver_set_static_rule",
		"mockserver_simulate_event",
		"mockserver_upsert_rule",
	}
	gotTools := make([]string, 0, len(listResp.Tools))
	for _, tool := range listResp.Tools {
		gotTools = append(gotTools, tool.Name)
	}
	sort.Strings(gotTools)
	if len(gotTools) != len(expectedTools) {
		t.Fatalf("unexpected tool count: got=%v want=%v", gotTools, expectedTools)
	}
	for i := range expectedTools {
		if gotTools[i] != expectedTools[i] {
			t.Fatalf("unexpected tools: got=%v want=%v", gotTools, expectedTools)
		}
	}

	callMCPTool := func(name string, args map[string]any) *mcp.CallToolResult {
		t.Helper()
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      name,
				Arguments: args,
			},
		})
		if err != nil {
			t.Fatalf("%s CallTool() error = %v", name, err)
		}
		if result == nil || result.IsError {
			t.Fatalf("%s returned tool error: %+v", name, result)
		}
		return result
	}

	scenarioID := "scn_mcp_e2e_case1"
	callMCPTool("mockserver_create_scenario", map[string]any{
		"scenario_id": scenarioID,
		"ttl_seconds": 3600,
	})
	callMCPTool("mockserver_set_static_rule", map[string]any{
		"scenario_id": scenarioID,
		"protocol":    eo.ProtocolHTTP,
		"rule_id":     "rule_api_orders",
		"condition": map[string]any{
			"field": "request.path",
			"op":    "eq",
			"value": "/api/orders",
		},
		"response_payload": map[string]any{
			"status": 202,
			"body":   map[string]any{"ok": true},
		},
	})
	callMCPTool("mockserver_upsert_rule", map[string]any{
		"scenario_id": scenarioID,
		"protocol":    eo.ProtocolHTTP,
		"rule": map[string]any{
			"id":       "rule_api_users",
			"name":     "User API",
			"enabled":  true,
			"priority": 50,
			"when": map[string]any{
				"field": "request.path",
				"op":    "eq",
				"value": "/api/users",
			},
			"action": map[string]any{
				"type":     "respond",
				"renderer": "static",
				"response": map[string]any{
					"protocol": "http",
					"payload": map[string]any{
						"status": 200,
						"body":   map[string]any{"user_id": 10001},
					},
				},
			},
		},
	})
	simulateResult := callMCPTool("mockserver_simulate_event", map[string]any{
		"scenario_id": scenarioID,
		"event": map[string]any{
			"protocol":  eo.ProtocolHTTP,
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"path":   "/api/orders",
			},
		},
	})
	if simulateResult.StructuredContent == nil {
		t.Fatalf("simulate result missing structured content: %+v", simulateResult)
	}
	callMCPTool("mockserver_list_scenario_traffic", map[string]any{
		"scenario_id": scenarioID,
		"limit":       10,
	})
	if trafficService.count != 0 {
		t.Fatalf("list traffic tool should not record SDK traffic, count=%d", trafficService.count)
	}
	callMCPTool("mockserver_delete_scenario", map[string]any{
		"scenario_id": scenarioID,
	})
}
