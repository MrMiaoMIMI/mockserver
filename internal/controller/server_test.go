package controller_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/controller"
	"github.com/MrMiaoMIMI/mockserver/internal/observability"
	"github.com/MrMiaoMIMI/mockserver/internal/router"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
	"github.com/MrMiaoMIMI/mockserver/internal/view"
)

func TestAdminPublishAndRuntimeFlow(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)
	runtimeMetrics := observability.NewRuntimeMetrics()

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, runtimeMetrics)
	metricsController := controller.NewMetricsController(runtimeMetrics)
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, metricsController)

	upsertBody := map[string]any{
		"name":      "http default",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "debug-api",
				"name":     "Debug Api",
				"enabled":  true,
				"priority": 100,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.method", "op": "eq", "value": "GET"},
						{"field": "request.path", "op": "eq", "value": "/api/v1/debug"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body": map[string]any{
						"code": 0,
					},
				},
			},
		},
	}
	nonMatchingRuleSet := map[string]any{
		"id":        "http-other-host",
		"name":      "http other host",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpHostSelectorBody("other.example.com"),
		"rules": []map[string]any{
			{
				"id":       "other-host-rule",
				"name":     "Other Host Rule",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.path", "op": "eq", "value": "/api/v1/debug"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"code": 1},
				},
			},
		},
	}

	upsertResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", upsertBody, http.StatusOK)
	var upsertEnvelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(readBody(t, upsertResp), &upsertEnvelope); err != nil {
		t.Fatalf("decode upsert response: %v", err)
	}
	ruleSetID := upsertEnvelope.Data.ID
	if ruleSetID == "" {
		t.Fatalf("expected generated ruleset id")
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", nonMatchingRuleSet, http.StatusOK)
	doJSON(t, handler, http.MethodPost, fmt.Sprintf("/mockserver/api/v1/admin/rulesets/%s/validate", ruleSetID), nil, http.StatusOK)
	publishResp := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/mockserver/api/v1/admin/rulesets/%s/publish", ruleSetID), nil, http.StatusOK)
	assertBytesContain(t, readBody(t, publishResp), fmt.Sprintf(`"snapshot_id":"%s-v1-`, ruleSetID))

	simulateResp := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/mockserver/api/v1/admin/rulesets/%s/simulate", ruleSetID), map[string]any{
		"explain_only":    true,
		"explain_compact": true,
		"explain_summary": true,
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"path":   "/api/v1/debug",
			},
		},
	}, http.StatusOK)
	simulateBody := readBody(t, simulateResp)
	assertBytesContain(t, simulateBody, `"rule_explanations"`)
	assertBytesContain(t, simulateBody, `"action_info"`)
	assertBytesContain(t, simulateBody, `"rule_set_explanations"`)
	assertBytesContain(t, simulateBody, `"action execution skipped because explain_only=true"`)
	assertBytesContain(t, simulateBody, `"response":{"status":0}`)
	assertBytesNotContain(t, simulateBody, `"expected"`)
	assertBytesNotContain(t, simulateBody, `"actual"`)
	assertBytesNotContain(t, simulateBody, `"rendered_result"`)
	assertBytesNotContain(t, simulateBody, `"selector_checks"`)
	assertBytesNotContain(t, simulateBody, `"candidate_rules"`)
	assertBytesNotContain(t, simulateBody, `"children"`)

	overrideResp := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/mockserver/api/v1/admin/rulesets/%s/simulate", ruleSetID), map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "POST",
				"path":   "/api/v1/debug",
			},
		},
		"draft_override": map[string]any{
			"id":        ruleSetID,
			"name":      "http default",
			"enabled":   true,
			"protocol":  "http",
			"namespace": "default",
			"selector":  httpPathSelectorBody("/api/"),
			"rules": []map[string]any{
				{
					"id":       "override-rule",
					"name":     "Override Rule",
					"enabled":  true,
					"priority": 10,
					"when": map[string]any{
						"all": []map[string]any{
							{"field": "request.method", "op": "eq", "value": "POST"},
							{"field": "request.path", "op": "eq", "value": "/api/v1/debug"},
						},
					},
					"action": map[string]any{
						"type":   "static_response",
						"status": 209,
						"body":   map[string]any{"override": true},
					},
				},
			},
		},
	}, http.StatusOK)
	overrideBody := readBody(t, overrideResp)
	assertBytesContain(t, overrideBody, `"rule_id":"override-rule"`)
	assertBytesContain(t, overrideBody, `"status":209`)

	selectorExplainResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/http-other-host/simulate", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"host":   "demo.com",
				"path":   "/api/v1/debug",
			},
		},
	}, http.StatusOK)
	selectorExplainBody := readBody(t, selectorExplainResp)
	assertBytesContain(t, selectorExplainBody, `"ruleset_id":"http-other-host"`)
	assertBytesContain(t, selectorExplainBody, `"field request.host did not match operator eq"`)

	publishedResp := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/mockserver/api/v1/admin/published/rulesets/%s", ruleSetID), nil, http.StatusOK)
	assertBytesContain(t, readBody(t, publishedResp), fmt.Sprintf(`"id":"%s"`, ruleSetID))

	publishedSimulateResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/published/simulate", map[string]any{
		"explain_summary": true,
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"path":   "/api/v1/debug",
			},
		},
	}, http.StatusOK)
	publishedSimulateBody := readBody(t, publishedSimulateResp)
	assertBytesContain(t, publishedSimulateBody, `"matched":true`)
	assertBytesContain(t, publishedSimulateBody, fmt.Sprintf(`"ruleset_id":"%s"`, ruleSetID))
	assertBytesContain(t, publishedSimulateBody, `"rule_id":"debug-api"`)

	snapshotsResp := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/mockserver/api/v1/admin/published/rulesets/%s/snapshots", ruleSetID), nil, http.StatusOK)
	assertBytesContain(t, readBody(t, snapshotsResp), `"total":1`)

	resp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/default/http/api/v1/debug", nil, http.StatusOK)
	if got := resp.Header.Get("X-Mockserver-Rule"); got != "debug-api" {
		t.Fatalf("unexpected matched rule header: %s", got)
	}
	if traceID := resp.Header.Get("X-Trace-ID"); traceID == "" || traceID == "trace-auto-generated" {
		t.Fatalf("unexpected generated trace id: %q", traceID)
	}

	metricsResp := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/metrics/runtime", nil, http.StatusOK)
	metricsBody := readBody(t, metricsResp)
	assertBytesContain(t, metricsBody, `"total_requests":1`)
	assertBytesContain(t, metricsBody, `"matched_requests":1`)
	assertBytesContain(t, metricsBody, `"debug-api":1`)
	assertBytesContain(t, metricsBody, `"recent_requests"`)
	assertBytesContain(t, metricsBody, `"outcome":"matched"`)
	assertBytesContain(t, metricsBody, `"path":"/api/v1/debug"`)
	assertBytesContain(t, metricsBody, `"status":200`)
}

