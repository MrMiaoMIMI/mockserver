package response

import (
	"encoding/json"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

func TestTrafficListResponseOmitsPayloadAndIndexes(t *testing.T) {
	response := NewListTrafficEventsResponse(bo.TrafficEventList{
		Items: []bo.TrafficEvent{trafficEventForResponseTest()},
		Total: 1,
		Stats: bo.TrafficStats{Total: 1},
	})
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	items := decoded["items"].([]any)
	item := items[0].(map[string]any)
	for _, key := range []string{"event", "decision", "explain", "indexes"} {
		if _, ok := item[key]; ok {
			t.Fatalf("summary response must omit %s: %s", key, raw)
		}
	}
}

func TestTrafficDetailResponseIncludesPayloadAndIndexes(t *testing.T) {
	item := NewTrafficEventDetailResponse(trafficEventForResponseTest())
	if item.Event == nil || item.Decision == nil || item.Explain == nil {
		t.Fatalf("detail response should include payloads: %+v", item)
	}
	if len(item.Indexes) != 1 || item.Indexes[0].FieldPath != "event.request.path" {
		t.Fatalf("detail response should include indexes: %+v", item.Indexes)
	}
}

func trafficEventForResponseTest() bo.TrafficEvent {
	return bo.TrafficEvent{
		ID:            9,
		EventID:       "te_9",
		TrafficSource: bo.TrafficSourceSDKDecision,
		ProtocolName:  "http",
		NamespaceID:   "default",
		Outcome:       bo.TrafficOutcomeMatched,
		DurationMS:    3,
		EventTime:     100,
		ExpireTime:    200,
		EventJSON:     `{"protocol":"http"}`,
		DecisionJSON:  `{"kind":"response"}`,
		ExplainJSON:   `{"trace":{}}`,
		Indexes: []bo.TrafficEventIndex{{
			ID:                1,
			TrafficEventID:    9,
			ProtocolName:      "http",
			FieldPath:         "event.request.path",
			FieldValuePreview: "/orders",
			FieldValueHash:    123,
			FieldValueText:    "/orders",
			EventTime:         100,
		}},
	}
}
