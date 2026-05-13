package modeldo

type TrafficEvent struct {
	CommonDo
	EventID        string `gorm:"column:event_id"`
	TraceID        string `gorm:"column:trace_id"`
	TrafficSource  string `gorm:"column:traffic_source"`
	ProtocolName   string `gorm:"column:protocol_name"`
	NamespaceID    string `gorm:"column:namespace_id"`
	OperationName  string `gorm:"column:operation_name"`
	Outcome        string `gorm:"column:outcome"`
	DecisionKind   string `gorm:"column:decision_kind"`
	RuleSetID      string `gorm:"column:ruleset_id"`
	RuleID         string `gorm:"column:rule_id"`
	SnapshotID     string `gorm:"column:snapshot_id"`
	FallbackReason string `gorm:"column:fallback_reason"`
	DurationMS     uint32 `gorm:"column:duration_ms"`
	EventTime      uint64 `gorm:"column:event_time"`
	ExpireTime     uint64 `gorm:"column:expire_time"`
	EventJSON      string `gorm:"column:event_json"`
	DecisionJSON   string `gorm:"column:decision_json"`
	ExplainJSON    string `gorm:"column:explain_json"`
	ErrorMessage   string `gorm:"column:error_message"`
}

func (*TrafficEvent) TableName() string {
	return "mockserver_traffic_event_tab"
}

func (*TrafficEvent) IdFieldName() string {
	return "id"
}

type TrafficEventIndex struct {
	CommonDo
	TrafficEventID    uint64 `gorm:"column:traffic_event_id"`
	EventID           string `gorm:"column:event_id"`
	ProtocolName      string `gorm:"column:protocol_name"`
	FieldPath         string `gorm:"column:field_path"`
	FieldValuePreview string `gorm:"column:field_value_preview"`
	FieldValueHash    uint64 `gorm:"column:field_value_hash"`
	FieldValueText    string `gorm:"column:field_value_text"`
	EventTime         uint64 `gorm:"column:event_time"`
	ExpireTime        uint64 `gorm:"column:expire_time"`
}

func (*TrafficEventIndex) TableName() string {
	return "mockserver_traffic_event_index_tab"
}

func (*TrafficEventIndex) IdFieldName() string {
	return "id"
}
