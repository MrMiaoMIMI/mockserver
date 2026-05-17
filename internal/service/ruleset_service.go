package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/MrMiaoMIMI/goshared/logger"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/engine"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/mockprotocol"
)

type rulesetService struct {
	ruleSets   dao.RuleSetRepository
	namespaces NamespaceService
}

const (
	maxRuleSetCodeLength   = 96
	maxNamespaceCodeLength = 64
	maxRuleCodeLength      = 64
)

func NewRuleSetService(ruleSetRepository dao.RuleSetRepository, namespaceService NamespaceService) RuleSetService {
	return &rulesetService{
		ruleSets:   ruleSetRepository,
		namespaces: namespaceService,
	}
}

func (s *rulesetService) UpsertDraft(ctx context.Context, ruleSet bo.RuleSet) (bo.RuleSet, error) {
	ruleSet.ID = normalizeRuleSetID(ruleSet.ID)
	ruleSet.Namespace = normalizeNamespaceID(ruleSet.Namespace)
	if _, err := s.namespaces.GetNamespace(ctx, ruleSet.Namespace); err != nil {
		return bo.RuleSet{}, err
	}
	if ruleSet.ID != "" && !isValidBusinessCode(ruleSet.ID, maxRuleSetCodeLength) {
		return bo.RuleSet{}, validationErrorf("ruleset id can only contain letters, numbers, underscores and hyphens, and must be at most %d characters", maxRuleSetCodeLength)
	}
	if ruleSet.ID == "" {
		id, err := s.newRuleSetID(ctx, ruleSet)
		if err != nil {
			return bo.RuleSet{}, err
		}
		ruleSet.ID = id
	}
	var err error
	ruleSet, err = normalizeAndValidateRuleSetIdentifiers(ruleSet)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if err := s.validateRuleSetNameUnique(ctx, ruleSet); err != nil {
		return bo.RuleSet{}, err
	}
	return s.ruleSets.UpsertDraft(ctx, ruleSet)
}

func (s *rulesetService) GetDraft(ctx context.Context, id string) (bo.RuleSet, error) {
	id = normalizeRuleSetID(id)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("ruleset %s not found", id)
	}
	return ruleSet, nil
}

func (s *rulesetService) ListDrafts(ctx context.Context) ([]bo.RuleSet, error) {
	return s.ruleSets.ListDrafts(ctx)
}

func (s *rulesetService) GetPublished(ctx context.Context, id string) (bo.PublishedRuleSetSnapshot, error) {
	id = normalizeRuleSetID(id)
	snapshot, ok, err := s.ruleSets.GetPublished(ctx, id)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	if !ok {
		return bo.PublishedRuleSetSnapshot{}, notFoundErrorf("published ruleset %s not found", id)
	}
	return snapshot, nil
}

func (s *rulesetService) ListPublished(ctx context.Context) ([]bo.PublishedRuleSetSnapshot, error) {
	return s.ruleSets.ListPublished(ctx)
}

func (s *rulesetService) ListPublishedSnapshots(ctx context.Context, id string) ([]bo.PublishedRuleSetSnapshot, error) {
	id = normalizeRuleSetID(id)
	if _, ok, err := s.ruleSets.GetPublished(ctx, id); err != nil {
		return nil, err
	} else if !ok {
		return nil, notFoundErrorf("published ruleset %s not found", id)
	}
	return s.ruleSets.ListPublishedSnapshots(ctx, id)
}

func (s *rulesetService) ValidateDraft(ctx context.Context, id string) (bo.ValidationResult, error) {
	id = normalizeRuleSetID(id)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.ValidationResult{}, err
	}
	if !ok {
		return bo.ValidationResult{}, notFoundErrorf("ruleset %s not found", id)
	}
	return engine.ValidateRuleSet(ruleSet), nil
}

func (s *rulesetService) Publish(ctx context.Context, id string, audit bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	id = normalizeRuleSetID(id)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	if !ok {
		logger.Warn(ctx, "PublishRuleSet draft not found", logger.String("ruleset_id", id))
		return bo.PublishedRuleSetSnapshot{}, notFoundErrorf("ruleset %s not found", id)
	}
	if _, err := engine.CompileRuleSet(ruleSet); err != nil {
		logger.Error(ctx, "PublishRuleSet compile failed", logger.String("ruleset_id", id), logger.Err(err))
		return bo.PublishedRuleSetSnapshot{}, validationErrorf("ruleset %s is not publishable: %v", id, err)
	}
	if err := s.validatePublishedSelectorConflict(ctx, ruleSet); err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	normalizedAudit := normalizeAudit(audit, "publish", "")
	snapshot, err := s.ruleSets.Publish(ctx, id, &normalizedAudit)
	if err != nil {
		logger.Error(ctx, "PublishRuleSet failed", logger.String("ruleset_id", id), logger.Err(err))
		return bo.PublishedRuleSetSnapshot{}, err
	}
	logger.Info(ctx, "PublishRuleSet completed",
		logger.String("ruleset_id", id),
		logger.String("snapshot_id", snapshot.SnapshotID),
		logger.Int("version", snapshot.RuleSet.Version),
		logger.String("operator", snapshotAuditOperator(snapshot.Audit)),
	)
	return snapshot, nil
}

