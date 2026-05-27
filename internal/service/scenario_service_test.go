package service

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

type testScenarioRepository struct {
	mu        sync.RWMutex
	scenarios map[string]bo.Scenario
	rules     map[string]map[string]bo.ScenarioRule
}

func newTestScenarioRepository() *testScenarioRepository {
	return &testScenarioRepository{
		scenarios: make(map[string]bo.Scenario),
		rules:     make(map[string]map[string]bo.ScenarioRule),
	}
}

func (r *testScenarioRepository) CreateScenario(ctx context.Context, scenario bo.Scenario) (bo.Scenario, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.scenarios[scenario.ID]; exists {
		return bo.Scenario{}, dao.ErrConflict
	}
	scenario.DBID = uint64(len(r.scenarios) + 1)
	r.scenarios[scenario.ID] = scenario
	return scenario, nil
}

func (r *testScenarioRepository) GetScenario(ctx context.Context, id string) (bo.Scenario, bool, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	scenario, ok := r.scenarios[id]
	return scenario, ok, nil
}

func (r *testScenarioRepository) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]bo.Scenario, error) {
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
	if query.Offset > 0 {
		if query.Offset >= len(items) {
			return nil, nil
		}
		items = items[query.Offset:]
	}
	if query.Limit > 0 && query.Limit < len(items) {
		items = items[:query.Limit]
	}
	return items, nil
}

func (r *testScenarioRepository) UpdateScenario(ctx context.Context, scenario bo.Scenario, expectedVersion int) (bo.Scenario, error) {
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

func (r *testScenarioRepository) DeleteScenario(ctx context.Context, id string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.scenarios, id)
	delete(r.rules, id)
	return nil
}

func (r *testScenarioRepository) UpsertScenarioRule(ctx context.Context, rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.scenarios[rule.ScenarioID]; !ok {
		return bo.ScenarioRule{}, dao.ErrNotFound
	}
	if r.rules[rule.ScenarioID] == nil {
		r.rules[rule.ScenarioID] = map[string]bo.ScenarioRule{}
	}
	current, exists := r.rules[rule.ScenarioID][rule.RuleID]
	if exists {
		rule.Version = current.Version + 1
	} else {
		rule.Version = 1
	}
	r.rules[rule.ScenarioID][rule.RuleID] = rule
	return rule, nil
}

func (r *testScenarioRepository) ListScenarioRules(ctx context.Context, query bo.ScenarioRuleQuery) ([]bo.ScenarioRule, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]bo.ScenarioRule, 0)
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

func (r *testScenarioRepository) DeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules[scenarioID], ruleID)
	return nil
}

func TestScenarioServiceCreateValidatesScenarioID(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	if _, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "bad scenario"}); err == nil {
		t.Fatalf("expected invalid scenario id error")
	}
	created, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "scn_order_refund_ai_20260526"})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	if created.ID != "scn_order_refund_ai_20260526" || created.Status != bo.ScenarioStatusActive {
		t.Fatalf("unexpected scenario: %+v", created)
	}
	now := uint64(time.Now().UnixMilli())
	if created.ExpireTime <= now || created.ExpireTime > now+2*defaultScenarioTTLSeconds*millisecondsPerSecond {
		t.Fatalf("expected default expire time in Unix milliseconds, got %d", created.ExpireTime)
	}
}

func TestScenarioServiceCreateRejectsDuplicateScenarioID(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	if _, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "scn_duplicate_case1"}); err != nil {
		t.Fatalf("CreateScenario(first) error = %v", err)
	}
	if _, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "scn_duplicate_case1"}); err == nil {
		t.Fatalf("expected duplicate scenario error")
	}
}