func TestSDKDecisionEndpointReturnsPublishedHitDecision(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)
	runtimeMetrics := observability.NewRuntimeMetrics()

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, runtimeMetrics)
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	ruleSetBody := map[string]any{
		"id":        "sdk-decision-ruleset",
		"name":      "sdk decision ruleset",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/sdk/"),
		"rules": []map[string]any{
			{
				"id":       "sdk-hit-rule",
				"name":     "Sdk Hit Rule",
				"enabled":  true,
				"priority": 100,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.method", "op": "eq", "value": "GET"},
						{"field": "request.path", "op": "eq", "value": "/api/sdk/hit"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 202,
					"headers": map[string]any{
						"x-sdk-decision": []string{"hit"},
					},
					"body": map[string]any{
						"source": "sdk",
						"hit":    true,
					},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", ruleSetBody, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/sdk-decision-ruleset/publish", nil, http.StatusOK)

	decisionResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"host":   "demo.com",
				"path":   "/api/sdk/hit",
			},
			"meta": map[string]any{
				"trace_id": "sdk-trace-001",
			},
		},
	}, http.StatusOK)
	decisionBody := readBody(t, decisionResp)
	assertBytesContain(t, decisionBody, `"decision":{"kind":"response"`)
	assertBytesContain(t, decisionBody, `"matched":true`)
	assertBytesContain(t, decisionBody, `"ruleset_id":"sdk-decision-ruleset"`)
	assertBytesContain(t, decisionBody, `"rule_id":"sdk-hit-rule"`)
	assertBytesContain(t, decisionBody, `"trace_id":"sdk-trace-001"`)
	assertBytesContain(t, decisionBody, `"status":202`)
	assertBytesContain(t, decisionBody, `"x-sdk-decision":["hit"]`)
	assertBytesContain(t, decisionBody, `"source":"sdk"`)

	runtimeResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/default/http/api/sdk/hit", nil, http.StatusAccepted)
	if got := runtimeResp.Header.Get("X-Mockserver-Rule"); got != "sdk-hit-rule" {
		t.Fatalf("unexpected runtime matched rule header: %q", got)
	}
	assertBytesContain(t, readBody(t, runtimeResp), `"source":"sdk"`)
}