func (s *rulesetService) AddDraftRule(ctx context.Context, id string, rule bo.Rule) (bo.RuleSet, error) {
	id = normalizeRuleSetID(id)
	rule.ID = normalizeRuleID(rule.ID)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("ruleset %s not found", id)
	}
	if _, ok := findRuleIndex(ruleSet.Rules, rule.ID); ok {
		return bo.RuleSet{}, conflictErrorf("rule %s already exists in ruleset %s", rule.ID, id)
	}
	ruleSet.Rules = append(ruleSet.Rules, rule)
	return s.validateAndSaveDraft(ctx, ruleSet)
}

func (s *rulesetService) UpdateDraftRule(ctx context.Context, id string, ruleID string, rule bo.Rule) (bo.RuleSet, error) {
	id = normalizeRuleSetID(id)
	ruleID = normalizeRuleID(ruleID)
	rule.ID = normalizeRuleID(rule.ID)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("ruleset %s not found", id)
	}
	index, ok := findRuleIndex(ruleSet.Rules, ruleID)
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("rule %s not found in ruleset %s", ruleID, id)
	}
	if rule.ID == "" {
		rule.ID = ruleID
	}
	if rule.ID != ruleID {
		return bo.RuleSet{}, validationErrorf("rule id in body must match path rule id %s", ruleID)
	}
	ruleSet.Rules[index] = rule
	return s.validateAndSaveDraft(ctx, ruleSet)
}

func (s *rulesetService) DeleteDraftRule(ctx context.Context, id string, ruleID string) (bo.RuleSet, error) {
	id = normalizeRuleSetID(id)
	ruleID = normalizeRuleID(ruleID)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("ruleset %s not found", id)
	}
	index, ok := findRuleIndex(ruleSet.Rules, ruleID)
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("rule %s not found in ruleset %s", ruleID, id)
	}
	ruleSet.Rules = append(ruleSet.Rules[:index], ruleSet.Rules[index+1:]...)
	return s.validateAndSaveDraft(ctx, ruleSet)
}

func (s *rulesetService) SetDraftRuleEnabled(ctx context.Context, id string, ruleID string, enabled bool) (bo.RuleSet, error) {
	id = normalizeRuleSetID(id)
	ruleID = normalizeRuleID(ruleID)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("ruleset %s not found", id)
	}
	index, ok := findRuleIndex(ruleSet.Rules, ruleID)
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("rule %s not found in ruleset %s", ruleID, id)
	}
	ruleSet.Rules[index].Enabled = enabled
	return s.validateAndSaveDraft(ctx, ruleSet)
}

func (s *rulesetService) SetDraftRulePriority(ctx context.Context, id string, ruleID string, priority int) (bo.RuleSet, error) {
	id = normalizeRuleSetID(id)
	ruleID = normalizeRuleID(ruleID)
	ruleSet, ok, err := s.ruleSets.GetDraft(ctx, id)
	if err != nil {
		return bo.RuleSet{}, err
	}
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("ruleset %s not found", id)
	}
	index, ok := findRuleIndex(ruleSet.Rules, ruleID)
	if !ok {
		return bo.RuleSet{}, notFoundErrorf("rule %s not found in ruleset %s", ruleID, id)
	}
	ruleSet.Rules[index].Priority = priority
	return s.validateAndSaveDraft(ctx, ruleSet)
}

