package service

import (
	"context"
	"strings"
	"testing"

	"github.com/MrMiaoMIMI/goshared/util/servererr"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

func TestNamespaceServiceEnforcesUniqueNameAndVersion(t *testing.T) {
	ctx := context.Background()
	repository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(repository)

	saved, err := namespaceService.UpsertNamespace(ctx, bo.Namespace{
		ID:   "tenant-a",
		Name: "Tenant A",
	})
	if err != nil {
		t.Fatalf("UpsertNamespace(create) error = %v", err)
	}
	if saved.Version != 1 {
		t.Fatalf("unexpected created version: %d", saved.Version)
	}

	if _, err := namespaceService.UpsertNamespace(ctx, bo.Namespace{
		ID:   "tenant-b",
		Name: " tenant   a ",
	}); servererr.CodeOf(err) != servererr.ErrConflict {
		t.Fatalf("expected duplicate namespace name conflict, got %v", err)
	}

	updated, err := namespaceService.UpsertNamespace(ctx, bo.Namespace{
		ID:      saved.ID,
		Name:    "Tenant A",
		Version: saved.Version,
	})
	if err != nil {
		t.Fatalf("UpsertNamespace(update) error = %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("unexpected updated version: %d", updated.Version)
	}

	if _, err := namespaceService.UpsertNamespace(ctx, bo.Namespace{
		ID:      saved.ID,
		Name:    "Tenant A stale",
		Version: saved.Version,
	}); servererr.CodeOf(err) != servererr.ErrConflict {
		t.Fatalf("expected stale namespace version conflict, got %v", err)
	}
}

func TestRuleSetServiceEnforcesUniqueNamesAndPublishSelectorConflicts(t *testing.T) {
	ctx := context.Background()
	repository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(repository)
	ruleSetService := NewRuleSetService(repository, namespaceService)

	first := serviceHTTPRuleSet("orders-a", "Orders", "/api/orders")
	if _, err := ruleSetService.UpsertDraft(ctx, first); err != nil {
		t.Fatalf("UpsertDraft(first) error = %v", err)
	}
	if _, err := ruleSetService.UpsertDraft(ctx, serviceHTTPRuleSet("orders-b", " orders ", "/api/other")); servererr.CodeOf(err) != servererr.ErrConflict {
		t.Fatalf("expected duplicate ruleset name conflict, got %v", err)
	}
	if _, err := namespaceService.UpsertNamespace(ctx, bo.Namespace{ID: "tenant-b", Name: "Tenant B"}); err != nil {
		t.Fatalf("UpsertNamespace(tenant-b) error = %v", err)
	}
	scopedDuplicateName := serviceHTTPRuleSet("orders-c", " orders ", "/api/tenant-b")
	scopedDuplicateName.Namespace = "tenant-b"
	if _, err := ruleSetService.UpsertDraft(ctx, scopedDuplicateName); err != nil {
		t.Fatalf("expected same ruleset name to be allowed in another namespace, got %v", err)
	}
	if _, err := ruleSetService.Publish(ctx, first.ID, bo.AuditInfo{Operator: "test"}); err != nil {
		t.Fatalf("Publish(first) error = %v", err)
	}

	conflicting := serviceHTTPRuleSet("orders-conflict", "Orders conflict", "/api/orders")
	if _, err := ruleSetService.UpsertDraft(ctx, conflicting); err != nil {
		t.Fatalf("UpsertDraft(conflicting) error = %v", err)
	}
	if _, err := ruleSetService.Publish(ctx, conflicting.ID, bo.AuditInfo{Operator: "test"}); servererr.CodeOf(err) != servererr.ErrConflict {
		t.Fatalf("expected published selector conflict, got %v", err)
	}
}

func TestRuntimeServiceCachesCompiledPublishedRuleSetsUntilRevisionChanges(t *testing.T) {
	ctx := context.Background()
	repository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(repository)
	ruleSetService := NewRuleSetService(repository, namespaceService)
	runtimeService := NewRuntimeService(repository, repository, namespaceService)

	ruleSet := serviceHTTPRuleSet("cache-hit", "Cache hit", "/api/cache")
	if _, err := ruleSetService.UpsertDraft(ctx, ruleSet); err != nil {
		t.Fatalf("UpsertDraft() error = %v", err)
	}
	if _, err := ruleSetService.Publish(ctx, ruleSet.ID, bo.AuditInfo{Operator: "test"}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	before := repository.ListPublishedCallCount()
	for i := 0; i < 2; i++ {
		decision, err := runtimeService.DecidePublished(ctx, bo.Event{
			Protocol:  eo.ProtocolHTTP,
			Namespace: "default",
			Request: map[string]any{
				"method": "GET",
				"path":   "/api/cache/hit",
			},
		})
		if err != nil {
			t.Fatalf("DecidePublished(%d) error = %v", i, err)
		}
		if !decision.Matched || decision.Trace.RuleID != "hit" {
			t.Fatalf("unexpected decision: %#v", decision)
		}
	}
	after := repository.ListPublishedCallCount()
	if after != before+1 {
		t.Fatalf("expected one published load across two runtime decisions, before=%d after=%d", before, after)
	}
}

func serviceHTTPRuleSet(id string, name string, selectorPrefix string) bo.RuleSet {
	return bo.RuleSet{
		ID:        id,
		Name:      name,
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector: bo.Selector{All: []bo.Condition{{
			Field: "request.path",
			Op:    eo.OperatorPrefix,
			Value: selectorPrefix,
		}}},
		Rules: []bo.Rule{{
			ID:       "hit",
			Name:     "Hit",
			Enabled:  true,
			Priority: 100,
			When: bo.Condition{
				Field: "request.path",
				Op:    eo.OperatorPrefix,
				Value: strings.TrimRight(selectorPrefix, "/") + "/hit",
			},
			Action: bo.Action{
				Type:     eo.ActionTypeRespond,
				Renderer: eo.ActionRendererStatic,
				Response: &bo.ProtocolResponse{
					Protocol: eo.ProtocolHTTP,
					Payload: map[string]any{
						"status": 200,
						"body": map[string]any{
							"ok": true,
						},
					},
				},
			},
		}},
	}
}
