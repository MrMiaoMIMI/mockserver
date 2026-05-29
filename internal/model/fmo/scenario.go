package modelfmo

import (
	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

type ScenarioFields struct {
	CommonFmo
	ScenarioCode dbspi.Field[string]
	ScenarioName dbspi.Field[string]
	Status       dbspi.Field[string]
	ExpireTime   dbspi.Field[uint64]
	Version      dbspi.Field[int]
	ScenarioJSON dbspi.Field[string]
}

func NewScenarioFields() ScenarioFields {
	return ScenarioFields{
		CommonFmo:    NewCommonFmo(),
		ScenarioCode: dbhelper.NewField[string]("scenario_code"),
		ScenarioName: dbhelper.NewField[string]("scenario_name"),
		Status:       dbhelper.NewField[string]("status"),
		ExpireTime:   dbhelper.NewField[uint64]("expire_time"),
		Version:      dbhelper.NewField[int]("version"),
		ScenarioJSON: dbhelper.NewField[string]("scenario_json"),
	}
}

type ScenarioRuleFields struct {
	CommonFmo
	ScenarioID    dbspi.Field[uint64]
	ScenarioCode  dbspi.Field[string]
	RuleCode      dbspi.Field[string]
	RuleName      dbspi.Field[string]
	ProtocolName  dbspi.Field[string]
	NamespaceName dbspi.Field[string]
	Enabled       dbspi.Field[bool]
	Priority      dbspi.Field[int]
	ExpireTime    dbspi.Field[uint64]
	Version       dbspi.Field[int]
	RuleJSON      dbspi.Field[string]
}

func NewScenarioRuleFields() ScenarioRuleFields {
	return ScenarioRuleFields{
		CommonFmo:     NewCommonFmo(),
		ScenarioID:    dbhelper.NewField[uint64]("scenario_id"),
		ScenarioCode:  dbhelper.NewField[string]("scenario_code"),
		RuleCode:      dbhelper.NewField[string]("rule_code"),
		RuleName:      dbhelper.NewField[string]("rule_name"),
		ProtocolName:  dbhelper.NewField[string]("protocol_name"),
		NamespaceName: dbhelper.NewField[string]("namespace_name"),
		Enabled:       dbhelper.NewField[bool]("enabled"),
		Priority:      dbhelper.NewField[int]("priority"),
		ExpireTime:    dbhelper.NewField[uint64]("expire_time"),
		Version:       dbhelper.NewField[int]("version"),
		RuleJSON:      dbhelper.NewField[string]("rule_json"),
	}
}