func (s *rulesetService) RollbackPreview(ctx context.Context, id, snapshotID string, event *bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.RollbackPreviewResult, error) {
	id = normalizeRuleSetID(id)
	snapshotID = strings.TrimSpace(snapshotID)
	snapshot, ok, err := s.ruleSets.GetPublishedSnapshot(ctx, id, snapshotID)
	if err != nil {
		return bo.RollbackPreviewResult{}, err
	}
	if !ok {
		logger.Warn(ctx, "RollbackPreview snapshot not found",
			logger.String("ruleset_id", id),
			logger.String("snapshot_id", snapshotID),
		)
		return bo.RollbackPreviewResult{}, notFoundErrorf("snapshot %s for ruleset %s not found", snapshotID, id)
	}
	current, _, err := s.ruleSets.GetPublished(ctx, id)
	if err != nil {
		return bo.RollbackPreviewResult{}, err
	}
	validation := engine.ValidateRuleSet(snapshot.RuleSet)
	result := bo.RollbackPreviewResult{
		Snapshot:   snapshot,
		Valid:      validation.Valid,
		Validation: validation,
		Diff:       summarizeRollbackDiff(current, snapshot),
	}
	if !validation.Valid {
		return result, nil
	}
	if event != nil {
		compiled, err := engine.CompileRuleSet(snapshot.RuleSet)
		if err != nil {
			logger.Error(ctx, "RollbackPreview compile failed",
				logger.String("ruleset_id", id),
				logger.String("snapshot_id", snapshotID),
				logger.Err(err),
			)
			return bo.RollbackPreviewResult{}, err
		}
		simulation, err := engine.MatchWithOptions([]engine.CompiledRuleSet{compiled}, *event, engine.MatchOptions{
			ExplainOnly:     explainOnly,
			ExplainMaxDepth: explainMaxDepth,
			ExplainCompact:  explainCompact,
			ExplainSummary:  explainSummary,
		})
		if err != nil {
			logger.Error(ctx, "RollbackPreview simulate failed",
				logger.String("ruleset_id", id),
				logger.String("snapshot_id", snapshotID),
				logger.Err(err),
			)
			return bo.RollbackPreviewResult{}, err
		}
		result.Simulation = &simulation
	}
	logger.Info(ctx, "RollbackPreview completed",
		logger.String("ruleset_id", id),
		logger.String("snapshot_id", snapshotID),
		logger.Bool("valid", result.Valid),
		logger.Bool("has_event", event != nil),
		logger.Bool("matched", simulationMatched(result.Simulation)),
		logger.Bool("explain_only", explainOnly),
		logger.Bool("explain_compact", explainCompact),
		logger.Bool("explain_summary", explainSummary),
		logger.Int("explain_max_depth", explainMaxDepth),
	)
	return result, nil
}

func (s *rulesetService) Rollback(ctx context.Context, id, snapshotID string, audit bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	id = normalizeRuleSetID(id)
	snapshotID = strings.TrimSpace(snapshotID)
	snapshot, ok, err := s.ruleSets.GetPublishedSnapshot(ctx, id, snapshotID)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	if !ok {
		logger.Warn(ctx, "RollbackRuleSet snapshot not found",
			logger.String("ruleset_id", id),
			logger.String("snapshot_id", snapshotID),
		)
		return bo.PublishedRuleSetSnapshot{}, notFoundErrorf("snapshot %s for ruleset %s not found", snapshotID, id)
	}
	if _, err := engine.CompileRuleSet(snapshot.RuleSet); err != nil {
		logger.Error(ctx, "RollbackRuleSet compile failed",
			logger.String("ruleset_id", id),
			logger.String("snapshot_id", snapshotID),
			logger.Err(err),
		)
		return bo.PublishedRuleSetSnapshot{}, validationErrorf("snapshot %s for ruleset %s is not publishable: %v", snapshotID, id, err)
	}
	if err := s.validatePublishedSelectorConflict(ctx, snapshot.RuleSet); err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	normalizedAudit := normalizeAudit(audit, "rollback", snapshotID)
	rolledBack, err := s.ruleSets.Rollback(ctx, id, snapshotID, &normalizedAudit)
	if err != nil {
		logger.Error(ctx, "RollbackRuleSet failed",
			logger.String("ruleset_id", id),
			logger.String("snapshot_id", snapshotID),
			logger.Err(err),
		)
		return bo.PublishedRuleSetSnapshot{}, err
	}
	logger.Info(ctx, "RollbackRuleSet completed",
		logger.String("ruleset_id", id),
		logger.String("source_snapshot_id", snapshotID),
		logger.String("snapshot_id", rolledBack.SnapshotID),
		logger.Int("version", rolledBack.RuleSet.Version),
		logger.String("operator", snapshotAuditOperator(rolledBack.Audit)),
	)
	return rolledBack, nil
}

