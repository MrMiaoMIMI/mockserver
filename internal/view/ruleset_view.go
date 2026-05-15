package view

import (
	"context"
	"time"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/request"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
)

type RuleSetView interface {
	UpsertDraft(ctx context.Context, req request.UpsertRuleSetRequest) (response.RuleSetResponse, error)
	GetDraft(ctx context.Context, id string) (response.RuleSetResponse, error)
	ListDrafts(ctx context.Context) (response.ListRuleSetsResponse, error)
	GetPublished(ctx context.Context, id string) (response.PublishedRuleSetResponse, error)
	ListPublished(ctx context.Context) (response.ListPublishedRuleSetsResponse, error)
	ListPublishedSnapshots(ctx context.Context, id string) (response.ListPublishedRuleSetsResponse, error)
	ValidateDraft(ctx context.Context, id string) (response.ValidateRuleSetResponse, error)
	Publish(ctx context.Context, id string, req request.PublishRuleSetRequest) (response.PublishRuleSetResponse, error)
	AddDraftRule(ctx context.Context, id string, req request.DraftRuleRequest) (response.RuleSetResponse, error)
	UpdateDraftRule(ctx context.Context, id string, ruleID string, req request.DraftRuleRequest) (response.RuleSetResponse, error)
	DeleteDraftRule(ctx context.Context, id string, ruleID string) (response.RuleSetResponse, error)
	SetDraftRuleEnabled(ctx context.Context, id string, ruleID string, enabled bool) (response.RuleSetResponse, error)
	SetDraftRulePriority(ctx context.Context, id string, ruleID string, req request.UpdateRulePriorityRequest) (response.RuleSetResponse, error)
	RollbackPreview(ctx context.Context, id string, req request.RollbackPreviewRuleSetRequest) (response.RollbackPreviewRuleSetResponse, error)
	Rollback(ctx context.Context, id string, req request.RollbackRuleSetRequest) (response.RollbackRuleSetResponse, error)
	SimulateDraft(ctx context.Context, id string, req request.SimulateRuleSetRequest) (response.SimulateRuleSetResponse, error)
	SimulatePublished(ctx context.Context, req request.SimulateRuleSetRequest) (response.SimulateRuleSetResponse, error)
}

type NamespaceView interface {
	UpsertNamespace(ctx context.Context, req request.UpsertNamespaceRequest) (response.NamespaceResponse, error)
	GetNamespace(ctx context.Context, id string) (response.NamespaceResponse, error)
	ListNamespaces(ctx context.Context) (response.ListNamespacesResponse, error)
}

type ruleSetView struct {
	service service.RuleSetService
}

func NewRuleSetView(ruleSetService service.RuleSetService) RuleSetView {
	return &ruleSetView{service: ruleSetService}
}

type namespaceView struct {
	service service.NamespaceService
}

func NewNamespaceView(namespaceService service.NamespaceService) NamespaceView {
	return &namespaceView{service: namespaceService}
}

func (v *ruleSetView) UpsertDraft(ctx context.Context, req request.UpsertRuleSetRequest) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.UpsertDraft(ctx, bo.RuleSet{
		ID:        req.ID,
		Name:      req.Name,
		Enabled:   req.Enabled,
		Protocol:  req.Protocol,
		Namespace: req.Namespace,
		Selector:  req.Selector,
		Rules:     req.Rules,
	})
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) GetDraft(ctx context.Context, id string) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.GetDraft(ctx, id)
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) ListDrafts(ctx context.Context) (response.ListRuleSetsResponse, error) {
	items, err := v.service.ListDrafts(ctx)
	if err != nil {
		return response.ListRuleSetsResponse{}, err
	}
	result := response.ListRuleSetsResponse{
		Items: make([]response.RuleSetResponse, 0, len(items)),
		Total: len(items),
	}
	for _, item := range items {
		result.Items = append(result.Items, toRuleSetResponse(item))
	}
	return result, nil
}

