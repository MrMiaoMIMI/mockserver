package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/gin-gonic/gin"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"

	authlib "github.com/MrMiaoMIMI/mockserver/internal/auth"
	"github.com/MrMiaoMIMI/mockserver/internal/controller"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

func TestTraceMiddlewareSetsLoggerTraceID(t *testing.T) {
	var observedTraceID string
	engine := gin.New()
	engine.Use(traceMiddleware())
	engine.GET("/mockserver/runtime/default/http/api", func(c *gin.Context) {
		observedTraceID = logger.GetTraceID(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/runtime/default/http/api", nil)
	req.Header.Set("X-Trace-ID", "trace-router-test")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if observedTraceID != "trace-router-test" {
		t.Fatalf("unexpected context trace id: %q", observedTraceID)
	}
	if got := recorder.Header().Get("X-Trace-ID"); got != "trace-router-test" {
		t.Fatalf("unexpected response trace id: %q", got)
	}
}

func TestOperatorMiddlewareSetsDBOperator(t *testing.T) {
	var observedOperator string
	engine := gin.New()
	engine.Use(operatorMiddleware())
	engine.POST("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		observedOperator, _ = dbspi.OperatorFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("X-Mockserver-Operator", " admin@example.com ")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if observedOperator != "admin@example.com" {
		t.Fatalf("unexpected context operator: %q", observedOperator)
	}
}

func TestJWTAuthMiddlewareSetsEmailAsDefaultOperator(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	token, err := authlib.GenerateToken(config.JWT, "jwt-user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var observedEmail string
	var observedOperator string
	engine := gin.New()
	engine.Use(operatorMiddleware(), jwtAuthMiddleware(config))
	engine.GET("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		email, _ := c.Get(authlib.UserEmailContextKey)
		observedEmail, _ = email.(string)
		observedOperator, _ = dbspi.OperatorFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if observedEmail != "jwt-user@example.com" {
		t.Fatalf("unexpected user email: %q", observedEmail)
	}
	if observedOperator != "jwt-user@example.com" {
		t.Fatalf("unexpected context operator: %q", observedOperator)
	}
}

func TestJWTAuthMiddlewarePreservesExplicitOperatorHeader(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	token, err := authlib.GenerateToken(config.JWT, "jwt-user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var observedOperator string
	engine := gin.New()
	engine.Use(operatorMiddleware(), jwtAuthMiddleware(config))
	engine.POST("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		observedOperator, _ = dbspi.OperatorFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Mockserver-Operator", "manual@example.com")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if observedOperator != "manual@example.com" {
		t.Fatalf("unexpected context operator: %q", observedOperator)
	}
}

func TestJWTAuthMiddlewareRejectsMissingToken(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	engine := gin.New()
	engine.Use(jwtAuthMiddleware(config))
	engine.GET("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["message"] != "jwt token is required" {
		t.Fatalf("unexpected response body: %v", body)
	}
}

func TestJWTAuthMiddlewareRejectsLegacyAdminTokenHeader(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	engine := gin.New()
	engine.Use(jwtAuthMiddleware(config))
	engine.GET("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("X-Mockserver-Admin-Token", "legacy-token")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestRecoveryMiddlewareCapturesPanic(t *testing.T) {
	engine := gin.New()
	engine.Use(traceMiddleware(), accessLogMiddleware(), recoveryMiddleware())
	engine.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Trace-ID", "trace-panic-test")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if got := recorder.Header().Get("X-Trace-ID"); got != "trace-panic-test" {
		t.Fatalf("unexpected response trace id: %q", got)
	}
}

func TestNewWithConfigAppliesAPIPrefixToAllRoutes(t *testing.T) {
	adminController := controller.NewAdminController(nil, nil)
	runtimeController := controller.NewRuntimeController(nil, nil)
	engine := NewWithConfig(adminController, runtimeController, AuthConfig{}, RouteConfig{APIPrefix: "tenant-a/"}, nil, nil)

	paths := map[string]bool{}
	for _, route := range engine.Routes() {
		paths[route.Path] = true
	}

	expectedPaths := []string{
		"/tenant-a/mockserver/api/v1/auth/debug/login",
		"/tenant-a/mockserver/api/v1/auth/me",
		"/tenant-a/mockserver/api/v1/admin/rulesets",
		"/tenant-a/mockserver/api/v1/sdk/decision",
		"/tenant-a/mockserver/runtime/:namespace/http",
		"/tenant-a/mockserver/runtime/:namespace/http/*runtime_path",
	}
	for _, path := range expectedPaths {
		if !paths[path] {
			t.Fatalf("expected route %s to be registered", path)
		}
	}
	if paths["/mockserver/api/v1/admin/rulesets"] {
		t.Fatalf("unexpected unprefixed admin route")
	}
}

type routerMCPScenarioService struct{}

func (s *routerMCPScenarioService) CreateScenario(ctx context.Context, scenario bo.Scenario) (bo.Scenario, error) {
	_, _ = s, ctx
	return scenario, nil
}

func (s *routerMCPScenarioService) GetScenario(ctx context.Context, id string) (bo.Scenario, error) {
	_, _, _ = s, ctx, id
	return bo.Scenario{}, nil
}

func (s *routerMCPScenarioService) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]bo.Scenario, error) {
	_, _, _ = s, ctx, query
	return nil, nil
}

func (s *routerMCPScenarioService) UpdateScenario(ctx context.Context, id string, update bo.ScenarioUpdate) (bo.Scenario, error) {
	_, _, _, _ = s, ctx, id, update
	return bo.Scenario{}, nil
}

func (s *routerMCPScenarioService) DeleteScenario(ctx context.Context, id string) error {
	_, _, _ = s, ctx, id
	return nil
}

func (s *routerMCPScenarioService) UpsertScenarioRule(ctx context.Context, rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	_, _ = s, ctx
	return rule, nil
}

func (s *routerMCPScenarioService) UpsertHTTPQuickRule(ctx context.Context, scenarioID string, quick bo.HTTPQuickRule) (bo.ScenarioRule, error) {
	_, _, _ = s, ctx, quick
	return bo.ScenarioRule{ScenarioID: scenarioID}, nil
}

func (s *routerMCPScenarioService) ListScenarioRules(ctx context.Context, scenarioID string) ([]bo.ScenarioRule, error) {
	_, _, _ = s, ctx, scenarioID
	return nil, nil
}

func (s *routerMCPScenarioService) DeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	_, _, _, _ = s, ctx, scenarioID, ruleID
	return nil
}

func (s *routerMCPScenarioService) ListActiveScenarioRules(ctx context.Context, scenarioID, protocol, namespace string) ([]bo.ScenarioRule, error) {
	_, _, _, _, _ = s, ctx, scenarioID, protocol, namespace
	return nil, nil
}

func (s *routerMCPScenarioService) SimulateScenario(ctx context.Context, scenarioID string, event bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error) {
	_, _, _, _, _, _, _, _ = s, ctx, scenarioID, event, explainOnly, explainMaxDepth, explainCompact, explainSummary
	return bo.SimulationResult{}, nil
}

type routerMCPTrafficService struct{}

func (s *routerMCPTrafficService) RecordSDKDecision(ctx context.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, durationMS uint32) (bo.TrafficEvent, error) {
	_, _, _, _, _, _ = s, ctx, event, decision, decisionErr, durationMS
	return bo.TrafficEvent{}, nil
}

func (s *routerMCPTrafficService) GetTrafficEvent(ctx context.Context, id uint64) (bo.TrafficEvent, error) {
	_, _, _ = s, ctx, id
	return bo.TrafficEvent{}, nil
}

func (s *routerMCPTrafficService) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	_, _, _ = s, ctx, query
	return bo.TrafficEventList{}, nil
}

func TestAgentMCPRouteRequiresJWTAndListsTools(t *testing.T) {
	authConfig := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	scenarioService := &routerMCPScenarioService{}
	trafficService := &routerMCPTrafficService{}
	engine := NewWithConfigAndAgentMCP(
		controller.NewAdminController(nil, nil),
		controller.NewRuntimeController(nil, nil),
		authConfig,
		RouteConfig{},
		nil,
		controller.NewTrafficController(trafficService),
		controller.NewAgentController(scenarioService, trafficService),
		controller.NewAgentMCPController(scenarioService, trafficService),
	)

	req := httptest.NewRequest(http.MethodPost, "/mockserver/api/v1/agent/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status without jwt: %d", recorder.Code)
	}

	token, err := authlib.GenerateToken(authConfig.JWT, "mcp-user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	server := httptest.NewServer(engine)
	defer server.Close()

	client, err := mcpclient.NewStreamableHttpClient(
		server.URL+"/mockserver/api/v1/agent/mcp",
		mcptransport.WithHTTPHeaders(map[string]string{"Authorization": "Bearer " + token}),
	)
	if err != nil {
		t.Fatalf("NewStreamableHttpClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mockserver-router-test", Version: "1.0.0"}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}
	if _, err := client.Initialize(ctx, initReq); err != nil {
		t.Fatalf("Initialize() through router error = %v", err)
	}

	listResp, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() through router error = %v", err)
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
}