func (s *rulesetService) SimulateDraft(ctx context.Context, id string, event bo.Event, draftOverride *bo.RuleSet, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error) {
	id = normalizeRuleSetID(id)
	var ruleSet bo.RuleSet
	if draftOverride != nil {
		ruleSet = *draftOverride
		if ruleSet.ID == "" {
			ruleSet.ID = id
		}
	} else {
		var ok bool
		var err error
		ruleSet, ok, err = s.ruleSets.GetDraft(ctx, id)
		if err != nil {
			return bo.SimulationResult{}, err
		}
		if !ok {
			logger.Warn(ctx, "SimulateDraft ruleset not found", logger.String("ruleset_id", id))
			return bo.SimulationResult{}, notFoundErrorf("ruleset %s not found", id)
		}
	}
	compiled, err := engine.CompileRuleSet(ruleSet)
	if err != nil {
		logger.Error(ctx, "SimulateDraft compile failed", logger.String("ruleset_id", id), logger.Err(err))
		return bo.SimulationResult{}, err
	}
	result, err := engine.MatchWithOptions([]engine.CompiledRuleSet{compiled}, event, engine.MatchOptions{
		ExplainOnly:     explainOnly,
		ExplainMaxDepth: explainMaxDepth,
		ExplainCompact:  explainCompact,
		ExplainSummary:  explainSummary,
	})
	if err != nil {
		logger.Error(ctx, "SimulateDraft failed", logger.String("ruleset_id", id), logger.Err(err))
		return bo.SimulationResult{}, err
	}
	logSimulationResult(ctx, "SimulateDraft completed", result, event, explainOnly, explainMaxDepth, explainCompact, explainSummary)
	return result, nil
}

func (s *rulesetService) SimulatePublished(ctx context.Context, event bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error) {
	snapshots, err := s.ruleSets.ListPublished(ctx)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	compiledRuleSets, err := compilePublishedRuleSets(snapshots)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	result, err := engine.MatchWithOptions(compiledRuleSets, event, engine.MatchOptions{
		ExplainOnly:     explainOnly,
		ExplainMaxDepth: explainMaxDepth,
		ExplainCompact:  explainCompact,
		ExplainSummary:  explainSummary,
	})
	if err != nil {
		logger.Error(ctx, "SimulatePublished failed", logger.Err(err))
		return bo.SimulationResult{}, err
	}
	attachPublishedSnapshotTrace(&result, snapshots)
	logSimulationResult(ctx, "SimulatePublished completed", result, event, explainOnly, explainMaxDepth, explainCompact, explainSummary)
	return result, nil
}

func attachPublishedSnapshotTrace(result *bo.SimulationResult, snapshots []bo.PublishedRuleSetSnapshot) {
	if result == nil {
		return
	}
	snapshotIDs := make(map[string]string, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotIDs[snapshot.RuleSet.ID] = snapshot.SnapshotID
	}
	for i := range result.Explain.RuleSetCandidates {
		if snapshotID := snapshotIDs[result.Explain.RuleSetCandidates[i].RuleSetID]; snapshotID != "" {
			result.Explain.RuleSetCandidates[i].SnapshotID = snapshotID
		}
	}
	if result.Trace.RulesetID == "" {
		return
	}
	for _, snapshot := range snapshots {
		if snapshot.RuleSet.ID != result.Trace.RulesetID {
			continue
		}
		result.Trace.RulesetDBID = snapshot.RuleSet.DBID
		result.Trace.SnapshotDBID = snapshot.DBID
		result.Trace.SnapshotID = snapshot.SnapshotID
		return
	}
}

func compilePublishedRuleSets(snapshots []bo.PublishedRuleSetSnapshot) ([]engine.CompiledRuleSet, error) {
	items := make([]engine.CompiledRuleSet, 0, len(snapshots))
	for _, snapshot := range snapshots {
		compiled, err := engine.CompileRuleSet(snapshot.RuleSet)
		if err != nil {
			return nil, err
		}
		items = append(items, compiled)
	}
	return items, nil
}

func (s *rulesetService) validateAndSaveDraft(ctx context.Context, ruleSet bo.RuleSet) (bo.RuleSet, error) {
	var err error
	ruleSet, err = normalizeAndValidateRuleSetIdentifiers(ruleSet)
	if err != nil {
		return bo.RuleSet{}, err
	}
	validation := engine.ValidateRuleSet(ruleSet)
	if !validation.Valid {
		return bo.RuleSet{}, validationErrorf("ruleset validation failed: %+v", validation.Issues)
	}
	return s.ruleSets.UpsertDraft(ctx, ruleSet)
}

func findRuleIndex(rules []bo.Rule, ruleID string) (int, bool) {
	for i, rule := range rules {
		if rule.ID == ruleID {
			return i, true
		}
	}
	return -1, false
}

