package request

import "github.com/MrMiaoMIMI/mockserver/internal/model/bo"

type UpsertRuleSetRequest struct {
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name"`
	Enabled   bool        `json:"enabled"`
	Protocol  string      `json:"protocol"`
	Namespace string      `json:"namespace"`
	Selector  bo.Selector `json:"selector"`
	Rules     []bo.Rule   `json:"rules"`
}

type UpsertNamespaceRequest struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name,omitempty"`
	Description string                        `json:"description,omitempty"`
	Policies    map[string]bo.NamespacePolicy `json:"policies"`
}

type SimulateRuleSetRequest struct {
	Event           bo.Event    `json:"event"`
	DraftOverride   *bo.RuleSet `json:"draft_override,omitempty"`
	ExplainOnly     bool        `json:"explain_only,omitempty"`
	ExplainMaxDepth int         `json:"explain_max_depth,omitempty"`
	ExplainCompact  bool        `json:"explain_compact,omitempty"`
	ExplainSummary  bool        `json:"explain_summary,omitempty"`
}

type DecidePublishedRequest struct {
	Event bo.Event `json:"event"`
}

type DraftRuleRequest struct {
	Rule bo.Rule `json:"rule"`
}

type UpdateRulePriorityRequest struct {
	Priority int `json:"priority"`
}

type PublishRuleSetRequest struct {
	Reason string       `json:"reason,omitempty"`
	Audit  bo.AuditInfo `json:"-"`
}

type RollbackRuleSetRequest struct {
	SnapshotID string       `json:"snapshot_id"`
	Reason     string       `json:"reason,omitempty"`
	Audit      bo.AuditInfo `json:"-"`
}

type RollbackPreviewRuleSetRequest struct {
	SnapshotID      string    `json:"snapshot_id"`
	Event           *bo.Event `json:"event,omitempty"`
	ExplainOnly     bool      `json:"explain_only,omitempty"`
	ExplainMaxDepth int       `json:"explain_max_depth,omitempty"`
	ExplainCompact  bool      `json:"explain_compact,omitempty"`
	ExplainSummary  bool      `json:"explain_summary,omitempty"`
}