func (v *namespaceView) UpsertNamespace(ctx context.Context, req request.UpsertNamespaceRequest) (response.NamespaceResponse, error) {
	namespace, err := v.service.UpsertNamespace(ctx, bo.Namespace{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Policies:    req.Policies,
	})
	if err != nil {
		return response.NamespaceResponse{}, err
	}
	return toNamespaceResponse(namespace), nil
}

func (v *namespaceView) GetNamespace(ctx context.Context, id string) (response.NamespaceResponse, error) {
	namespace, err := v.service.GetNamespace(ctx, id)
	if err != nil {
		return response.NamespaceResponse{}, err
	}
	return toNamespaceResponse(namespace), nil
}

func (v *namespaceView) ListNamespaces(ctx context.Context) (response.ListNamespacesResponse, error) {
	items, err := v.service.ListNamespaces(ctx)
	if err != nil {
		return response.ListNamespacesResponse{}, err
	}
	result := response.ListNamespacesResponse{
		Items: make([]response.NamespaceResponse, 0, len(items)),
		Total: len(items),
	}
	for _, item := range items {
		result.Items = append(result.Items, toNamespaceResponse(item))
	}
	return result, nil
}

func (v *ruleSetView) GetPublished(ctx context.Context, id string) (response.PublishedRuleSetResponse, error) {
	snapshot, err := v.service.GetPublished(ctx, id)
	if err != nil {
		return response.PublishedRuleSetResponse{}, err
	}
	return toPublishedRuleSetResponse(snapshot), nil
}

func (v *ruleSetView) ListPublished(ctx context.Context) (response.ListPublishedRuleSetsResponse, error) {
	items, err := v.service.ListPublished(ctx)
	if err != nil {
		return response.ListPublishedRuleSetsResponse{}, err
	}
	return toListPublishedRuleSetsResponse(items), nil
}

func (v *ruleSetView) ListPublishedSnapshots(ctx context.Context, id string) (response.ListPublishedRuleSetsResponse, error) {
	items, err := v.service.ListPublishedSnapshots(ctx, id)
	if err != nil {
		return response.ListPublishedRuleSetsResponse{}, err
	}
	return toListPublishedRuleSetsResponse(items), nil
}

func (v *ruleSetView) ValidateDraft(ctx context.Context, id string) (response.ValidateRuleSetResponse, error) {
	result, err := v.service.ValidateDraft(ctx, id)
	if err != nil {
		return response.ValidateRuleSetResponse{}, err
	}
	return response.ValidateRuleSetResponse{Result: result}, nil
}

func (v *ruleSetView) Publish(ctx context.Context, id string, req request.PublishRuleSetRequest) (response.PublishRuleSetResponse, error) {
	audit := req.Audit
	audit.Reason = req.Reason
	snapshot, err := v.service.Publish(ctx, id, audit)
	if err != nil {
		return response.PublishRuleSetResponse{}, err
	}
	return response.PublishRuleSetResponse{
		RuleSet:  toRuleSetResponse(snapshot.RuleSet),
		Snapshot: toPublishedRuleSetResponse(snapshot),
	}, nil
}

func (v *ruleSetView) AddDraftRule(ctx context.Context, id string, req request.DraftRuleRequest) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.AddDraftRule(ctx, id, req.Rule)
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) UpdateDraftRule(ctx context.Context, id string, ruleID string, req request.DraftRuleRequest) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.UpdateDraftRule(ctx, id, ruleID, req.Rule)
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) DeleteDraftRule(ctx context.Context, id string, ruleID string) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.DeleteDraftRule(ctx, id, ruleID)
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) SetDraftRuleEnabled(ctx context.Context, id string, ruleID string, enabled bool) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.SetDraftRuleEnabled(ctx, id, ruleID, enabled)
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) SetDraftRulePriority(ctx context.Context, id string, ruleID string, req request.UpdateRulePriorityRequest) (response.RuleSetResponse, error) {
	ruleSet, err := v.service.SetDraftRulePriority(ctx, id, ruleID, req.Priority)
	if err != nil {
		return response.RuleSetResponse{}, err
	}
	return toRuleSetResponse(ruleSet), nil
}