func normalizeAudit(audit bo.AuditInfo, action string, sourceSnapshotID string) bo.AuditInfo {
	audit.Action = action
	if audit.Operator == "" {
		audit.Operator = "anonymous"
	}
	if sourceSnapshotID != "" {
		audit.SourceSnapshotID = sourceSnapshotID
	}
	return audit
}

func logSimulationResult(ctx context.Context, msg string, result bo.SimulationResult, event bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) {
	logger.Info(ctx, msg,
		logger.String("namespace", event.Namespace),
		logger.String("protocol", event.Protocol),
		logger.String("method", bo.RequestString(event.Request, "method")),
		logger.String("path", bo.RequestString(event.Request, "path")),
		logger.Bool("matched", result.Matched),
		logger.String("ruleset_id", result.Trace.RulesetID),
		logger.String("rule_id", result.Trace.RuleID),
		logger.Bool("explain_only", explainOnly),
		logger.Bool("explain_compact", explainCompact),
		logger.Bool("explain_summary", explainSummary),
		logger.Int("explain_max_depth", explainMaxDepth),
	)
}

func simulationMatched(result *bo.SimulationResult) bool {
	return result != nil && result.Matched
}

func snapshotAuditOperator(audit *bo.AuditInfo) string {
	if audit == nil {
		return ""
	}
	return audit.Operator
}

func normalizeNamespace(namespace bo.Namespace) (bo.Namespace, error) {
	namespace.ID = normalizeNamespaceID(namespace.ID)
	if namespace.ID == "" {
		return bo.Namespace{}, validationErrorf("namespace id is required")
	}
	if !isValidNamespaceID(namespace.ID) {
		return bo.Namespace{}, validationErrorf("namespace id can only contain letters, numbers, underscores and hyphens")
	}
	namespace.Name = strings.TrimSpace(namespace.Name)
	namespace.Description = strings.TrimSpace(namespace.Description)
	if namespace.Name == "" {
		namespace.Name = namespace.ID
	}
	if namespace.Policies == nil {
		namespace.Policies = map[string]bo.NamespacePolicy{}
	}
	normalizedPolicies := make(map[string]bo.NamespacePolicy, len(namespace.Policies))
	for protocol, policy := range namespace.Policies {
		protocol = strings.ToLower(strings.TrimSpace(protocol))
		if protocol == "" {
			return bo.Namespace{}, validationErrorf("policy protocol is required")
		}
		normalizedPolicies[protocol] = policy
	}
	namespace.Policies = normalizedPolicies
	for _, spec := range mockprotocol.RegisteredSpecs() {
		protocol := strings.ToLower(strings.TrimSpace(spec.Name))
		policy := namespace.Policies[protocol]
		if isNamespaceActionEmpty(policy.RulesetMissAction) {
			policy.RulesetMissAction = bo.DefaultNamespaceForwardAction()
		}
		if isNamespaceActionEmpty(policy.RuleMissAction) {
			policy.RuleMissAction = bo.DefaultNamespaceForwardAction()
		}
		if err := validateNamespacePolicyAction(protocol, policy.RulesetMissAction, "policies."+protocol+".ruleset_miss_action"); err != nil {
			return bo.Namespace{}, err
		}
		if err := validateNamespacePolicyAction(protocol, policy.RuleMissAction, "policies."+protocol+".rule_miss_action"); err != nil {
			return bo.Namespace{}, err
		}
		namespace.Policies[protocol] = policy
	}
	for protocol, policy := range namespace.Policies {
		if _, ok := mockprotocol.DefaultRegistry().Get(protocol); !ok {
			return bo.Namespace{}, validationErrorf("policies.%s uses unsupported protocol", protocol)
		}
		if err := validateNamespacePolicyAction(protocol, policy.RulesetMissAction, "policies."+protocol+".ruleset_miss_action"); err != nil {
			return bo.Namespace{}, err
		}
		if err := validateNamespacePolicyAction(protocol, policy.RuleMissAction, "policies."+protocol+".rule_miss_action"); err != nil {
			return bo.Namespace{}, err
		}
	}
	return namespace, nil
}

func isNamespaceActionEmpty(action bo.Action) bool {
	return action.Type == "" && action.Renderer == "" && action.Response == nil && action.Forward == nil
}

func normalizeNamespaceID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func isValidNamespaceID(id string) bool {
	if !isValidBusinessCode(id, maxNamespaceCodeLength) {
		return false
	}
	return true
}

