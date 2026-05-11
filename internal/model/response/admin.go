package response

import (
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/mockprotocol"
)

type RuleSetResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Enabled   bool        `json:"enabled"`
	Protocol  string      `json:"protocol"`
	Namespace string      `json:"namespace"`
	Selector  bo.Selector `json:"selector"`
	Rules     []bo.Rule   `json:"rules"`
	Version   int         `json:"version"`
}

type ListRuleSetsResponse struct {
	Items []RuleSetResponse `json:"items"`
	Total int               `json:"total"`
}

type NamespaceResponse struct {
	ID                string                     `json:"id"`
	Name              string                     `json:"name,omitempty"`
	Description       string                     `json:"description,omitempty"`
	RulesetMissAction bo.NamespaceFallbackAction `json:"ruleset_miss_action"`
	RuleMissAction    bo.NamespaceFallbackAction `json:"rule_miss_action"`
}

type ListNamespacesResponse struct {
	Items []NamespaceResponse `json:"items"`
	Total int                 `json:"total"`
}

type PublishedRuleSetResponse struct {
	SnapshotID  string          `json:"snapshot_id"`
	PublishedAt string          `json:"published_at"`
	RuleSet     RuleSetResponse `json:"ruleset"`
	Audit       *bo.AuditInfo   `json:"audit,omitempty"`
}

type ListPublishedRuleSetsResponse struct {
	Items []PublishedRuleSetResponse `json:"items"`
	Total int                        `json:"total"`
}

type ValidateRuleSetResponse struct {
	Result bo.ValidationResult `json:"result"`
}

type PublishRuleSetResponse struct {
	RuleSet  RuleSetResponse          `json:"ruleset"`
	Snapshot PublishedRuleSetResponse `json:"snapshot"`
}

type RollbackRuleSetResponse struct {
	RuleSet  RuleSetResponse          `json:"ruleset"`
	Snapshot PublishedRuleSetResponse `json:"snapshot"`
}

type RollbackPreviewRuleSetResponse struct {
	Result bo.RollbackPreviewResult `json:"result"`
}

type SimulateRuleSetResponse struct {
	Result bo.SimulationResult `json:"result"`
}

type DecidePublishedResponse struct {
	Decision bo.RuntimeDecision `json:"decision"`
}

type ListProtocolsResponse struct {
	Items []mockprotocol.ProtocolSpec `json:"items"`
	Total int                         `json:"total"`
}
