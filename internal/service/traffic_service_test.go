package service

import (
	"context"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type testTrafficRepository struct {
	event bo.TrafficEvent
}

func (r *testTrafficRepository) CreateTrafficEvent(ctx context.Context, event bo.TrafficEvent) (bo.TrafficEvent, error) {
	_ = ctx
	event.ID = 101
	r.event = event
	return event, nil
}

func (r *testTrafficRepository) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	_, _ = ctx, query
	return bo.TrafficEventList{}, nil
}

func TestTrafficServiceRecordSDKDecisionBuildsProtocolNeutralEventAndIndexes(t *testing.T) {
	repository := &testTrafficRepository{}
	service := NewTrafficService(repository)

	event := bo.Event{
		Protocol:  "http",
		Operation: "request",
		Namespace: "shop",
		Request: bo.EventRequest{
			"method": "GET",
			"host":   "api.example.test",
			"path":   "/orders",
			"query": map[string][]string{
				"shop_id": {"123"},
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
		Response: &bo.ActionExecution{Status: 202},
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

	indexes := map[string]bo.TrafficEventIndex{}
	for _, index := range repository.event.Indexes {
		indexes[index.FieldPath] = index
		if index.FieldPath == "status_code" {
			t.Fatalf("HTTP status must not be stored as a top-level status_code index")
		}
	}
	for _, path := range []string{
		"event.operation",
		"event.request.method",
		"event.request.host",
		"event.request.path",
		"event.request.query.shop_id",
		"decision.response.status",
	} {
		index, ok := indexes[path]
		if !ok {
			t.Fatalf("missing index %s from %#v", path, indexes)
		}
		if index.FieldValueText == "" || index.FieldValueHash == 0 {
			t.Fatalf("index %s should include full value and uint64 hash: %+v", path, index)
		}
	}
	if indexes["decision.response.status"].FieldValueText != "202" {
		t.Fatalf("unexpected response status index: %+v", indexes["decision.response.status"])
	}
}
