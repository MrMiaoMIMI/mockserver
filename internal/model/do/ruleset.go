package modeldo

type RuleSetDraft struct {
	CommonDo
	RuleSetCode   string `gorm:"column:ruleset_code"`
	RuleSetName   string `gorm:"column:ruleset_name"`
	ProtocolName  string `gorm:"column:protocol_name"`
	NamespaceCode string `gorm:"column:namespace_code"`
	Version       int    `gorm:"column:version"`
	RuleSetJSON   string `gorm:"column:ruleset_json"`
}

func (*RuleSetDraft) TableName() string {
	return "mockserver_rule_set_draft_tab"
}

func (*RuleSetDraft) IdFieldName() string {
	return "id"
}

type PublishedRuleSetCurrent struct {
	CommonDo
	RuleSetID         uint64 `gorm:"column:ruleset_id"`
	CurrentSnapshotID uint64 `gorm:"column:current_snapshot_id"`
}

func (*PublishedRuleSetCurrent) TableName() string {
	return "mockserver_published_rule_set_tab"
}

func (*PublishedRuleSetCurrent) IdFieldName() string {
	return "id"
}

type PublishedRuleSetSnapshot struct {
	CommonDo
	SnapshotCode   string `gorm:"column:snapshot_code"`
	RuleSetID      uint64 `gorm:"column:ruleset_id"`
	RuleSetVersion int    `gorm:"column:ruleset_version"`
	RuleSetJSON    string `gorm:"column:ruleset_json"`
	AuditJSON      string `gorm:"column:audit_json"`
	PublishTime    uint64 `gorm:"column:publish_time"`
}

func (*PublishedRuleSetSnapshot) TableName() string {
	return "mockserver_published_snapshot_tab"
}

func (*PublishedRuleSetSnapshot) IdFieldName() string {
	return "id"
}

type NamespaceConfig struct {
	CommonDo
	NamespaceCode string `gorm:"column:namespace_code"`
	NamespaceName string `gorm:"column:namespace_name"`
	Version       int    `gorm:"column:version"`
	NamespaceJSON string `gorm:"column:namespace_json"`
}

func (*NamespaceConfig) TableName() string {
	return "mockserver_namespace_tab"
}

func (*NamespaceConfig) IdFieldName() string {
	return "id"
}
