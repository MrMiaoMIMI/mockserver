package request

import "github.com/MrMiaoMIMI/mockserver/internal/model/bo"

type CreateScenarioRequest struct {
	ScenarioID  string `json:"scenario_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TTLSeconds  uint64 `json:"ttl_seconds,omitempty"`
}

type UpdateScenarioRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	TTLSeconds  *uint64 `json:"ttl_seconds,omitempty"`
	ExpireTime  *uint64 `json:"expire_time,omitempty"`
}

type UpsertScenarioRuleRequest struct {
	Protocol string  `json:"protocol"`
	Rule     bo.Rule `json:"rule"`
}

type HTTPQuickRuleRequest struct {
	RuleID    string               `json:"rule_id,omitempty"`
	Name      string               `json:"name,omitempty"`
	Namespace string               `json:"namespace,omitempty"`
	Enabled   *bool                `json:"enabled,omitempty"`
	Priority  int                  `json:"priority,omitempty"`
	Match     bo.HTTPQuickMatch    `json:"match"`
	Respond   bo.HTTPQuickResponse `json:"respond"`
}

type SimulateScenarioRequest struct {
	Event           bo.Event `json:"event"`
	ExplainOnly     bool     `json:"explain_only,omitempty"`
	ExplainMaxDepth int      `json:"explain_max_depth,omitempty"`
	ExplainCompact  bool     `json:"explain_compact,omitempty"`
	ExplainSummary  bool     `json:"explain_summary,omitempty"`
}