func TestScenarioServiceListsAndUpdatesScenario(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	expireTime := uint64(time.Now().Add(time.Hour).UnixMilli())
	created, err := service.CreateScenario(context.Background(), bo.Scenario{
		ID:         "scn_update_case1",
		ExpireTime: expireTime,
	})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	if _, err := service.UpsertHTTPQuickRule(context.Background(), created.ID, bo.HTTPQuickRule{
		RuleID:    "ttl-rule",
		Namespace: "default",
		Match:     bo.HTTPQuickMatch{Path: "/ttl"},
		Respond:   bo.HTTPQuickResponse{Status: 200},
	}); err != nil {
		t.Fatalf("UpsertHTTPQuickRule() error = %v", err)
	}
	name := "Checkout timeout"
	description := "agent-owned checkout test"
	newExpireTime := uint64(time.Now().Add(2 * time.Hour).UnixMilli())
	updated, err := service.UpdateScenario(context.Background(), created.ID, bo.ScenarioUpdate{
		Name:        &name,
		Description: &description,
		ExpireTime:  &newExpireTime,
	})
	if err != nil {
		t.Fatalf("UpdateScenario() error = %v", err)
	}
	if updated.Name != name || updated.Description != description || updated.ExpireTime != newExpireTime || updated.Version != created.Version+1 {
		t.Fatalf("unexpected updated scenario: %+v", updated)
	}

	items, err := service.ListScenarios(context.Background(), bo.ScenarioQuery{IncludeExpired: false})
	if err != nil {
		t.Fatalf("ListScenarios() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("unexpected scenarios: %+v", items)
	}
	rules, err := service.ListScenarioRules(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("ListScenarioRules() error = %v", err)
	}
	if len(rules) != 1 || rules[0].ExpireTime != newExpireTime {
		t.Fatalf("expected rule TTL to follow scenario TTL, got %+v", rules)
	}
}

func TestScenarioServiceQuickHTTPRuleSimulates(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	created, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "scn_http_quick_case1"})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	_, err = service.UpsertHTTPQuickRule(context.Background(), created.ID, bo.HTTPQuickRule{
		RuleID:    "user-vip",
		Namespace: "default",
		Match: bo.HTTPQuickMatch{
			Method: "GET",
			Host:   "user.internal",
			Path:   "/v1/users/10001",
		},
		Respond: bo.HTTPQuickResponse{
			Status: 200,
			Body:   map[string]any{"user_id": 10001, "level": "vip"},
		},
	})
	if err != nil {
		t.Fatalf("UpsertHTTPQuickRule() error = %v", err)
	}
	result, err := service.SimulateScenario(context.Background(), created.ID, bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"host":   "user.internal",
			"path":   "/v1/users/10001",
		},
	}, false, 0, false, false)
	if err != nil {
		t.Fatalf("SimulateScenario() error = %v", err)
	}
	if !result.Matched || result.Trace.RuleID != "user-vip" {
		t.Fatalf("expected quick rule hit, got %+v", result)
	}
}

func TestScenarioServiceRulesMatchAcrossEventNamespaces(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	created, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "scn_namespace_agnostic"})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	if _, err := service.UpsertScenarioRule(context.Background(), bo.ScenarioRule{
		ScenarioID: created.ID,
		RuleID:     "spex-order",
		Name:       "SPEX order",
		Protocol:   eo.ProtocolSPEX,
		Namespace:  "legacy-ignored",
		Enabled:    true,
		Rule: bo.Rule{
			ID:      "spex-order",
			Name:    "SPEX order",
			Enabled: true,
			When:    bo.Condition{Field: "request.cmd", Op: eo.OperatorEQ, Value: "order.Get"},
			Action: bo.Action{
				Type:     eo.ActionTypeRespond,
				Renderer: eo.ActionRendererStatic,
				Response: &bo.ProtocolResponse{
					Protocol: eo.ProtocolSPEX,
					Payload:  map[string]any{"ok": true},
				},
			},
		},
	}); err != nil {
		t.Fatalf("UpsertScenarioRule() error = %v", err)
	}

	for _, namespace := range []string{"default", "shop-sg"} {
		result, err := service.SimulateScenario(context.Background(), created.ID, bo.Event{
			Protocol:  eo.ProtocolSPEX,
			Namespace: namespace,
			Request:   bo.EventRequest{"cmd": "order.Get"},
		}, false, 0, false, false)
		if err != nil {
			t.Fatalf("SimulateScenario(%s) error = %v", namespace, err)
		}
		if !result.Matched || result.Trace.RuleID != "spex-order" {
			t.Fatalf("expected scenario rule hit for namespace %s, got %+v", namespace, result)
		}
	}
}

func TestScenarioServiceListsAndDeletesScenarioRule(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	created, err := service.CreateScenario(context.Background(), bo.Scenario{ID: "scn_rule_list_case1"})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	if _, err := service.UpsertHTTPQuickRule(context.Background(), created.ID, bo.HTTPQuickRule{
		RuleID:    "rule-one",
		Namespace: "default",
		Match:     bo.HTTPQuickMatch{Path: "/one"},
		Respond:   bo.HTTPQuickResponse{Status: 200},
	}); err != nil {
		t.Fatalf("UpsertHTTPQuickRule() error = %v", err)
	}
	rules, err := service.ListScenarioRules(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("ListScenarioRules() error = %v", err)
	}
	if len(rules) != 1 || rules[0].RuleID != "rule-one" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
	if err := service.DeleteScenarioRule(context.Background(), created.ID, "rule-one"); err != nil {
		t.Fatalf("DeleteScenarioRule() error = %v", err)
	}
	rules, err = service.ListScenarioRules(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("ListScenarioRules(after delete) error = %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected rules to be deleted, got %+v", rules)
	}
}

