package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type testRuntimeViewForTraffic struct {
	decision bo.RuntimeDecision
	err      error
}

func (v *testRuntimeViewForTraffic) MatchPublished(ctx context.Context, event bo.Event) (bo.SimulationResult, error) {
	_, _ = ctx, event
	return bo.SimulationResult{}, nil
}

func (v *testRuntimeViewForTraffic) DecidePublished(ctx context.Context, event bo.Event) (bo.RuntimeDecision, error) {
	_, _ = ctx, event
	return v.decision, v.err
}

type testTrafficServiceForController struct {
	count       int
	event       bo.Event
	decision    bo.RuntimeDecision
	decisionErr error
}

func (s *testTrafficServiceForController) RecordSDKDecision(ctx context.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, durationMS uint32) (bo.TrafficEvent, error) {
	_, _ = ctx, durationMS
	s.count++
	s.event = event
	s.decision = decision
	s.decisionErr = decisionErr
	return bo.TrafficEvent{ID: 1}, nil
}

func (s *testTrafficServiceForController) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	_, _ = ctx, query
	return bo.TrafficEventList{}, nil
}

func TestRuntimeControllerDecidePublishedRecordsSDKTraffic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	trafficService := &testTrafficServiceForController{}
	controller := NewRuntimeControllerWithTraffic(&testRuntimeViewForTraffic{
		decision: bo.RuntimeDecision{
			Kind:    "response",
			Matched: true,
			Trace: bo.MatchTrace{
				RulesetID: "rs",
				RuleID:    "rule",
			},
			Response: &bo.ActionExecution{Status: http.StatusAccepted},
		},
	}, nil, trafficService)
	engine := gin.New()
	engine.POST("/mockserver/api/v1/sdk/decision", controller.DecidePublished)

	body := []byte(`{"event":{"protocol":"http","operation":"request","namespace":"shop","request":{"method":"GET","path":"/orders"},"meta":{"trace_id":"trace-001"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/mockserver/api/v1/sdk/decision", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if trafficService.count != 1 {
		t.Fatalf("expected one traffic record, got %d", trafficService.count)
	}
	if trafficService.event.Protocol != "http" || trafficService.event.Namespace != "shop" {
		t.Fatalf("unexpected recorded event: %+v", trafficService.event)
	}
	if trafficService.decision.Kind != "response" || !trafficService.decision.Matched {
		t.Fatalf("unexpected recorded decision: %+v", trafficService.decision)
	}
	if trafficService.decisionErr != nil {
		t.Fatalf("unexpected decision error: %v", trafficService.decisionErr)
	}
}
