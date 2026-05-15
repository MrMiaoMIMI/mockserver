package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

type testTrafficRepository struct {
	event       bo.TrafficEvent
	createCalls int
}

func (r *testTrafficRepository) CreateTrafficEvent(ctx context.Context, event bo.TrafficEvent) (bo.TrafficEvent, error) {
	_ = ctx
	r.createCalls++
	event.ID = 101
	r.event = event
	return event, nil
}

func (r *testTrafficRepository) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	_, _ = ctx, query
	return bo.TrafficEventList{}, nil
}

func (r *testTrafficRepository) GetTrafficEvent(ctx context.Context, id uint64) (bo.TrafficEvent, bool, error) {
	_ = ctx
	if r.event.ID != id {
		return bo.TrafficEvent{}, false, nil
	}
	return r.event, true, nil
}

func TestTrafficServiceRecordSDKDecisionBuildsMinimalQueryableIndexes(t *testing.T) {
	repository := &testTrafficRepository{}
	service := NewTrafficService(repository)

	event := bo.Event{
		Protocol:  "http",
		Operation: "request",
		Namespace: "shop",
		Request: bo.EventRequest{
			"method":        "GET",
			"host":          "api.example.test",
			"original_host": "public.example.test",
			"path":          "/orders",
			"query": map[string][]string{
				"shop_id": {"123"},
			},
			"headers": map[string][]string{
				"authorization": {"Bearer secret"},
			},
		},
		Meta: bo.EventMeta{TraceID: "trace-001"},
	}
	decision := bo.RuntimeDecision{
		Kind:    "response",
		Matched: true,
		Trace: bo.MatchTrace{
			RulesetID: "rs-http",
			RuleID:    "rule-hit",
		},
		Response: &bo.ProtocolResponse{
			Protocol: eo.ProtocolHTTP,
			Payload:  map[string]any{"status": 202},
		},
		Forward: &bo.ForwardDecision{TimeoutMS: 800},
		Diagnostics: &bo.DecisionDiagnostics{
			RuleSetSelection: bo.RuleSetSelectionDiagnostics{
				WinnerRuleSetID: "rs-http",
				Candidates: []bo.RuleSetCandidate{{
					RuleSetID:           "rs-http",
					SelectorMatched:     true,
					Selected:            true,
					SelectorSpecificity: 120,
				}},
			},
			RuleSelection: bo.RuleSelectionDiagnostics{
				CandidateRuleIDs: []string{"rule-hit"},
				WinnerRuleID:     "rule-hit",
			},
		},
	}

	created, err := service.RecordSDKDecision(context.Background(), event, decision, nil, 12)
	if err != nil {
		t.Fatalf("RecordSDKDecision() error = %v", err)
	}
	if created.ID != 101 {
		t.Fatalf("expected repository id to be returned, got %d", created.ID)
	}
	if repository.event.TrafficSource != bo.TrafficSourceSDKDecision {
		t.Fatalf("unexpected traffic source: %s", repository.event.TrafficSource)
	}
	if repository.event.ProtocolName != "http" || repository.event.NamespaceID != "shop" {
		t.Fatalf("unexpected protocol namespace: %+v", repository.event)
	}
	if repository.event.Outcome != bo.TrafficOutcomeMatched || repository.event.DecisionKind != "response" {
		t.Fatalf("unexpected outcome/decision kind: %+v", repository.event)
	}
	if repository.event.EventTime == 0 || repository.event.ExpireTime <= repository.event.EventTime {
		t.Fatalf("event time and expire time should be set: %+v", repository.event)
	}
	var explain map[string]any
	if err := json.Unmarshal([]byte(repository.event.ExplainJSON), &explain); err != nil {
		t.Fatalf("unmarshal explain json: %v", err)
	}
	if _, ok := explain["ruleset_selection"]; !ok {
		t.Fatalf("expected ruleset selection diagnostics in explain json: %s", repository.event.ExplainJSON)
	}
	if _, ok := explain["rule_selection"]; !ok {
		t.Fatalf("expected rule selection diagnostics in explain json: %s", repository.event.ExplainJSON)
	}

	indexes := map[string]bo.TrafficEventIndex{}
	for _, index := range repository.event.Indexes {
		indexes[index.FieldPath] = index
		switch index.FieldPath {
		case "status_code",
			"event.operation",
			"event.request.original_host",
			"event.request.query.shop_id",
			"event.request.headers.authorization",
			"decision.forward.timeout_ms":
			t.Fatalf("unexpected high-cardinality or duplicate index %s", index.FieldPath)
		}
	}
	for _, path := range []string{
		"event.request.method",
		"event.request.host",
		"event.request.path",
		"decision.response.protocol",
		"decision.response.payload.status",
	} {
		index, ok := indexes[path]
		if !ok {
			t.Fatalf("missing index %s from %#v", path, indexes)
		}
		if index.FieldValueText == "" || index.FieldValueHash == 0 {
			t.Fatalf("index %s should include full value and uint64 hash: %+v", path, index)
		}
	}
	if indexes["decision.response.payload.status"].FieldValueText != "202" {
		t.Fatalf("unexpected response status index: %+v", indexes["decision.response.payload.status"])
	}
}

