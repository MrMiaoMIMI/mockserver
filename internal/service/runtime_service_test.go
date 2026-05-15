package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

func TestMatchPublishedForwardFallbackExecutionIsHTTPRuntimeOnly(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	runtimeService := NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	event := bo.Event{
		Protocol:  "spex",
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "Call",
			"path":   "/service.Method",
		},
	}

	_, err := runtimeService.MatchPublished(context.Background(), event)
	if err == nil {
		t.Fatalf("expected non-HTTP runtime forward execution to fail")
	}
	if !strings.Contains(err.Error(), "HTTP runtime") {
		t.Fatalf("expected HTTP runtime boundary error, got %v", err)
	}

	decision, err := runtimeService.DecidePublished(context.Background(), event)
	if err != nil {
		t.Fatalf("DecidePublished() error = %v", err)
	}
	if decision.Kind != eo.DecisionKindForward {
		t.Fatalf("expected forward decision for protocol-specific mockinject handling, got %+v", decision)
	}
}

func TestDecidePublishedRuleMissKeepsWinnerRuleSetDiagnostics(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	runtimeService := NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	_, err := ruleSetRepository.publishSnapshot(bo.RuleSet{
		ID:        "http-api",
		Name:      "http api",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector: bo.Selector{All: []bo.Condition{
			{Field: "request.path", Op: eo.OperatorPrefix, Value: "/api/"},
		}},
		Rules: []bo.Rule{{
			ID:       "other-path",
			Name:     "Other path",
			Enabled:  true,
			Priority: 10,
			When:     bo.Condition{Field: "request.query.foo", Op: eo.OperatorEQ, Value: "bar"},
			Action: bo.Action{
				Type:     eo.ActionTypeRespond,
				Renderer: eo.ActionRendererStatic,
				Response: &bo.ProtocolResponse{
					Protocol: eo.ProtocolHTTP,
					Payload:  map[string]any{"status": 200},
				},
			},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("publish snapshot: %v", err)
	}

	decision, err := runtimeService.DecidePublished(context.Background(), bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "path": "/api/order"},
	})
	if err != nil {
		t.Fatalf("DecidePublished() error = %v", err)
	}
	if !decision.Fallback || decision.Trace.FallbackReason != eo.FallbackReasonRuleMiss {
		t.Fatalf("expected rule_miss fallback decision, got %+v", decision)
	}
	if decision.Trace.RulesetID != "http-api" {
		t.Fatalf("expected winner ruleset in trace, got %+v", decision.Trace)
	}
	if decision.Diagnostics == nil || decision.Diagnostics.RuleSetSelection.WinnerRuleSetID != "http-api" {
		t.Fatalf("expected winner ruleset diagnostics, got %+v", decision.Diagnostics)
	}
	if len(decision.Diagnostics.RuleSelection.CandidateRuleIDs) == 0 {
		t.Fatalf("expected candidate rule diagnostics, got %+v", decision.Diagnostics.RuleSelection)
	}
	raw, err := json.Marshal(decision)
	if err != nil {
		t.Fatalf("marshal decision: %v", err)
	}
	if strings.Contains(string(raw), "ruleset_selection") || strings.Contains(string(raw), "rule_selection") {
		t.Fatalf("SDK decision JSON must not expose internal diagnostics: %s", raw)
	}
}
