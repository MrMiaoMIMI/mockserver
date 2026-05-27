package modeldo

type Scenario struct {
	CommonDo
	ScenarioCode string `gorm:"column:scenario_code"`
	ScenarioName string `gorm:"column:scenario_name"`
	Status       string `gorm:"column:status"`
	ExpireTime   uint64 `gorm:"column:expire_time"`
	Version      int    `gorm:"column:version"`
	ScenarioJSON string `gorm:"column:scenario_json"`
}

func (*Scenario) TableName() string {
	return "mockserver_scenario_tab"
}

func (*Scenario) IdFieldName() string {
	return "id"
}

type ScenarioRule struct {
	CommonDo
	ScenarioID    uint64 `gorm:"column:scenario_id"`
	ScenarioCode  string `gorm:"column:scenario_code"`
	RuleCode      string `gorm:"column:rule_code"`
	RuleName      string `gorm:"column:rule_name"`
	ProtocolName  string `gorm:"column:protocol_name"`
	NamespaceCode string `gorm:"column:namespace_code"`
	Enabled       bool   `gorm:"column:enabled"`
	Priority      int    `gorm:"column:priority"`
	ExpireTime    uint64 `gorm:"column:expire_time"`
	Version       int    `gorm:"column:version"`
	RuleJSON      string `gorm:"column:rule_json"`
}

func (*ScenarioRule) TableName() string {
	return "mockserver_scenario_rule_tab"
}

func (*ScenarioRule) IdFieldName() string {
	return "id"
}