func TestTrafficServiceRecordSDKDecisionSkipsRulesetMissRawTraffic(t *testing.T) {
	repository := &testTrafficRepository{}
	service := NewTrafficService(repository)

	event := bo.Event{
		Protocol:  "http",
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "host": "unexpected.test", "path": "/noise"},
	}
	decision := bo.RuntimeDecision{
		Kind:     eo.DecisionKindForward,
		Fallback: true,
		Trace:    bo.MatchTrace{FallbackReason: eo.FallbackReasonRulesetMiss},
		Forward:  &bo.ForwardDecision{TimeoutMS: 5000},
	}

	created, err := service.RecordSDKDecision(context.Background(), event, decision, nil, 1)
	if err != nil {
		t.Fatalf("RecordSDKDecision() error = %v", err)
	}
	if created.ID != 0 {
		t.Fatalf("expected zero event when raw traffic is skipped, got %+v", created)
	}
	if repository.createCalls != 0 {
		t.Fatalf("ruleset_miss should not persist raw traffic, create calls=%d", repository.createCalls)
	}
}

func TestTrafficServiceRecordSDKDecisionSkipsRulesetMissResponseFallback(t *testing.T) {
	repository := &testTrafficRepository{}
	service := NewTrafficService(repository)

	decision := bo.RuntimeDecision{
		Kind:     eo.DecisionKindResponse,
		Fallback: true,
		Trace:    bo.MatchTrace{FallbackReason: eo.FallbackReasonRulesetMiss},
		Response: &bo.ProtocolResponse{
			Protocol: eo.ProtocolSPEX,
			Payload:  map[string]any{"code": 404, "resp": "{}"},
		},
	}

	created, err := service.RecordSDKDecision(context.Background(), bo.Event{
		Protocol:  "spex",
		Namespace: "default",
		Request:   bo.EventRequest{"cmd": "shop.GetOrder"},
	}, decision, nil, 1)
	if err != nil {
		t.Fatalf("RecordSDKDecision() error = %v", err)
	}
	if created.ID != 0 || repository.createCalls != 0 {
		t.Fatalf("ruleset_miss response fallback should skip raw traffic, created=%+v create calls=%d", created, repository.createCalls)
	}
}

func TestTrafficServiceRecordSDKDecisionDoesNotSkipMalformedRulesetMissWithoutFallback(t *testing.T) {
	repository := &testTrafficRepository{}
	service := NewTrafficService(repository)

	decision := bo.RuntimeDecision{
		Kind:  eo.DecisionKindResponse,
		Trace: bo.MatchTrace{FallbackReason: eo.FallbackReasonRulesetMiss},
	}

	created, err := service.RecordSDKDecision(context.Background(), bo.Event{
		Protocol:  "http",
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "path": "/noise"},
	}, decision, nil, 1)
	if err != nil {
		t.Fatalf("RecordSDKDecision() error = %v", err)
	}
	if created.ID == 0 || repository.createCalls != 1 {
		t.Fatalf("malformed non-fallback ruleset_miss should remain persisted, created=%+v create calls=%d", created, repository.createCalls)
	}
}

func TestTrafficServiceRecordSDKDecisionPersistsRuleMissRawTraffic(t *testing.T) {
	repository := &testTrafficRepository{}
	service := NewTrafficService(repository)

	decision := bo.RuntimeDecision{
		Kind:     eo.DecisionKindForward,
		Fallback: true,
		Trace: bo.MatchTrace{
			RulesetID:      "http-api",
			FallbackReason: eo.FallbackReasonRuleMiss,
		},
		Forward: &bo.ForwardDecision{TimeoutMS: 5000},
	}

	created, err := service.RecordSDKDecision(context.Background(), bo.Event{
		Protocol:  "http",
		Namespace: "default",
		Request:   bo.EventRequest{"method": "GET", "host": "api.test", "path": "/api/no-rule"},
	}, decision, nil, 1)
	if err != nil {
		t.Fatalf("RecordSDKDecision() error = %v", err)
	}
	if created.ID == 0 || repository.createCalls != 1 {
		t.Fatalf("rule_miss should persist raw traffic, created=%+v create calls=%d", created, repository.createCalls)
	}
	if repository.event.RuleSetID != "http-api" || repository.event.FallbackReason != eo.FallbackReasonRuleMiss {
		t.Fatalf("expected rule_miss winner ruleset in persisted traffic, got %+v", repository.event)
	}
}

func TestTrafficServiceGetTrafficEventReturnsDetailOrNotFound(t *testing.T) {
	repository := &testTrafficRepository{event: bo.TrafficEvent{ID: 12, EventID: "te_12"}}
	service := NewTrafficService(repository)

	got, err := service.GetTrafficEvent(context.Background(), 12)
	if err != nil {
		t.Fatalf("GetTrafficEvent() error = %v", err)
	}
	if got.EventID != "te_12" {
		t.Fatalf("unexpected event: %+v", got)
	}

	if _, err := service.GetTrafficEvent(context.Background(), 13); err == nil {
		t.Fatalf("expected not found error")
	}
}