func (v *ruleSetView) RollbackPreview(ctx context.Context, id string, req request.RollbackPreviewRuleSetRequest) (response.RollbackPreviewRuleSetResponse, error) {
	result, err := v.service.RollbackPreview(ctx, id, req.SnapshotID, req.Event, req.ExplainOnly, req.ExplainMaxDepth, req.ExplainCompact, req.ExplainSummary)
	if err != nil {
		return response.RollbackPreviewRuleSetResponse{}, err
	}
	return response.RollbackPreviewRuleSetResponse{Result: result}, nil
}

func (v *ruleSetView) Rollback(ctx context.Context, id string, req request.RollbackRuleSetRequest) (response.RollbackRuleSetResponse, error) {
	audit := req.Audit
	audit.Reason = req.Reason
	snapshot, err := v.service.Rollback(ctx, id, req.SnapshotID, audit)
	if err != nil {
		return response.RollbackRuleSetResponse{}, err
	}
	return response.RollbackRuleSetResponse{
		RuleSet:  toRuleSetResponse(snapshot.RuleSet),
		Snapshot: toPublishedRuleSetResponse(snapshot),
	}, nil
}

func (v *ruleSetView) SimulateDraft(ctx context.Context, id string, req request.SimulateRuleSetRequest) (response.SimulateRuleSetResponse, error) {
	result, err := v.service.SimulateDraft(ctx, id, req.Event, req.DraftOverride, req.ExplainOnly, req.ExplainMaxDepth, req.ExplainCompact, req.ExplainSummary)
	if err != nil {
		return response.SimulateRuleSetResponse{}, err
	}
	return response.SimulateRuleSetResponse{Result: result}, nil
}

func (v *ruleSetView) SimulatePublished(ctx context.Context, req request.SimulateRuleSetRequest) (response.SimulateRuleSetResponse, error) {
	result, err := v.service.SimulatePublished(ctx, req.Event, req.ExplainOnly, req.ExplainMaxDepth, req.ExplainCompact, req.ExplainSummary)
	if err != nil {
		return response.SimulateRuleSetResponse{}, err
	}
	return response.SimulateRuleSetResponse{Result: result}, nil
}

func toRuleSetResponse(ruleSet bo.RuleSet) response.RuleSetResponse {
	return response.RuleSetResponse{
		ID:        ruleSet.ID,
		Name:      ruleSet.Name,
		Enabled:   ruleSet.Enabled,
		Protocol:  ruleSet.Protocol,
		Namespace: ruleSet.Namespace,
		Selector:  ruleSet.Selector,
		Rules:     ruleSet.Rules,
		Version:   ruleSet.Version,
	}
}

func toNamespaceResponse(namespace bo.Namespace) response.NamespaceResponse {
	return response.NamespaceResponse{
		ID:          namespace.ID,
		Name:        namespace.Name,
		Description: namespace.Description,
		Policies:    namespace.Policies,
	}
}

func toPublishedRuleSetResponse(snapshot bo.PublishedRuleSetSnapshot) response.PublishedRuleSetResponse {
	return response.PublishedRuleSetResponse{
		SnapshotID:  snapshot.SnapshotID,
		PublishedAt: snapshot.PublishedAt.Format(time.RFC3339),
		RuleSet:     toRuleSetResponse(snapshot.RuleSet),
		Audit:       snapshot.Audit,
	}
}

func toListPublishedRuleSetsResponse(items []bo.PublishedRuleSetSnapshot) response.ListPublishedRuleSetsResponse {
	result := response.ListPublishedRuleSetsResponse{
		Items: make([]response.PublishedRuleSetResponse, 0, len(items)),
		Total: len(items),
	}
	for _, item := range items {
		result.Items = append(result.Items, toPublishedRuleSetResponse(item))
	}
	return result
}
