package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type testTrafficServiceForTrafficController struct {
	listQuery bo.TrafficQuery
	getID     uint64
}

func (s *testTrafficServiceForTrafficController) RecordSDKDecision(ctx context.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, durationMS uint32) (bo.TrafficEvent, error) {
	_, _, _, _, _ = ctx, event, decision, decisionErr, durationMS
	return bo.TrafficEvent{}, nil
}

func (s *testTrafficServiceForTrafficController) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	_ = ctx
	s.listQuery = query
	return bo.TrafficEventList{
		Items: []bo.TrafficEvent{trafficControllerEvent()},
		Total: 1,
		Stats: bo.TrafficStats{Total: 1},
	}, nil
}

func (s *testTrafficServiceForTrafficController) GetTrafficEvent(ctx context.Context, id uint64) (bo.TrafficEvent, error) {
	_ = ctx
	s.getID = id
	return trafficControllerEvent(), nil
}

func TestTrafficControllerListEventsBindsEventIDAndReturnsSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &testTrafficServiceForTrafficController{}
	engine := gin.New()
	engine.GET("/traffic/events", NewTrafficController(service).ListEvents)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/traffic/events?event_id=te_001&limit=10&offset=20", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.listQuery.EventID != "te_001" || service.listQuery.Limit != 10 || service.listQuery.Offset != 20 {
		t.Fatalf("unexpected query: %+v", service.listQuery)
	}
	body := recorder.Body.String()
	for _, unexpected := range []string{`"event"`, `"decision"`, `"explain"`, `"indexes"`} {
		if strings.Contains(body, unexpected) {
			t.Fatalf("list response should omit %s: %s", unexpected, body)
		}
	}
}

func TestTrafficControllerGetEventReturnsDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &testTrafficServiceForTrafficController{}
	engine := gin.New()
	engine.GET("/traffic/events/:traffic_event_id", NewTrafficController(service).GetEvent)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/traffic/events/42", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.getID != 42 {
		t.Fatalf("unexpected detail id: %d", service.getID)
	}
	body := recorder.Body.String()
	for _, expected := range []string{`"event"`, `"decision"`, `"explain"`, `"indexes"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("detail response should include %s: %s", expected, body)
		}
	}
}

func trafficControllerEvent() bo.TrafficEvent {
	return bo.TrafficEvent{
		ID:            42,
		EventID:       "te_001",
		TrafficSource: bo.TrafficSourceSDKDecision,
		ProtocolName:  "http",
		NamespaceID:   "default",
		Outcome:       bo.TrafficOutcomeMatched,
		DecisionKind:  "response",
		DurationMS:    7,
		EventTime:     1000,
		ExpireTime:    2000,
		EventJSON:     `{"protocol":"http"}`,
		DecisionJSON:  `{"kind":"response"}`,
		ExplainJSON:   `{"trace":{}}`,
		Indexes: []bo.TrafficEventIndex{{
			ID:                1,
			TrafficEventID:    42,
			ProtocolName:      "http",
			FieldPath:         "event.request.path",
			FieldValuePreview: "/orders",
			FieldValueHash:    99,
			FieldValueText:    "/orders",
			EventTime:         1000,
		}},
	}
}