func TestSDKDecisionEndpointReturnsForwardDecisionsAndDoesNotForward(t *testing.T) {
	var forwardedRequests int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&forwardedRequests, 1)
		w.Header().Set("X-Upstream", "forwarded")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"forwarded":true}`))
	}))
	defer upstream.Close()
	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	rulesetMissDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method":        "POST",
				"scheme":        "http",
				"host":          "demo.com",
				"original_host": upstreamHost,
				"path":          "/api/pass-through",
			},
			"meta": map[string]any{
				"trace_id": "sdk-forward-ruleset",
			},
		},
	}, http.StatusOK)
	rulesetMissBody := readBody(t, rulesetMissDecision)
	assertBytesContain(t, rulesetMissBody, `"decision":{"kind":"forward"`)
	assertBytesContain(t, rulesetMissBody, `"matched":false`)
	assertBytesContain(t, rulesetMissBody, `"fallback":true`)
	assertBytesContain(t, rulesetMissBody, `"fallback_reason":"ruleset_miss"`)
	assertBytesContain(t, rulesetMissBody, `"timeout_ms":5000`)
	assertBytesContain(t, rulesetMissBody, `"trace_id":"sdk-forward-ruleset"`)
	if got := atomic.LoadInt32(&forwardedRequests); got != 0 {
		t.Fatalf("sdk decision endpoint should not forward upstream requests, got %d", got)
	}

	runtimeRulesetMiss := doJSONWithHeaders(
		t,
		handler,
		http.MethodPost,
		"/mockserver/runtime/default/http/api/pass-through",
		map[string]any{"request": "original"},
		map[string]string{
			"X-Forwarded-Host":  upstreamHost,
			"X-Forwarded-Proto": "http",
		},
		http.StatusAccepted,
	)
	if got := runtimeRulesetMiss.Header.Get("X-Mockserver-Fallback"); got != "ruleset_miss" {
		t.Fatalf("unexpected runtime fallback header: %q", got)
	}
	if got := runtimeRulesetMiss.Header.Get("X-Upstream"); got != "forwarded" {
		t.Fatalf("unexpected upstream header: %q", got)
	}
	if got := atomic.LoadInt32(&forwardedRequests); got != 1 {
		t.Fatalf("runtime endpoint should forward exactly once, got %d", got)
	}

	ruleSetBody := map[string]any{
		"id":        "sdk-forward-rule-miss",
		"name":      "sdk forward rule miss",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "only-hit",
				"name":     "Only Hit",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"field": "request.path",
					"op":    "eq",
					"value": "/api/hit",
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"mocked": true},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", ruleSetBody, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/sdk-forward-rule-miss/publish", nil, http.StatusOK)

	ruleMissDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method":        "GET",
				"scheme":        "http",
				"host":          "demo.com",
				"original_host": upstreamHost,
				"path":          "/api/miss",
			},
			"meta": map[string]any{
				"trace_id": "sdk-forward-rule",
			},
		},
	}, http.StatusOK)
	ruleMissBody := readBody(t, ruleMissDecision)
	assertBytesContain(t, ruleMissBody, `"decision":{"kind":"forward"`)
	assertBytesContain(t, ruleMissBody, `"fallback_reason":"rule_miss"`)
	assertBytesContain(t, ruleMissBody, `"timeout_ms":5000`)
	assertBytesContain(t, ruleMissBody, `"trace_id":"sdk-forward-rule"`)
	if got := atomic.LoadInt32(&forwardedRequests); got != 1 {
		t.Fatalf("sdk rule miss decision should not forward upstream requests, got %d", got)
	}
}

func TestNamespaceFallbackResponseFlow(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	namespaceBody := map[string]any{
		"name": "fallback namespace",
		"ruleset_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 418,
				"body":   map[string]any{"fallback": "ruleset"},
			},
		},
		"rule_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 409,
				"body":   map[string]any{"fallback": "rule"},
			},
		},
	}
	namespaceResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/namespaces", namespaceBody, http.StatusOK)
	var namespaceEnvelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(readBody(t, namespaceResp), &namespaceEnvelope); err != nil {
		t.Fatalf("decode namespace response: %v", err)
	}
	namespaceID := namespaceEnvelope.Data.ID
	if namespaceID == "" {
		t.Fatalf("expected generated namespace id")
	}

	rulesetMissResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/"+namespaceID+"/http/no-ruleset", nil, http.StatusTeapot)
	if got := rulesetMissResp.Header.Get("X-Mockserver-Fallback"); got != "ruleset_miss" {
		t.Fatalf("unexpected ruleset fallback header: %q", got)
	}
	assertBytesContain(t, readBody(t, rulesetMissResp), `"fallback":"ruleset"`)

	ruleSetBody := map[string]any{
		"id":        "fallback-ruleset",
		"name":      "fallback ruleset",
		"enabled":   true,
		"protocol":  "http",
		"namespace": namespaceID,
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "only-hit",
				"name":     "Only Hit",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"field": "request.path",
					"op":    "eq",
					"value": "/api/hit",
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"hit": true},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", ruleSetBody, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/fallback-ruleset/publish", nil, http.StatusOK)

	ruleMissResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/"+namespaceID+"/http/api/miss", nil, http.StatusConflict)
	if got := ruleMissResp.Header.Get("X-Mockserver-Fallback"); got != "rule_miss" {
		t.Fatalf("unexpected rule fallback header: %q", got)
	}
	assertBytesContain(t, readBody(t, ruleMissResp), `"fallback":"rule"`)
}

func TestSDKDecisionEndpointReturnsResponseFallbackDecisions(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	namespaceResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/namespaces", map[string]any{
		"name": "sdk response fallback",
		"ruleset_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 418,
				"headers": map[string]any{
					"x-sdk-fallback": []string{"ruleset"},
				},
				"body": map[string]any{"fallback": "ruleset"},
			},
		},
		"rule_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 409,
				"headers": map[string]any{
					"x-sdk-fallback": []string{"rule"},
				},
				"body": map[string]any{"fallback": "rule"},
			},
		},
	}, http.StatusOK)
	var namespaceEnvelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(readBody(t, namespaceResp), &namespaceEnvelope); err != nil {
		t.Fatalf("decode namespace response: %v", err)
	}
	namespaceID := namespaceEnvelope.Data.ID

	rulesetMissDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": namespaceID,
			"request": map[string]any{
				"method": "GET",
				"path":   "/no-ruleset",
			},
			"meta": map[string]any{
				"trace_id": "sdk-response-ruleset",
			},
		},
	}, http.StatusOK)
	rulesetMissBody := readBody(t, rulesetMissDecision)
	assertBytesContain(t, rulesetMissBody, `"decision":{"kind":"response"`)
	assertBytesContain(t, rulesetMissBody, `"matched":false`)
	assertBytesContain(t, rulesetMissBody, `"fallback":true`)
	assertBytesContain(t, rulesetMissBody, `"fallback_reason":"ruleset_miss"`)
	assertBytesContain(t, rulesetMissBody, `"status":418`)
	assertBytesContain(t, rulesetMissBody, `"x-sdk-fallback":["ruleset"]`)
	assertBytesContain(t, rulesetMissBody, `"fallback":"ruleset"`)
	assertBytesContain(t, rulesetMissBody, `"trace_id":"sdk-response-ruleset"`)

	ruleSetBody := map[string]any{
		"id":        "sdk-response-rule-miss",
		"name":      "sdk response rule miss",
		"enabled":   true,
		"protocol":  "http",
		"namespace": namespaceID,
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "only-hit",
				"name":     "Only Hit",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"field": "request.path",
					"op":    "eq",
					"value": "/api/hit",
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"mocked": true},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", ruleSetBody, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/sdk-response-rule-miss/publish", nil, http.StatusOK)

	ruleMissDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": namespaceID,
			"request": map[string]any{
				"method": "GET",
				"path":   "/api/miss",
			},
			"meta": map[string]any{
				"trace_id": "sdk-response-rule",
			},
		},
	}, http.StatusOK)
	ruleMissBody := readBody(t, ruleMissDecision)
	assertBytesContain(t, ruleMissBody, `"decision":{"kind":"response"`)
	assertBytesContain(t, ruleMissBody, `"fallback_reason":"rule_miss"`)
	assertBytesContain(t, ruleMissBody, `"status":409`)
	assertBytesContain(t, ruleMissBody, `"x-sdk-fallback":["rule"]`)
	assertBytesContain(t, ruleMissBody, `"fallback":"rule"`)

	runtimeRuleMiss := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/"+namespaceID+"/http/api/miss", nil, http.StatusConflict)
	if got := runtimeRuleMiss.Header.Get("X-Mockserver-Fallback"); got != "rule_miss" {
		t.Fatalf("unexpected runtime fallback header: %q", got)
	}
	assertBytesContain(t, readBody(t, runtimeRuleMiss), `"fallback":"rule"`)
}

func TestSDKDecisionEndpointEndToEndDecisions(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", map[string]any{
		"id":        "sdk-e2e",
		"name":      "sdk e2e",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/sdk/"),
		"rules": []map[string]any{
			{
				"id":       "sdk-hit",
				"name":     "Sdk Hit",
				"enabled":  true,
				"priority": 100,
				"when": map[string]any{
					"field": "request.path",
					"op":    "eq",
					"value": "/sdk/hit",
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"decision": "hit"},
				},
			},
		},
	}, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/sdk-e2e/publish", nil, http.StatusOK)

	namespaceResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/namespaces", map[string]any{
		"name": "sdk e2e response fallback",
		"ruleset_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 451,
				"body":   map[string]any{"decision": "response-fallback"},
			},
		},
		"rule_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 452,
				"body":   map[string]any{"decision": "rule-response-fallback"},
			},
		},
	}, http.StatusOK)
	var namespaceEnvelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(readBody(t, namespaceResp), &namespaceEnvelope); err != nil {
		t.Fatalf("decode namespace response: %v", err)
	}

	hitDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"path":   "/sdk/hit",
			},
		},
	}, http.StatusOK)
	hitBody := readBody(t, hitDecision)
	assertBytesContain(t, hitBody, `"decision":{"kind":"response"`)
	assertBytesContain(t, hitBody, `"matched":true`)
	assertBytesContain(t, hitBody, `"status":200`)
	assertBytesContain(t, hitBody, `"decision":"hit"`)

	forwardDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"path":   "/sdk/miss",
			},
		},
	}, http.StatusOK)
	forwardBody := readBody(t, forwardDecision)
	assertBytesContain(t, forwardBody, `"decision":{"kind":"forward"`)
	assertBytesContain(t, forwardBody, `"fallback":true`)
	assertBytesContain(t, forwardBody, `"fallback_reason":"rule_miss"`)

	responseFallbackDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "http",
			"namespace": namespaceEnvelope.Data.ID,
			"request": map[string]any{
				"method": "GET",
				"path":   "/no-ruleset",
			},
		},
	}, http.StatusOK)
	responseFallbackBody := readBody(t, responseFallbackDecision)
	assertBytesContain(t, responseFallbackBody, `"decision":{"kind":"response"`)
	assertBytesContain(t, responseFallbackBody, `"fallback":true`)
	assertBytesContain(t, responseFallbackBody, `"status":451`)
	assertBytesContain(t, responseFallbackBody, `"decision":"response-fallback"`)
}

func TestAdminListProtocols(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	resp := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/protocols", nil, http.StatusOK)
	body := readBody(t, resp)
	assertBytesContain(t, body, `"name":"http"`)
	assertBytesContain(t, body, `"path":"request.body"`)
	assertBytesContain(t, body, `"name":"cache"`)
	assertBytesContain(t, body, `"path":"request.value"`)
	assertBytesContain(t, body, `"selectors"`)
}

func TestSDKDecisionEndpointCacheDecisionEndToEnd(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", map[string]any{
		"id":        "cache-sdk",
		"name":      "cache sdk",
		"enabled":   true,
		"protocol":  "cache",
		"namespace": "default",
		"selector":  cacheKeySelectorBody("get", "user:"),
		"rules": []map[string]any{
			{
				"id":       "cache-get-user",
				"name":     "Cache Get User",
				"enabled":  true,
				"priority": 100,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.operation", "op": "eq", "value": "get"},
						{"field": "request.key", "op": "eq", "value": "user:123"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"value": "mocked"},
				},
			},
		},
	}, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/cache-sdk/publish", nil, http.StatusOK)

	hitDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "cache",
			"operation": "request",
			"namespace": "default",
			"request": map[string]any{
				"operation": "get",
				"key":       "user:123",
			},
		},
	}, http.StatusOK)
	hitBody := readBody(t, hitDecision)
	assertBytesContain(t, hitBody, `"decision":{"kind":"response"`)
	assertBytesContain(t, hitBody, `"matched":true`)
	assertBytesContain(t, hitBody, `"value":"mocked"`)

	missDecision := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/sdk/decision", map[string]any{
		"event": map[string]any{
			"protocol":  "cache",
			"operation": "request",
			"namespace": "default",
			"request": map[string]any{
				"operation": "get",
				"key":       "user:456",
			},
		},
	}, http.StatusOK)
	missBody := readBody(t, missDecision)
	assertBytesContain(t, missBody, `"decision":{"kind":"forward"`)
	assertBytesContain(t, missBody, `"fallback":true`)
	assertBytesContain(t, missBody, `"fallback_reason":"rule_miss"`)
}

func TestNamespaceCreateDefaultsToForwardFallback(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	createResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/namespaces", map[string]any{
		"name":        "pass through",
		"description": "created without explicit fallback",
	}, http.StatusOK)
	createBody := readBody(t, createResp)
	assertBytesContain(t, createBody, `"name":"pass through"`)
	assertBytesContain(t, createBody, `"ruleset_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)
	assertBytesContain(t, createBody, `"rule_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)

	var envelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createBody, &envelope); err != nil {
		t.Fatalf("decode namespace create response: %v", err)
	}
	if envelope.Data.ID == "" {
		t.Fatalf("expected generated namespace id")
	}

	getResp := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/namespaces/"+envelope.Data.ID, nil, http.StatusOK)
	getBody := readBody(t, getResp)
	assertBytesContain(t, getBody, `"ruleset_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)
	assertBytesContain(t, getBody, `"rule_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)

	listResp := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/namespaces", nil, http.StatusOK)
	listBody := readBody(t, listResp)
	assertBytesContain(t, listBody, `"id":"`+envelope.Data.ID+`"`)
	assertBytesContain(t, listBody, `"ruleset_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)
	assertBytesContain(t, listBody, `"rule_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)

	updateResp := doJSON(t, handler, http.MethodPut, "/mockserver/api/v1/admin/namespaces/"+envelope.Data.ID, map[string]any{
		"name":        "strict namespace",
		"description": "explicit response fallback",
		"ruleset_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 418,
				"body":   map[string]any{"fallback": "ruleset"},
			},
		},
		"rule_miss_action": map[string]any{
			"type": "response",
			"response": map[string]any{
				"status": 409,
				"body":   map[string]any{"fallback": "rule"},
			},
		},
	}, http.StatusOK)
	updateBody := readBody(t, updateResp)
	assertBytesContain(t, updateBody, `"ruleset_miss_action":{"type":"response"`)
	assertBytesContain(t, updateBody, `"status":418`)
	assertBytesContain(t, updateBody, `"rule_miss_action":{"type":"response"`)
	assertBytesContain(t, updateBody, `"status":409`)
}

func TestDefaultNamespaceRulesetMissForwardsOriginalRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected forwarded method: %s", r.Method)
		}
		if r.URL.Path != "/api/pass-through" {
			t.Fatalf("unexpected forwarded path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("token") != "abc" {
			t.Fatalf("unexpected forwarded query: %s", r.URL.RawQuery)
		}
		if got := r.Header.Get("X-Client"); got != "mobile" {
			t.Fatalf("unexpected forwarded header: %q", got)
		}
		if got := r.Header.Get("Connection"); got != "" {
			t.Fatalf("hop-by-hop header should not be forwarded: %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read forwarded body: %v", err)
		}
		if !bytes.Contains(body, []byte(`"request":"original"`)) {
			t.Fatalf("unexpected forwarded body: %s", string(body))
		}
		w.Header().Set("X-Upstream", "default-forward")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"forwarded":true}`))
	}))
	defer upstream.Close()
	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	namespaceResp := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/namespaces/default", nil, http.StatusOK)
	namespaceBody := readBody(t, namespaceResp)
	assertBytesContain(t, namespaceBody, `"id":"default"`)
	assertBytesContain(t, namespaceBody, `"ruleset_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)
	assertBytesContain(t, namespaceBody, `"rule_miss_action":{"type":"forward","forward":{"timeout_ms":5000}}`)

	runtimeResp := doJSONWithHeaders(
		t,
		handler,
		http.MethodPost,
		"/mockserver/runtime/default/http/api/pass-through?token=abc",
		map[string]any{"request": "original"},
		map[string]string{
			"X-Forwarded-Host":  upstreamHost,
			"X-Forwarded-Proto": "http",
			"X-Client":          "mobile",
			"Connection":        "close",
		},
		http.StatusAccepted,
	)
	if got := runtimeResp.Header.Get("X-Mockserver-Fallback"); got != "ruleset_miss" {
		t.Fatalf("unexpected fallback header: %q", got)
	}
	if got := runtimeResp.Header.Get("X-Upstream"); got != "default-forward" {
		t.Fatalf("unexpected upstream header: %q", got)
	}
	assertBytesContain(t, readBody(t, runtimeResp), `"forwarded":true`)
}

func TestDefaultNamespaceRuleMissForwardsOriginalRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected forwarded method: %s", r.Method)
		}
		if r.URL.Path != "/api/miss" {
			t.Fatalf("unexpected forwarded path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("token") != "rule-miss" {
			t.Fatalf("unexpected forwarded query: %s", r.URL.RawQuery)
		}
		if got := r.Header.Get("X-Client"); got != "desktop" {
			t.Fatalf("unexpected forwarded header: %q", got)
		}
		if got := r.Header.Get("Connection"); got != "" {
			t.Fatalf("hop-by-hop header should not be forwarded: %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read forwarded body: %v", err)
		}
		if !bytes.Contains(body, []byte(`"rule":"miss"`)) {
			t.Fatalf("unexpected forwarded body: %s", string(body))
		}
		w.Header().Set("X-Upstream", "default-rule-miss-forward")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"forwarded":"rule_miss"}`))
	}))
	defer upstream.Close()
	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	ruleSetBody := map[string]any{
		"id":        "default-forward-rule-miss",
		"name":      "default forward rule miss",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "only-hit",
				"name":     "Only Hit",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"field": "request.path",
					"op":    "eq",
					"value": "/api/hit",
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"mocked": true},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", ruleSetBody, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/default-forward-rule-miss/publish", nil, http.StatusOK)

	hitResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/default/http/api/hit", nil, http.StatusOK)
	if got := hitResp.Header.Get("X-Mockserver-Rule"); got != "only-hit" {
		t.Fatalf("unexpected matched rule header: %q", got)
	}
	if got := hitResp.Header.Get("X-Mockserver-Fallback"); got != "" {
		t.Fatalf("mock hit should not be marked fallback: %q", got)
	}
	assertBytesContain(t, readBody(t, hitResp), `"mocked":true`)

	ruleMissResp := doJSONWithHeaders(
		t,
		handler,
		http.MethodPut,
		"/mockserver/runtime/default/http/api/miss?token=rule-miss",
		map[string]any{"rule": "miss"},
		map[string]string{
			"X-Forwarded-Host":  upstreamHost,
			"X-Forwarded-Proto": "http",
			"X-Client":          "desktop",
			"Connection":        "close",
		},
		http.StatusCreated,
	)
	if got := ruleMissResp.Header.Get("X-Mockserver-Fallback"); got != "rule_miss" {
		t.Fatalf("unexpected fallback header: %q", got)
	}
	if got := ruleMissResp.Header.Get("X-Upstream"); got != "default-rule-miss-forward" {
		t.Fatalf("unexpected upstream header: %q", got)
	}
	assertBytesContain(t, readBody(t, ruleMissResp), `"forwarded":"rule_miss"`)
}

func TestNamespaceForwardFallbackUsesOriginalRequestTarget(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/miss" {
			t.Fatalf("unexpected forwarded path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("token") != "abc" {
			t.Fatalf("unexpected forwarded query: %s", r.URL.RawQuery)
		}
		w.Header().Set("X-Upstream", "forwarded")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"forwarded":true}`))
	}))
	defer upstream.Close()
	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	namespaceResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/namespaces", map[string]any{
		"name": "forward namespace",
		"ruleset_miss_action": map[string]any{
			"type":    "forward",
			"forward": map[string]any{"timeout_ms": 3000},
		},
		"rule_miss_action": map[string]any{
			"type":    "forward",
			"forward": map[string]any{"timeout_ms": 3000},
		},
	}, http.StatusOK)
	var namespaceEnvelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(readBody(t, namespaceResp), &namespaceEnvelope); err != nil {
		t.Fatalf("decode namespace response: %v", err)
	}
	namespaceID := namespaceEnvelope.Data.ID

	ruleSetBody := map[string]any{
		"id":        "forward-ruleset",
		"name":      "forward ruleset",
		"enabled":   true,
		"protocol":  "http",
		"namespace": namespaceID,
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "only-hit",
				"name":     "Only Hit",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"field": "request.path",
					"op":    "eq",
					"value": "/api/hit",
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"hit": true},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", ruleSetBody, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/forward-ruleset/publish", nil, http.StatusOK)

	ruleMissResp := doJSONWithHeaders(
		t,
		handler,
		http.MethodGet,
		"/mockserver/runtime/"+namespaceID+"/http/api/miss?token=abc",
		nil,
		map[string]string{"X-Forwarded-Host": upstreamHost, "X-Forwarded-Proto": "http"},
		http.StatusCreated,
	)
	if got := ruleMissResp.Header.Get("X-Mockserver-Fallback"); got != "rule_miss" {
		t.Fatalf("unexpected fallback header: %q", got)
	}
	if got := ruleMissResp.Header.Get("X-Upstream"); got != "forwarded" {
		t.Fatalf("unexpected upstream header: %q", got)
	}
	assertBytesContain(t, readBody(t, ruleMissResp), `"forwarded":true`)
}

func TestAdminAuthMiddleware(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{
		AdminToken:   "admin-token",
		ReadToken:    "read-token",
		WriteToken:   "write-token",
		PublishToken: "publish-token",
	}, nil)

	doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil, http.StatusUnauthorized)
	doJSONWithHeaders(t, handler, http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil, map[string]string{
		"Authorization": "Bearer read-token",
	}, http.StatusOK)

	authRuleSet := map[string]any{
		"id":        "auth-case",
		"name":      "auth case",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "auth-rule",
				"name":     "Auth Rule",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.method", "op": "eq", "value": "GET"},
						{"field": "request.path", "op": "eq", "value": "/api/v1/auth"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"auth": true},
				},
			},
		},
	}

	doJSONWithHeaders(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", authRuleSet, map[string]string{
		"Authorization": "Bearer read-token",
	}, http.StatusForbidden)
	doJSONWithHeaders(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", authRuleSet, map[string]string{
		"Authorization": "Bearer write-token",
	}, http.StatusOK)
	doJSONWithHeaders(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/auth-case/publish", nil, map[string]string{
		"Authorization": "Bearer write-token",
	}, http.StatusForbidden)
	publishResp := doJSONWithHeaders(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/auth-case/publish", map[string]any{
		"reason": "auth publish",
	}, map[string]string{
		"X-Mockserver-Admin-Token": "publish-token",
		"X-Mockserver-Operator":    "admin@example.com",
		"X-Trace-ID":               "trace-auth-publish",
	}, http.StatusOK)
	publishBody := readBody(t, publishResp)
	assertBytesContain(t, publishBody, `"audit"`)
	assertBytesContain(t, publishBody, `"action":"publish"`)
	assertBytesContain(t, publishBody, `"operator":"admin@example.com"`)
	assertBytesContain(t, publishBody, `"reason":"auth publish"`)
	assertBytesContain(t, publishBody, `"trace_id":"trace-auth-publish"`)

	runtimeResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/default/http/api/v1/auth", nil, http.StatusOK)
	assertBytesContain(t, readBody(t, runtimeResp), `"auth":true`)
}

func TestAdminDraftRuleManagementFlow(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	baseRuleSet := map[string]any{
		"id":        "rule-management",
		"name":      "rule management",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "base-rule",
				"name":     "Base Rule",
				"enabled":  true,
				"priority": 10,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.method", "op": "eq", "value": "GET"},
						{"field": "request.path", "op": "eq", "value": "/api/v1/base"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"rule": "base"},
				},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", baseRuleSet, http.StatusOK)

	newRule := map[string]any{
		"id":       "dynamic-rule",
		"name":     "Dynamic Rule",
		"enabled":  true,
		"priority": 50,
		"when": map[string]any{
			"all": []map[string]any{
				{"field": "request.method", "op": "eq", "value": "GET"},
				{"field": "request.path", "op": "eq", "value": "/api/v1/dynamic"},
			},
		},
		"action": map[string]any{
			"type":   "static_response",
			"status": 200,
			"body":   map[string]any{"rule": "dynamic"},
		},
	}
	addResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rule-management/rules", map[string]any{
		"rule": newRule,
	}, http.StatusOK)
	assertBytesContain(t, readBody(t, addResp), `"id":"dynamic-rule"`)

	priorityResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rule-management/rules/dynamic-rule/priority", map[string]any{
		"priority": 250,
	}, http.StatusOK)
	assertBytesContain(t, readBody(t, priorityResp), `"priority":250`)

	disableResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rule-management/rules/dynamic-rule/disable", nil, http.StatusOK)
	assertBytesContain(t, readBody(t, disableResp), `"enabled":false`)

	enableResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rule-management/rules/dynamic-rule/enable", nil, http.StatusOK)
	assertBytesContain(t, readBody(t, enableResp), `"enabled":true`)

	updatedRule := cloneMap(newRule)
	updatedRule["priority"] = 300
	updatedRule["action"] = map[string]any{
		"type":   "static_response",
		"status": 201,
		"body":   map[string]any{"rule": "updated"},
	}
	updateResp := doJSON(t, handler, http.MethodPut, "/mockserver/api/v1/admin/rulesets/rule-management/rules/dynamic-rule", map[string]any{
		"rule": updatedRule,
	}, http.StatusOK)
	assertBytesContain(t, readBody(t, updateResp), `"status":201`)

	deleteResp := doJSON(t, handler, http.MethodDelete, "/mockserver/api/v1/admin/rulesets/rule-management/rules/base-rule", nil, http.StatusOK)
	deleteBody := readBody(t, deleteResp)
	assertBytesContain(t, deleteBody, `"id":"dynamic-rule"`)
	assertBytesNotContain(t, deleteBody, `"id":"base-rule"`)

	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rule-management/validate", nil, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rule-management/publish", nil, http.StatusOK)

	runtimeResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/default/http/api/v1/dynamic", nil, http.StatusCreated)
	if got := runtimeResp.Header.Get("X-Mockserver-Rule"); got != "dynamic-rule" {
		t.Fatalf("unexpected matched rule header: %s", got)
	}
	assertBytesContain(t, readBody(t, runtimeResp), `"rule":"updated"`)
}

func TestAdminRollbackPublishedSnapshot(t *testing.T) {
	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := service.NewNamespaceService(ruleSetRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)

	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, observability.NewRuntimeMetrics())
	handler := router.New(adminController, runtimeController, router.AdminAuthConfig{}, nil)

	v1Body := map[string]any{
		"id":        "rollback-case",
		"name":      "rollback case",
		"enabled":   true,
		"protocol":  "http",
		"namespace": "default",
		"selector":  httpPathSelectorBody("/api/"),
		"rules": []map[string]any{
			{
				"id":       "state-rule",
				"name":     "State Rule",
				"enabled":  true,
				"priority": 100,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.method", "op": "eq", "value": "GET"},
						{"field": "request.path", "op": "eq", "value": "/api/v1/state"},
					},
				},
				"action": map[string]any{
					"type":   "static_response",
					"status": 200,
					"body":   map[string]any{"version": 1},
				},
			},
		},
	}

	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", v1Body, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rollback-case/publish", nil, http.StatusOK)

	v2Body := cloneMap(v1Body)
	v2Body["selector"] = httpPathSelectorBody("/api/v2/")
	v2Body["rules"] = []map[string]any{
		{
			"id":       "state-rule",
			"name":     "State Rule",
			"enabled":  true,
			"priority": 100,
			"when": map[string]any{
				"all": []map[string]any{
					{"field": "request.method", "op": "eq", "value": "GET"},
					{"field": "request.path", "op": "eq", "value": "/api/v1/state"},
				},
			},
			"action": map[string]any{
				"type":   "static_response",
				"status": 200,
				"body":   map[string]any{"version": 2},
			},
		},
	}
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", v2Body, http.StatusOK)
	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/rollback-case/publish", nil, http.StatusOK)

	snapshotsResp := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/published/rulesets/rollback-case/snapshots", nil, http.StatusOK)
	data := decodeEnvelopeData(t, snapshotsResp)
	var snapshots struct {
		Items []struct {
			SnapshotID string `json:"snapshot_id"`
			RuleSet    struct {
				Version int `json:"version"`
			} `json:"ruleset"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &snapshots); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(snapshots.Items) != 2 {
		t.Fatalf("unexpected snapshot count: %d", len(snapshots.Items))
	}
	var rollbackSnapshotID string
	for _, item := range snapshots.Items {
		if item.RuleSet.Version == 1 {
			rollbackSnapshotID = item.SnapshotID
			break
		}
	}
	if rollbackSnapshotID == "" {
		t.Fatalf("failed to find version 1 snapshot")
	}

	previewResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/published/rulesets/rollback-case/rollback/preview", map[string]any{
		"snapshot_id":     rollbackSnapshotID,
		"explain_compact": true,
		"explain_summary": true,
		"event": map[string]any{
			"protocol":  "http",
			"namespace": "default",
			"request": map[string]any{
				"method": "GET",
				"path":   "/api/v1/state",
			},
		},
	}, http.StatusOK)
	previewBody := readBody(t, previewResp)
	assertBytesContain(t, previewBody, `"valid":true`)
	assertBytesContain(t, previewBody, `"simulation"`)
	assertBytesContain(t, previewBody, `"version":1`)
	assertBytesContain(t, previewBody, `"diff"`)
	assertBytesContain(t, previewBody, `"rule_id":"state-rule"`)
	assertBytesContain(t, previewBody, `"action_changed":true`)
	assertBytesContain(t, previewBody, `"ruleset_field_diffs"`)
	assertBytesContain(t, previewBody, `"field_diffs"`)
	assertBytesContain(t, previewBody, `"path":"action.body.version"`)
	assertBytesContain(t, previewBody, `"path":"ruleset.selector.all[0].value"`)
	assertBytesNotContain(t, previewBody, `"selector_checks"`)
	assertBytesNotContain(t, previewBody, `"candidate_rules"`)

	doJSONWithHeaders(t, handler, http.MethodPost, "/mockserver/api/v1/admin/published/rulesets/rollback-case/rollback", map[string]any{
		"snapshot_id": rollbackSnapshotID,
		"reason":      "restore v1",
	}, map[string]string{
		"X-Mockserver-Operator": "rollbacker@example.com",
		"X-Trace-ID":            "trace-rollback",
	}, http.StatusOK)

	runtimeResp := doJSON(t, handler, http.MethodGet, "/mockserver/runtime/default/http/api/v1/state", nil, http.StatusOK)
	body := readBody(t, runtimeResp)
	if !bytes.Contains(body, []byte(`"version":1`)) {
		t.Fatalf("unexpected runtime body after rollback: %s", string(body))
	}

	latestPublished := doJSON(t, handler, http.MethodGet, "/mockserver/api/v1/admin/published/rulesets/rollback-case", nil, http.StatusOK)
	latestBody := readBody(t, latestPublished)
	assertBytesContain(t, latestBody, `"version":1`)
	assertBytesContain(t, latestBody, `"action":"rollback"`)
	assertBytesContain(t, latestBody, `"operator":"rollbacker@example.com"`)
	assertBytesContain(t, latestBody, `"reason":"restore v1"`)
	assertBytesContain(t, latestBody, `"trace_id":"trace-rollback"`)
	assertBytesContain(t, latestBody, `"source_snapshot_id":"`+rollbackSnapshotID+`"`)
}

