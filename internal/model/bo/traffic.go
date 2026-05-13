package bo

const (
	TrafficSourceSDKDecision = "sdk_decision"

	TrafficOutcomeMatched   = "matched"
	TrafficOutcomeFallback  = "fallback"
	TrafficOutcomeUnmatched = "unmatched"
	TrafficOutcomeError     = "error"
)

type TrafficEvent struct {
	ID             uint64              `json:"id"`
	EventID        string              `json:"event_id"`
	TraceID        string              `json:"trace_id,omitempty"`
	TrafficSource  string              `json:"traffic_source"`
	ProtocolName   string              `json:"protocol_name"`
	NamespaceID    string              `json:"namespace_id"`
	OperationName  string              `json:"operation_name,omitempty"`
	Outcome        string              `json:"outcome"`
	DecisionKind   string              `json:"decision_kind,omitempty"`
	RuleSetID      string              `json:"ruleset_id,omitempty"`
	RuleID         string              `json:"rule_id,omitempty"`
	SnapshotID     string              `json:"snapshot_id,omitempty"`
	FallbackReason string              `json:"fallback_reason,omitempty"`
	DurationMS     uint32              `json:"duration_ms"`
	EventTime      uint64              `json:"event_time"`
	ExpireTime     uint64              `json:"expire_time"`
	EventJSON      string              `json:"-"`
	DecisionJSON   string              `json:"-"`
	ExplainJSON    string              `json:"-"`
	ErrorMessage   string              `json:"error_message,omitempty"`
	Indexes        []TrafficEventIndex `json:"indexes,omitempty"`
}

type TrafficEventIndex struct {
	ID                uint64 `json:"id"`
	TrafficEventID    uint64 `json:"traffic_event_id"`
	EventID           string `json:"event_id"`
	ProtocolName      string `json:"protocol_name"`
	FieldPath         string `json:"field_path"`
	FieldValuePreview string `json:"field_value_preview"`
	FieldValueHash    uint64 `json:"field_value_hash"`
	FieldValueText    string `json:"field_value_text"`
	EventTime         uint64 `json:"event_time"`
	ExpireTime        uint64 `json:"expire_time"`
}

type TrafficIndexFilter struct {
	FieldPath  string
	FieldValue string
}

type TrafficQuery struct {
	Limit          int
	Offset         int
	StartTime      uint64
	EndTime        uint64
	TraceID        string
	TrafficSource  string
	ProtocolName   string
	NamespaceID    string
	OperationName  string
	Outcome        string
	DecisionKind   string
	RuleSetID      string
	RuleID         string
	FallbackReason string
	IndexFilters   []TrafficIndexFilter
	IncludeIndexes bool
}

type TrafficStats struct {
	Total       uint64            `json:"total"`
	ByOutcome   map[string]uint64 `json:"by_outcome"`
	ByProtocol  map[string]uint64 `json:"by_protocol"`
	ByNamespace map[string]uint64 `json:"by_namespace"`
}

type TrafficEventList struct {
	Items []TrafficEvent `json:"items"`
	Total uint64         `json:"total"`
	Stats TrafficStats   `json:"stats"`
}
