package modelfmo

import (
	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

type TrafficEventFields struct {
	CommonFmo
	EventCode      dbspi.Field[string]
	TraceID        dbspi.Field[string]
	ScenarioCode   dbspi.Field[string]
	TrafficSource  dbspi.Field[string]
	ProtocolName   dbspi.Field[string]
	NamespaceID    dbspi.Field[uint64]
	NamespaceName  dbspi.Field[string]
	OperationName  dbspi.Field[string]
	Outcome        dbspi.Field[string]
	DecisionKind   dbspi.Field[string]
	RuleSetID      dbspi.Field[uint64]
	RuleSetCode    dbspi.Field[string]
	RuleCode       dbspi.Field[string]
	SnapshotID     dbspi.Field[uint64]
	SnapshotCode   dbspi.Field[string]
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
		EventCode:      dbhelper.NewField[string]("event_code"),
		TraceID:        dbhelper.NewField[string]("trace_id"),
		ScenarioCode:   dbhelper.NewField[string]("scenario_code"),
		TrafficSource:  dbhelper.NewField[string]("traffic_source"),
		ProtocolName:   dbhelper.NewField[string]("protocol_name"),
		NamespaceID:    dbhelper.NewField[uint64]("namespace_id"),
		NamespaceName:  dbhelper.NewField[string]("namespace_name"),
		OperationName:  dbhelper.NewField[string]("operation_name"),
		Outcome:        dbhelper.NewField[string]("outcome"),
		DecisionKind:   dbhelper.NewField[string]("decision_kind"),
		RuleSetID:      dbhelper.NewField[uint64]("ruleset_id"),
		RuleSetCode:    dbhelper.NewField[string]("ruleset_code"),
		RuleCode:       dbhelper.NewField[string]("rule_code"),
		SnapshotID:     dbhelper.NewField[uint64]("snapshot_id"),
		SnapshotCode:   dbhelper.NewField[string]("snapshot_code"),
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
		ProtocolName:      dbhelper.NewField[string]("protocol_name"),
		FieldPath:         dbhelper.NewField[string]("field_path"),
		FieldValuePreview: dbhelper.NewField[string]("field_value_preview"),
		FieldValueHash:    dbhelper.NewField[uint64]("field_value_hash"),
		FieldValueText:    dbhelper.NewField[string]("field_value_text"),
		EventTime:         dbhelper.NewField[uint64]("event_time"),
		ExpireTime:        dbhelper.NewField[uint64]("expire_time"),
	}
}
