package service

import (
	"context"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type RuleSetService interface {
	UpsertDraft(ctx context.Context, ruleSet bo.RuleSet) (bo.RuleSet, error)
	GetDraft(ctx context.Context, id string) (bo.RuleSet, error)
	ListDrafts(ctx context.Context) ([]bo.RuleSet, error)
	GetPublished(ctx context.Context, id string) (bo.PublishedRuleSetSnapshot, error)
	ListPublished(ctx context.Context) ([]bo.PublishedRuleSetSnapshot, error)
	ListPublishedSnapshots(ctx context.Context, id string) ([]bo.PublishedRuleSetSnapshot, error)
	ValidateDraft(ctx context.Context, id string) (bo.ValidationResult, error)
	AddDraftRule(ctx context.Context, id string, rule bo.Rule) (bo.RuleSet, error)
	UpdateDraftRule(ctx context.Context, id string, ruleID string, rule bo.Rule) (bo.RuleSet, error)
	DeleteDraftRule(ctx context.Context, id string, ruleID string) (bo.RuleSet, error)
	SetDraftRuleEnabled(ctx context.Context, id string, ruleID string, enabled bool) (bo.RuleSet, error)
	SetDraftRulePriority(ctx context.Context, id string, ruleID string, priority int) (bo.RuleSet, error)
	RollbackPreview(ctx context.Context, id, snapshotID string, event *bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.RollbackPreviewResult, error)
	Rollback(ctx context.Context, id, snapshotID string, audit bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error)
	Publish(ctx context.Context, id string, audit bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error)
	SimulateDraft(ctx context.Context, id string, event bo.Event, draftOverride *bo.RuleSet, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error)
	SimulatePublished(ctx context.Context, event bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error)
}

type NamespaceService interface {
	UpsertNamespace(ctx context.Context, namespace bo.Namespace) (bo.Namespace, error)
	EnsureDefaultNamespace(ctx context.Context) (bo.Namespace, error)
	GetNamespace(ctx context.Context, id string) (bo.Namespace, error)
	ListNamespaces(ctx context.Context) ([]bo.Namespace, error)
}

type RuntimeService interface {
	MatchPublished(ctx context.Context, event bo.Event) (bo.SimulationResult, error)
	DecidePublished(ctx context.Context, event bo.Event) (bo.RuntimeDecision, error)
}

type TrafficService interface {
	RecordSDKDecision(ctx context.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, durationMS uint32) (bo.TrafficEvent, error)
	ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error)
}