func TestRuntimeServiceScenarioOverlayFallsBackOnRuleMiss(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	scenarioService := NewScenarioService(newTestScenarioRepository())
	runtimeService := NewRuntimeServiceWithScenarios(ruleSetRepository, ruleSetRepository, namespaceService, scenarioService)

	_, err := scenarioService.CreateScenario(context.Background(), bo.Scenario{
		ID:         "scn_overlay_case1",
		ExpireTime: uint64(time.Now().Add(time.Hour).UnixMilli()),
	})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	if _, err := scenarioService.UpsertHTTPQuickRule(context.Background(), "scn_overlay_case1", bo.HTTPQuickRule{
		RuleID:    "scenario-user",
		Namespace: "default",
		Match:     bo.HTTPQuickMatch{Method: "GET", Host: "user.internal", Path: "/scenario"},
		Respond:   bo.HTTPQuickResponse{Status: 209, Body: map[string]any{"source": "scenario"}},
	}); err != nil {
		t.Fatalf("UpsertHTTPQuickRule() error = %v", err)
	}
	_, err = ruleSetRepository.publishSnapshot(bo.RuleSet{
		ID:        "published-api",
		Name:      "published api",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  bo.Selector{All: []bo.Condition{{Field: "request.path", Op: eo.OperatorPrefix, Value: "/"}}},
		Rules: []bo.Rule{{
			ID:       "published-user",
			Name:     "Published user",
			Enabled:  true,
			Priority: 10,
			When:     bo.Condition{Field: "request.path", Op: eo.OperatorEQ, Value: "/published"},
			Action: bo.Action{
				Type:     eo.ActionTypeRespond,
				Renderer: eo.ActionRendererStatic,
				Response: &bo.ProtocolResponse{Protocol: eo.ProtocolHTTP, Payload: map[string]any{"status": 208}},
			},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("publish snapshot: %v", err)
	}

	overlayDecision, err := runtimeService.DecidePublished(context.Background(), bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "host": "user.internal", "path": "/scenario"},
		Meta:      bo.EventMeta{ScenarioID: "scn_overlay_case1"},
	})
	if err != nil {
		t.Fatalf("DecidePublished(overlay) error = %v", err)
	}
	if !overlayDecision.Matched || overlayDecision.Trace.RuleID != "scenario-user" || overlayDecision.Meta.ScenarioID != "scn_overlay_case1" {
		t.Fatalf("expected scenario overlay hit, got %+v", overlayDecision)
	}

	fallbackDecision, err := runtimeService.DecidePublished(context.Background(), bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "path": "/published"},
		Meta:      bo.EventMeta{ScenarioID: "scn_overlay_case1"},
	})
	if err != nil {
		t.Fatalf("DecidePublished(fallback) error = %v", err)
	}
	if !fallbackDecision.Fallback || fallbackDecision.Kind != eo.DecisionKindForward || fallbackDecision.Trace.RuleID == "published-user" {
		t.Fatalf("expected namespace fallback instead of published fallthrough, got %+v", fallbackDecision)
	}
}

func TestRuntimeServiceScenarioOverlayFallsBackWhenScenarioMissing(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	runtimeService := NewRuntimeServiceWithScenarios(ruleSetRepository, ruleSetRepository, namespaceService, NewScenarioService(newTestScenarioRepository()))

	decision, err := runtimeService.DecidePublished(context.Background(), bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "path": "/published"},
		Meta:      bo.EventMeta{ScenarioID: "scn_missing_case1"},
	})
	if err != nil {
		t.Fatalf("DecidePublished() error = %v", err)
	}
	if !decision.Fallback || decision.Kind != eo.DecisionKindForward {
		t.Fatalf("expected namespace forward fallback, got %+v", decision)
	}
}

func TestScenarioServiceRejectsExpiredScenarioRuleUpdates(t *testing.T) {
	service := NewScenarioService(newTestScenarioRepository())
	_, err := service.CreateScenario(context.Background(), bo.Scenario{
		ID:         "scn_expired_case1",
		ExpireTime: uint64(time.Now().Add(-time.Minute).UnixMilli()),
	})
	if err != nil {
		t.Fatalf("CreateScenario() error = %v", err)
	}
	_, err = service.UpsertHTTPQuickRule(context.Background(), "scn_expired_case1", bo.HTTPQuickRule{
		RuleID:    "rule",
		Namespace: "default",
		Match:     bo.HTTPQuickMatch{Path: "/x"},
		Respond:   bo.HTTPQuickResponse{Status: 200},
	})
	if err == nil {
		t.Fatalf("expected expired scenario error")
	}
	if errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected business validation error, got %v", err)
	}
}