func normalizeRuleSetID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func normalizeRuleID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func isValidBusinessCode(id string, maxLength int) bool {
	if id == "" || len(id) > maxLength {
		return false
	}
	for _, item := range id {
		if item >= 'a' && item <= 'z' || item >= '0' && item <= '9' || item == '_' || item == '-' {
			continue
		}
		return false
	}
	return true
}

func normalizeAndValidateRuleSetIdentifiers(ruleSet bo.RuleSet) (bo.RuleSet, error) {
	ruleSet.ID = normalizeRuleSetID(ruleSet.ID)
	ruleSet.Name = strings.TrimSpace(ruleSet.Name)
	ruleSet.Protocol = strings.ToLower(strings.TrimSpace(ruleSet.Protocol))
	ruleSet.Namespace = normalizeNamespaceID(ruleSet.Namespace)
	if ruleSet.Name == "" {
		return bo.RuleSet{}, validationErrorf("ruleset name is required")
	}
	if !isValidBusinessCode(ruleSet.ID, maxRuleSetCodeLength) {
		return bo.RuleSet{}, validationErrorf("ruleset id can only contain letters, numbers, underscores and hyphens, and must be at most %d characters", maxRuleSetCodeLength)
	}
	if !isValidBusinessCode(ruleSet.Namespace, maxNamespaceCodeLength) {
		return bo.RuleSet{}, validationErrorf("namespace id can only contain letters, numbers, underscores and hyphens, and must be at most %d characters", maxNamespaceCodeLength)
	}
	for i := range ruleSet.Rules {
		ruleSet.Rules[i].ID = normalizeRuleID(ruleSet.Rules[i].ID)
		ruleSet.Rules[i].Name = strings.TrimSpace(ruleSet.Rules[i].Name)
		if !isValidBusinessCode(ruleSet.Rules[i].ID, maxRuleCodeLength) {
			return bo.RuleSet{}, validationErrorf("rule id can only contain letters, numbers, underscores and hyphens, and must be at most %d characters", maxRuleCodeLength)
		}
	}
	return ruleSet, nil
}

func (s *rulesetService) validateRuleSetNameUnique(ctx context.Context, ruleSet bo.RuleSet) error {
	items, err := s.ruleSets.ListDrafts(ctx)
	if err != nil {
		return err
	}
	nameKey := normalizedDisplayName(ruleSet.Name)
	namespaceKey := normalizeNamespaceID(ruleSet.Namespace)
	protocolKey := strings.ToLower(strings.TrimSpace(ruleSet.Protocol))
	for _, item := range items {
		if item.ID == ruleSet.ID {
			continue
		}
		if normalizeNamespaceID(item.Namespace) != namespaceKey ||
			strings.ToLower(strings.TrimSpace(item.Protocol)) != protocolKey {
			continue
		}
		if normalizedDisplayName(item.Name) == nameKey {
			return conflictErrorf(
				"ruleset name %q is already used by ruleset %s in namespace %s protocol %s",
				ruleSet.Name,
				item.ID,
				namespaceKey,
				protocolKey,
			)
		}
	}
	return nil
}

func (s *rulesetService) validatePublishedSelectorConflict(ctx context.Context, candidate bo.RuleSet) error {
	if !candidate.Enabled {
		return nil
	}
	published, err := s.ruleSets.ListPublished(ctx)
	if err != nil {
		return err
	}
	candidateKey, err := publishedSelectorConflictKey(candidate)
	if err != nil {
		return err
	}
	for _, snapshot := range published {
		existing := snapshot.RuleSet
		if existing.ID == candidate.ID || !existing.Enabled {
			continue
		}
		existingKey, err := publishedSelectorConflictKey(existing)
		if err != nil {
			return err
		}
		if existingKey == candidateKey {
			return conflictErrorf(
				"ruleset %s conflicts with published ruleset %s: same namespace, protocol, and selector",
				candidate.ID,
				existing.ID,
			)
		}
	}
	return nil
}

func normalizedDisplayName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func publishedSelectorConflictKey(ruleSet bo.RuleSet) (string, error) {
	selector := canonicalSelector(ruleSet.Selector)
	raw, err := json.Marshal(selector)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{
		strings.ToLower(strings.TrimSpace(ruleSet.Namespace)),
		strings.ToLower(strings.TrimSpace(ruleSet.Protocol)),
		string(raw),
	}, "\x00"), nil
}

func canonicalSelector(selector bo.Selector) bo.Selector {
	return bo.Selector{All: canonicalConditions(selector.All)}
}

