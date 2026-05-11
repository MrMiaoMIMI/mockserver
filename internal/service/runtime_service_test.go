package service

import (
	"context"
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
