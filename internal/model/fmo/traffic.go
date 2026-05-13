package modelfmo

import (
	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

type TrafficEventFields struct {
	CommonFmo
	EventID        dbspi.Field[string]
	TraceID        dbspi.Field[string]
	TrafficSource  dbspi.Field[string]
	ProtocolName   dbspi.Field[string]
	NamespaceID    dbspi.Field[string]
	OperationName  dbspi.Field[string]
	Outcome        dbspi.Field[string]
	DecisionKind   dbspi.Field[string]
	RuleSetID      dbspi.Field[string]
	RuleID         dbspi.Field[string]
	SnapshotID     dbspi.Field[string]
	FallbackReason dbspi.Field[string]
	DurationMS     dbspi.Field[uint32]
	EventTime      dbspi.Field[uint64]
	ExpireTime     dbspi.Field[uint64]
	EventJSON      dbspi.Field[string]
	DecisionJSON   dbspi.Field[string]
	ExplainJSON    dbspi.Field[string]
	ErrorMessage   dbspi.Field[string]
}

func NewTrafficEventFields() TrafficEventFields {
	return TrafficEventFields{
		CommonFmo:      NewCommonFmo(),
		EventID:        dbhelper.NewField[string]("event_id"),
		TraceID:        dbhelper.NewField[string]("trace_id"),
		TrafficSource:  dbhelper.NewField[string]("traffic_source"),
		ProtocolName:   dbhelper.NewField[string]("protocol_name"),
		NamespaceID:    dbhelper.NewField[string]("namespace_id"),
		OperationName:  dbhelper.NewField[string]("operation_name"),
		Outcome:        dbhelper.NewField[string]("outcome"),
		DecisionKind:   dbhelper.NewField[string]("decision_kind"),
		RuleSetID:      dbhelper.NewField[string]("ruleset_id"),
		RuleID:         dbhelper.NewField[string]("rule_id"),
		SnapshotID:     dbhelper.NewField[string]("snapshot_id"),
		FallbackReason: dbhelper.NewField[string]("fallback_reason"),
		DurationMS:     dbhelper.NewField[uint32]("duration_ms"),
		EventTime:      dbhelper.NewField[uint64]("event_time"),
		ExpireTime:     dbhelper.NewField[uint64]("expire_time"),
		EventJSON:      dbhelper.NewField[string]("event_json"),
		DecisionJSON:   dbhelper.NewField[string]("decision_json"),
		ExplainJSON:    dbhelper.NewField[string]("explain_json"),
		ErrorMessage:   dbhelper.NewField[string]("error_message"),
	}
}

type TrafficEventIndexFields struct {
	CommonFmo
	TrafficEventID    dbspi.Field[uint64]
	EventID           dbspi.Field[string]
	ProtocolName      dbspi.Field[string]
	FieldPath         dbspi.Field[string]
	FieldValuePreview dbspi.Field[string]
	FieldValueHash    dbspi.Field[uint64]
	FieldValueText    dbspi.Field[string]
	EventTime         dbspi.Field[uint64]
	ExpireTime        dbspi.Field[uint64]
}

func NewTrafficEventIndexFields() TrafficEventIndexFields {
	return TrafficEventIndexFields{
		CommonFmo:         NewCommonFmo(),
		TrafficEventID:    dbhelper.NewField[uint64]("traffic_event_id"),
		EventID:           dbhelper.NewField[string]("event_id"),
		ProtocolName:      dbhelper.NewField[string]("protocol_name"),
		FieldPath:         dbhelper.NewField[string]("field_path"),
		FieldValuePreview: dbhelper.NewField[string]("field_value_preview"),
		FieldValueHash:    dbhelper.NewField[uint64]("field_value_hash"),
		FieldValueText:    dbhelper.NewField[string]("field_value_text"),
		EventTime:         dbhelper.NewField[uint64]("event_time"),
		ExpireTime:        dbhelper.NewField[uint64]("expire_time"),
	}
}
