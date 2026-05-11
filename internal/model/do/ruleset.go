package modeldo

type RuleSetDraft struct {
	CommonDo
	RuleSetID   string `gorm:"column:ruleset_id"`
	Version     int    `gorm:"column:version"`
	RuleSetJSON string `gorm:"column:ruleset_json"`
}

func (*RuleSetDraft) TableName() string {
	return "mockserver_rule_set_draft_tab"
}

func (*RuleSetDraft) IdFieldName() string {
	return "ruleset_id"
}

type PublishedRuleSetCurrent struct {
	CommonDo
	RuleSetID         string `gorm:"column:ruleset_id"`
	CurrentSnapshotID string `gorm:"column:current_snapshot_id"`
}

func (*PublishedRuleSetCurrent) TableName() string {
	return "mockserver_published_rule_set_tab"
}

func (*PublishedRuleSetCurrent) IdFieldName() string {
	return "ruleset_id"
}

type PublishedRuleSetSnapshot struct {
	CommonDo
	SnapshotID     string `gorm:"column:snapshot_id"`
	RuleSetID      string `gorm:"column:ruleset_id"`
	RuleSetVersion int    `gorm:"column:ruleset_version"`
	RuleSetJSON    string `gorm:"column:ruleset_json"`
	AuditJSON      string `gorm:"column:audit_json"`
	PublishTime    uint64 `gorm:"column:publish_time"`
}

func (*PublishedRuleSetSnapshot) TableName() string {
	return "mockserver_published_snapshot_tab"
}

func (*PublishedRuleSetSnapshot) IdFieldName() string {
	return "snapshot_id"
}

type NamespaceConfig struct {
	CommonDo
	NamespaceID   string `gorm:"column:namespace_id"`
	Version       int    `gorm:"column:version"`
	NamespaceJSON string `gorm:"column:namespace_json"`
}

func (*NamespaceConfig) TableName() string {
	return "mockserver_namespace_tab"
}

func (*NamespaceConfig) IdFieldName() string {
	return "namespace_id"
}