func selectorBody(conditions ...map[string]any) map[string]any {
	return map[string]any{"all": conditions}
}

func selectorCondition(field string, op string, value any) map[string]any {
	return map[string]any{"field": field, "op": op, "value": value}
}

func httpPathSelectorBody(prefix string) map[string]any {
	return selectorBody(selectorCondition("request.path", "prefix", prefix))
}

func httpHostSelectorBody(host string) map[string]any {
	return selectorBody(selectorCondition("request.host", "eq", host))
}

func cacheKeySelectorBody(operation string, keyPrefix string) map[string]any {
	return selectorBody(
		selectorCondition("request.operation", "eq", operation),
		selectorCondition("request.key", "prefix", keyPrefix),
	)
}

func doJSON(t *testing.T, handler http.Handler, method, path string, body any, wantStatus int) *http.Response {
	return doJSONWithHeaders(t, handler, method, path, body, nil, wantStatus)
}

func doJSONWithHeaders(t *testing.T, handler http.Handler, method, path string, body any, headers map[string]string, wantStatus int) *http.Response {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	resp := recorder.Result()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got=%d want=%d", resp.StatusCode, wantStatus)
	}
	return resp
}

func assertResponseContains(t *testing.T, resp *http.Response, want string) {
	t.Helper()

	body := readBody(t, resp)
	assertBytesContain(t, body, want)
}

func assertBytesContain(t *testing.T, body []byte, want string) {
	t.Helper()

	if !bytes.Contains(body, []byte(want)) {
		t.Fatalf("response body %q does not contain %q", string(body), want)
	}
}

func assertBytesNotContain(t *testing.T, body []byte, want string) {
	t.Helper()

	if bytes.Contains(body, []byte(want)) {
		t.Fatalf("response body %q unexpectedly contains %q", string(body), want)
	}
}

func decodeEnvelopeData(t *testing.T, resp *http.Response) []byte {
	t.Helper()

	body := readBody(t, resp)
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return envelope.Data
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	return body
}

func cloneMap(src map[string]any) map[string]any {
	out := make(map[string]any, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}
