package modelfmo

import (
	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

type RuleSetDraftFields struct {
	CommonFmo
	RuleSetCode dbspi.Field[string]
	Version     dbspi.Field[int]
	RuleSetJSON dbspi.Field[string]
}

func NewRuleSetDraftFields() RuleSetDraftFields {
	return RuleSetDraftFields{
		CommonFmo:   NewCommonFmo(),
		RuleSetCode: dbhelper.NewField[string]("ruleset_code"),
		Version:     dbhelper.NewField[int]("version"),
		RuleSetJSON: dbhelper.NewField[string]("ruleset_json"),
	}
}

type PublishedRuleSetCurrentFields struct {
	CommonFmo
	RuleSetID         dbspi.Field[uint64]
	CurrentSnapshotID dbspi.Field[uint64]
}

func NewPublishedRuleSetCurrentFields() PublishedRuleSetCurrentFields {
	return PublishedRuleSetCurrentFields{
		CommonFmo:         NewCommonFmo(),
		RuleSetID:         dbhelper.NewField[uint64]("ruleset_id"),
		CurrentSnapshotID: dbhelper.NewField[uint64]("current_snapshot_id"),
	}
}

type PublishedRuleSetSnapshotFields struct {
	CommonFmo
	SnapshotCode   dbspi.Field[string]
	RuleSetID      dbspi.Field[uint64]
	RuleSetVersion dbspi.Field[int]
	RuleSetJSON    dbspi.Field[string]
	AuditJSON      dbspi.Field[string]
	PublishTime    dbspi.Field[uint64]
}

func NewPublishedRuleSetSnapshotFields() PublishedRuleSetSnapshotFields {
	return PublishedRuleSetSnapshotFields{
		CommonFmo:      NewCommonFmo(),
		SnapshotCode:   dbhelper.NewField[string]("snapshot_code"),
		RuleSetID:      dbhelper.NewField[uint64]("ruleset_id"),
		RuleSetVersion: dbhelper.NewField[int]("ruleset_version"),
		RuleSetJSON:    dbhelper.NewField[string]("ruleset_json"),
		AuditJSON:      dbhelper.NewField[string]("audit_json"),
		PublishTime:    dbhelper.NewField[uint64]("publish_time"),
	}
}

type NamespaceConfigFields struct {
	CommonFmo
	NamespaceCode dbspi.Field[string]
	Version       dbspi.Field[int]
	NamespaceJSON dbspi.Field[string]
}

func NewNamespaceConfigFields() NamespaceConfigFields {
	return NamespaceConfigFields{
		CommonFmo:     NewCommonFmo(),
		NamespaceCode: dbhelper.NewField[string]("namespace_code"),
		Version:       dbhelper.NewField[int]("version"),
		NamespaceJSON: dbhelper.NewField[string]("namespace_json"),
	}
}
