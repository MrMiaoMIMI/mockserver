package modelfmo

import (
	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

type RuleSetDraftFields struct {
	CommonFmo
	RuleSetID   dbspi.Field[string]
	Version     dbspi.Field[int]
	RuleSetJSON dbspi.Field[string]
}

func NewRuleSetDraftFields() RuleSetDraftFields {
	return RuleSetDraftFields{
		CommonFmo:   NewCommonFmo(),
		RuleSetID:   dbhelper.NewField[string]("ruleset_id"),
		Version:     dbhelper.NewField[int]("version"),
		RuleSetJSON: dbhelper.NewField[string]("ruleset_json"),
	}
}

type PublishedRuleSetCurrentFields struct {
	CommonFmo
	RuleSetID         dbspi.Field[string]
	CurrentSnapshotID dbspi.Field[string]
}

func NewPublishedRuleSetCurrentFields() PublishedRuleSetCurrentFields {
	return PublishedRuleSetCurrentFields{
		CommonFmo:         NewCommonFmo(),
		RuleSetID:         dbhelper.NewField[string]("ruleset_id"),
		CurrentSnapshotID: dbhelper.NewField[string]("current_snapshot_id"),
	}
}

type PublishedRuleSetSnapshotFields struct {
	CommonFmo
	SnapshotID     dbspi.Field[string]
	RuleSetID      dbspi.Field[string]
	RuleSetVersion dbspi.Field[int]
	RuleSetJSON    dbspi.Field[string]
	AuditJSON      dbspi.Field[string]
	PublishTime    dbspi.Field[uint64]
}

func NewPublishedRuleSetSnapshotFields() PublishedRuleSetSnapshotFields {
	return PublishedRuleSetSnapshotFields{
		CommonFmo:      NewCommonFmo(),
		SnapshotID:     dbhelper.NewField[string]("snapshot_id"),
		RuleSetID:      dbhelper.NewField[string]("ruleset_id"),
		RuleSetVersion: dbhelper.NewField[int]("ruleset_version"),
		RuleSetJSON:    dbhelper.NewField[string]("ruleset_json"),
		AuditJSON:      dbhelper.NewField[string]("audit_json"),
		PublishTime:    dbhelper.NewField[uint64]("publish_time"),
	}
}

type NamespaceConfigFields struct {
	CommonFmo
	NamespaceID   dbspi.Field[string]
	Version       dbspi.Field[int]
	NamespaceJSON dbspi.Field[string]
}

func NewNamespaceConfigFields() NamespaceConfigFields {
	return NamespaceConfigFields{
		CommonFmo:     NewCommonFmo(),
		NamespaceID:   dbhelper.NewField[string]("namespace_id"),
		Version:       dbhelper.NewField[int]("version"),
		NamespaceJSON: dbhelper.NewField[string]("namespace_json"),
	}
}