func canonicalConditions(conditions []bo.Condition) []bo.Condition {
	result := make([]bo.Condition, 0, len(conditions))
	for _, condition := range conditions {
		result = append(result, canonicalCondition(condition))
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, _ := json.Marshal(result[i])
		right, _ := json.Marshal(result[j])
		return string(left) < string(right)
	})
	return result
}

func canonicalCondition(condition bo.Condition) bo.Condition {
	condition.Field = strings.TrimSpace(condition.Field)
	condition.Op = strings.TrimSpace(condition.Op)
	condition.Expr = strings.TrimSpace(condition.Expr)
	condition.All = canonicalConditions(condition.All)
	condition.Any = canonicalConditions(condition.Any)
	if condition.Not != nil {
		normalized := canonicalCondition(*condition.Not)
		condition.Not = &normalized
	}
	if value, ok := condition.Value.(string); ok {
		value = strings.TrimSpace(value)
		switch condition.Field {
		case "request.host", "request.original_host":
			value = strings.ToLower(value)
		case "request.method":
			value = strings.ToUpper(value)
		case "request.operation":
			value = strings.ToLower(value)
		}
		condition.Value = value
	}
	return condition
}

func validateNamespacePolicyAction(protocol string, action bo.Action, path string) error {
	switch action.Type {
	case eo.ActionTypeRespond:
		if renderer := strings.TrimSpace(action.Renderer); renderer != "" && renderer != eo.ActionRendererStatic {
			return validationErrorf("%s.renderer only supports static for namespace policies", path)
		}
		var payload map[string]any
		if action.Response != nil {
			payload = action.Response.Payload
		}
		if _, err := mockprotocol.NormalizeResponsePayload(protocol, payload); err != nil {
			return validationErrorf("%s.response.payload: %v", path, err)
		}
	case eo.ActionTypeForward:
		if action.Forward == nil {
			return validationErrorf("%s.forward is required", path)
		}
		if action.Forward.TimeoutMS < 0 || action.Forward.TimeoutMS > 30000 {
			return validationErrorf("%s.forward.timeout_ms must be between 0 and 30000", path)
		}
	default:
		return validationErrorf("%s.type must be respond or forward", path)
	}
	return nil
}

func namespacePolicyForProtocol(namespace bo.Namespace, protocol string) bo.NamespacePolicy {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	if namespace.Policies != nil {
		if policy, ok := namespace.Policies[protocol]; ok {
			return policy
		}
	}
	defaultPolicy := bo.NamespacePolicy{
		RulesetMissAction: bo.DefaultNamespaceForwardAction(),
		RuleMissAction:    bo.DefaultNamespaceForwardAction(),
	}
	return defaultPolicy
}

func namespaceStaticResponse(protocol string, action bo.Action) (bo.ProtocolResponse, error) {
	if action.Type != eo.ActionTypeRespond {
		return bo.ProtocolResponse{}, fmt.Errorf("namespace action type must be respond")
	}
	if renderer := strings.TrimSpace(action.Renderer); renderer != "" && renderer != eo.ActionRendererStatic {
		return bo.ProtocolResponse{}, fmt.Errorf("namespace response renderer %q is not supported", action.Renderer)
	}
	response := action.Response
	if response == nil {
		response = &bo.ProtocolResponse{}
	}
	if strings.TrimSpace(response.Protocol) != "" && !strings.EqualFold(response.Protocol, protocol) {
		return bo.ProtocolResponse{}, fmt.Errorf("namespace response protocol %q does not match event protocol %q", response.Protocol, protocol)
	}
	payload, err := mockprotocol.NormalizeResponsePayload(protocol, response.Payload)
	if err != nil {
		return bo.ProtocolResponse{}, err
	}
	return bo.ProtocolResponse{
		Protocol: strings.ToLower(strings.TrimSpace(protocol)),
		Payload:  payload,
	}, nil
}

func clonePayload(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		cloned := make(map[string]any, len(payload))
		for key, value := range payload {
			cloned[key] = value
		}
		return cloned
	}
	var cloned map[string]any
	if err := json.Unmarshal(raw, &cloned); err != nil {
		return nil
	}
	return cloned
}

func executeHTTPRuntimeNamespaceFallback(ctx context.Context, action bo.Action, event bo.Event) (bo.ProtocolResponse, error) {
	switch action.Type {
	case eo.ActionTypeRespond:
		return namespaceStaticResponse(event.Protocol, action)
	case eo.ActionTypeForward:
		if action.Forward == nil {
			return bo.ProtocolResponse{}, fmt.Errorf("namespace fallback forward is required")
		}
		return executeHTTPRuntimeForwardFallback(ctx, *action.Forward, event)
	default:
		return bo.ProtocolResponse{}, fmt.Errorf("unsupported namespace fallback action type %q", action.Type)
	}
}

