package observability

import (
	"fmt"
	"testing"
	"time"

	"mockserver/internal/model/bo"
)

func TestRuntimeMetricsRecordsRecentDiagnostics(t *testing.T) {
	metrics := NewRuntimeMetrics()
	event := bo.Event{
		Protocol:  "http",
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"scheme": "http",
			"host":   "demo.test",
			"path":   "/api/hit",
			"query": map[string][]string{
				"q": []string{"1"},
			},
		},
		Meta: bo.EventMeta{TraceID: "trace-hit"},
	}

	metrics.Observe(RuntimeObservation{
		Namespace: "default",
		Method:    "GET",
		Path:      "/api/hit",
		TraceID:   "trace-hit",
		Matched:   true,
		Status:    200,
		RulesetID: "rs-a",
		RuleID:    "rule-a",
		Duration:  12 * time.Millisecond,
		Event:     &event,
	})
	metrics.Observe(RuntimeObservation{
		Namespace:      "default",
		Method:         "POST",
		Path:           "/api/rule-miss",
		Fallback:       true,
		Status:         504,
		FallbackReason: "rule_miss",
		Duration:       8 * time.Millisecond,
	})
	metrics.Observe(RuntimeObservation{
		Namespace: "default",
		Method:    "GET",
		Path:      "/api/error",
		Error:     true,
		Status:    500,
		Message:   "boom",
		Duration:  5 * time.Millisecond,
	})

	snapshot := metrics.Snapshot()
	if snapshot.TotalRequests != 3 || snapshot.MatchedRequests != 1 || snapshot.UnmatchedRequests != 1 || snapshot.ErrorRequests != 1 {
		t.Fatalf("unexpected aggregate counts: %+v", snapshot)
	}
	if snapshot.RulesetMatches["rs-a"] != 1 || snapshot.RuleMatches["rule-a"] != 1 {
		t.Fatalf("expected matched ruleset/rule counts, got %+v %+v", snapshot.RulesetMatches, snapshot.RuleMatches)
	}
	if snapshot.FallbackReasons["rule_miss"] != 1 {
		t.Fatalf("expected fallback reason count, got %+v", snapshot.FallbackReasons)
	}
	if snapshot.StatusCodes["2xx"] != 1 || snapshot.StatusCodes["5xx"] != 2 {
		t.Fatalf("expected status class counts, got %+v", snapshot.StatusCodes)
	}
	if len(snapshot.RecentRequests) != 3 {
		t.Fatalf("expected 3 recent requests, got %d", len(snapshot.RecentRequests))
	}
	if snapshot.RecentRequests[0].Outcome != "error" || snapshot.RecentRequests[1].Outcome != "fallback" || snapshot.RecentRequests[2].Outcome != "matched" {
		t.Fatalf("recent requests should be newest first with outcomes, got %+v", snapshot.RecentRequests)
	}
	if snapshot.RecentRequests[2].Event == nil || snapshot.RecentRequests[2].Event.Meta.TraceID != "trace-hit" {
		t.Fatalf("expected replayable event to be cloned, got %+v", snapshot.RecentRequests[2].Event)
	}
}

func TestRuntimeMetricsRecentRequestBufferIsBounded(t *testing.T) {
	metrics := NewRuntimeMetrics()
	for i := 0; i < maxRecentRuntimeRequests+5; i++ {
		metrics.Observe(RuntimeObservation{
			Method:   "GET",
			Path:     fmt.Sprintf("/api/%d", i),
			Status:   200,
			Matched:  true,
			Duration: time.Millisecond,
		})
	}

	snapshot := metrics.Snapshot()
	if len(snapshot.RecentRequests) != maxRecentRuntimeRequests {
		t.Fatalf("expected bounded recent requests, got %d", len(snapshot.RecentRequests))
	}
	if snapshot.RecentRequests[0].ID != int64(maxRecentRuntimeRequests+5) {
		t.Fatalf("expected newest request first, got id %d", snapshot.RecentRequests[0].ID)
	}
	if snapshot.RecentRequests[len(snapshot.RecentRequests)-1].ID != 6 {
		t.Fatalf("expected oldest retained request id 6, got %d", snapshot.RecentRequests[len(snapshot.RecentRequests)-1].ID)
	}
}
