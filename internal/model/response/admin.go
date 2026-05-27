package response

import (
	"encoding/json"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/mockprotocol"
)

type RuleSetResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Enabled   bool        `json:"enabled"`
	Protocol  string      `json:"protocol"`
	Namespace string      `json:"namespace"`
	Selector  bo.Selector `json:"selector"`
	Rules     []bo.Rule   `json:"rules"`
	Version   int         `json:"version"`
}

type ListRuleSetsResponse struct {
	Items []RuleSetResponse `json:"items"`
	Total int               `json:"total"`
}

type NamespaceResponse struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name,omitempty"`
	Description string                        `json:"description,omitempty"`
	Policies    map[string]bo.NamespacePolicy `json:"policies"`
	Version     int                           `json:"version"`
}

type ListNamespacesResponse struct {
	Items []NamespaceResponse `json:"items"`
	Total int                 `json:"total"`
}

type PublishedRuleSetResponse struct {
	SnapshotID  string          `json:"snapshot_id"`
	PublishedAt string          `json:"published_at"`
	RuleSet     RuleSetResponse `json:"ruleset"`
	Audit       *bo.AuditInfo   `json:"audit,omitempty"`
}

type ListPublishedRuleSetsResponse struct {
	Items []PublishedRuleSetResponse `json:"items"`
	Total int                        `json:"total"`
}

type ValidateRuleSetResponse struct {
	Result bo.ValidationResult `json:"result"`
}

type PublishRuleSetResponse struct {
	RuleSet  RuleSetResponse          `json:"ruleset"`
	Snapshot PublishedRuleSetResponse `json:"snapshot"`
}

type RollbackRuleSetResponse struct {
	RuleSet  RuleSetResponse          `json:"ruleset"`
	Snapshot PublishedRuleSetResponse `json:"snapshot"`
}

type RollbackPreviewRuleSetResponse struct {
	Result bo.RollbackPreviewResult `json:"result"`
}

type SimulateRuleSetResponse struct {
	Result bo.SimulationResult `json:"result"`
}

type DecidePublishedResponse struct {
	Decision bo.RuntimeDecision `json:"decision"`
}

type TrafficEventIndexResponse struct {
	ID                uint64 `json:"id"`
	TrafficEventID    uint64 `json:"traffic_event_id"`
	ProtocolName      string `json:"protocol_name"`
	FieldPath         string `json:"field_path"`
	FieldValuePreview string `json:"field_value_preview"`
	FieldValueHash    uint64 `json:"field_value_hash"`
	FieldValueText    string `json:"field_value_text,omitempty"`
	EventTime         uint64 `json:"event_time"`
}

type TrafficEventResponse struct {
	ID             uint64                      `json:"id"`
	EventID        string                      `json:"event_id"`
	TraceID        string                      `json:"trace_id,omitempty"`
	ScenarioID     string                      `json:"scenario_id,omitempty"`
	TrafficSource  string                      `json:"traffic_source"`
	ProtocolName   string                      `json:"protocol_name"`
	NamespaceID    string                      `json:"namespace_id"`
	OperationName  string                      `json:"operation_name,omitempty"`
	Outcome        string                      `json:"outcome"`
	DecisionKind   string                      `json:"decision_kind,omitempty"`
	RuleSetID      string                      `json:"ruleset_id,omitempty"`
	RuleID         string                      `json:"rule_id,omitempty"`
	SnapshotID     string                      `json:"snapshot_id,omitempty"`
	FallbackReason string                      `json:"fallback_reason,omitempty"`
	DurationMS     uint32                      `json:"duration_ms"`
	EventTime      uint64                      `json:"event_time"`
	ExpireTime     uint64                      `json:"expire_time"`
	Event          *json.RawMessage            `json:"event,omitempty"`
	Decision       *json.RawMessage            `json:"decision,omitempty"`
	Explain        *json.RawMessage            `json:"explain,omitempty"`
	ErrorMessage   string                      `json:"error_message,omitempty"`
	Indexes        []TrafficEventIndexResponse `json:"indexes,omitempty"`
}

type ListTrafficEventsResponse struct {
	Items []TrafficEventResponse `json:"items"`
	Total uint64                 `json:"total"`
	Stats bo.TrafficStats        `json:"stats"`
}

func NewListTrafficEventsResponse(list bo.TrafficEventList) ListTrafficEventsResponse {
	items := make([]TrafficEventResponse, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, newTrafficEventResponse(item, false, false))
	}
	return ListTrafficEventsResponse{
		Items: items,
		Total: list.Total,
		Stats: list.Stats,
	}
}

func NewTrafficEventDetailResponse(item bo.TrafficEvent) TrafficEventResponse {
	return newTrafficEventResponse(item, true, true)
}

func newTrafficEventResponse(item bo.TrafficEvent, includePayload bool, includeIndexes bool) TrafficEventResponse {
	indexes := make([]TrafficEventIndexResponse, 0, len(item.Indexes))
	if includeIndexes {
		for _, index := range item.Indexes {
			indexes = append(indexes, TrafficEventIndexResponse{
				ID:                index.ID,
				TrafficEventID:    index.TrafficEventID,
				ProtocolName:      index.ProtocolName,
				FieldPath:         index.FieldPath,
				FieldValuePreview: index.FieldValuePreview,
				FieldValueHash:    index.FieldValueHash,
				FieldValueText:    index.FieldValueText,
				EventTime:         index.EventTime,
			})
		}
	}
	response := TrafficEventResponse{
		ID:             item.ID,
		EventID:        item.EventID,
		TraceID:        item.TraceID,
		ScenarioID:     item.ScenarioID,
		TrafficSource:  item.TrafficSource,
		ProtocolName:   item.ProtocolName,
		NamespaceID:    item.NamespaceID,
		OperationName:  item.OperationName,
		Outcome:        item.Outcome,
		DecisionKind:   item.DecisionKind,
		RuleSetID:      item.RuleSetID,
		RuleID:         item.RuleID,
		SnapshotID:     item.SnapshotID,
		FallbackReason: item.FallbackReason,
		DurationMS:     item.DurationMS,
		EventTime:      item.EventTime,
		ExpireTime:     item.ExpireTime,
		ErrorMessage:   item.ErrorMessage,
		Indexes:        indexes,
	}
	if includePayload {
		response.Event = safeRawJSONPtr(item.EventJSON)
		response.Decision = safeRawJSONPtr(item.DecisionJSON)
		response.Explain = safeRawJSONPtr(item.ExplainJSON)
	}
	return response
}

func safeRawJSONPtr(raw string) *json.RawMessage {
	if raw == "" || !json.Valid([]byte(raw)) {
		value := json.RawMessage(`{}`)
		return &value
	}
	value := json.RawMessage(raw)
	return &value
}

type ListProtocolsResponse struct {
	Items []mockprotocol.ProtocolSpec `json:"items"`
	Total int                         `json:"total"`
}