func executeHTTPRuntimeForwardFallback(ctx context.Context, action bo.NamespaceForwardFallback, event bo.Event) (bo.ProtocolResponse, error) {
	if event.Protocol != eo.ProtocolHTTP {
		return bo.ProtocolResponse{}, fmt.Errorf("server-side forward fallback execution is only supported by HTTP runtime; use a forward decision for protocol-specific mockinject forwarding")
	}
	targetURL, forwardHost, err := buildHTTPRuntimeForwardURL(event)
	if err != nil {
		return bo.ProtocolResponse{}, err
	}
	timeout := time.Duration(action.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	req, err := http.NewRequestWithContext(ctx, bo.RequestString(event.Request, "method"), targetURL, bytes.NewBufferString(bo.RequestString(event.Request, "raw_body")))
	if err != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("build forward request: %w", err)
	}
	req.Host = forwardHost
	for key, values := range bo.RequestStringMap(event.Request, "headers") {
		if isHTTPRuntimeForwardSkippedHeader(key) {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("forward fallback request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("read forward fallback response: %w", err)
	}
	var body any = string(bodyBytes)
	if len(bodyBytes) > 0 {
		var parsed any
		if err := json.Unmarshal(bodyBytes, &parsed); err == nil {
			body = parsed
		}
	}
	return bo.ProtocolResponse{
		Protocol: eo.ProtocolHTTP,
		Payload: map[string]any{
			"status":  resp.StatusCode,
			"headers": cloneHeaders(resp.Header),
			"body":    body,
		},
	}, nil
}

func buildHTTPRuntimeForwardURL(event bo.Event) (string, string, error) {
	scheme := strings.TrimSpace(bo.RequestString(event.Request, "scheme"))
	if scheme == "" {
		scheme = "http"
	}
	if scheme != "http" && scheme != "https" {
		return "", "", fmt.Errorf("forward fallback only supports http or https scheme")
	}
	forwardHost := strings.TrimSpace(bo.RequestString(event.Request, "original_host"))
	if forwardHost == "" {
		forwardHost = strings.TrimSpace(bo.RequestString(event.Request, "host"))
	}
	if forwardHost == "" {
		return "", "", fmt.Errorf("forward fallback requires original request host")
	}
	parsed := url.URL{
		Scheme: scheme,
		Host:   forwardHost,
		Path:   "/" + strings.TrimLeft(bo.RequestString(event.Request, "path"), "/"),
	}
	query := url.Values{}
	for key, values := range bo.RequestStringMap(event.Request, "query") {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), forwardHost, nil
}

func isHTTPRuntimeForwardSkippedHeader(key string) bool {
	switch strings.ToLower(key) {
	case "connection", "content-length", "host", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func cloneHeaders(headers map[string][]string) map[string][]string {
	if len(headers) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func (s *rulesetService) newRuleSetID(ctx context.Context, ruleSet bo.RuleSet) (string, error) {
	base := slugifyRuleSetID(strings.Join([]string{ruleSet.Protocol, ruleSet.Namespace, ruleSet.Name}, "-"))
	if base == "" {
		base = "ruleset"
	}
	const suffixLength = 8
	maxBaseLength := maxRuleSetCodeLength - suffixLength - 1
	if len(base) > maxBaseLength {
		base = strings.Trim(base[:maxBaseLength], "-")
	}
	if base == "" {
		base = "ruleset"
	}
	for i := 0; i < 20; i++ {
		candidate := base + "-" + randomIDToken()
		_, draftExists, err := s.ruleSets.GetDraft(ctx, candidate)
		if err != nil {
			return "", err
		}
		if draftExists {
			continue
		}
		_, publishedExists, err := s.ruleSets.GetPublished(ctx, candidate)
		if err != nil {
			return "", err
		}
		if publishedExists {
			continue
		}
		return candidate, nil
	}
	return "", fmt.Errorf("generate unique ruleset id failed")
}

func slugifyRuleSetID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, item := range value {
		isAlphaNumber := item >= 'a' && item <= 'z' || item >= '0' && item <= '9'
		if isAlphaNumber {
			builder.WriteRune(item)
			lastDash = false
			continue
		}
		if !lastDash && builder.Len() > 0 {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func randomIDToken() string {
	var data [4]byte
	if _, err := rand.Read(data[:]); err == nil {
		return hex.EncodeToString(data[:])
	}
	return strings.ToLower(time.Now().UTC().Format("15040500"))
}
