package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/engine"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

type runtimeService struct {
	ruleSets            dao.RuleSetRepository
	namespaceRepository dao.NamespaceRepository
	namespaceService    NamespaceService
	cache               publishedRuntimeCache
}

type publishedRuntimeCache struct {
	mu        sync.RWMutex
	loaded    bool
	revision  uint64
	snapshots []bo.PublishedRuleSetSnapshot
	compiled  []engine.CompiledRuleSet
}

func NewRuntimeService(ruleSetRepository dao.RuleSetRepository, namespaceRepository dao.NamespaceRepository, namespaceService NamespaceService) RuntimeService {
	return &runtimeService{
		ruleSets:            ruleSetRepository,
		namespaceRepository: namespaceRepository,
		namespaceService:    namespaceService,
	}
}

func (s *runtimeService) MatchPublished(ctx context.Context, event bo.Event) (bo.SimulationResult, error) {
	result, err := s.matchPublished(ctx, event)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	if result.Matched {
		return result, nil
	}

	reason, action, err := s.resolveNamespaceFallback(ctx, event, result)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	response, err := executeHTTPRuntimeNamespaceFallback(ctx, action, event)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	result.Fallback = true
	result.Trace.FallbackReason = reason
	result.Response = response
	return result, nil
}

func (s *runtimeService) DecidePublished(ctx context.Context, event bo.Event) (bo.RuntimeDecision, error) {
	result, err := s.matchPublished(ctx, event)
	if err != nil {
		return bo.RuntimeDecision{}, err
	}
	if !result.Matched {
		reason, action, err := s.resolveNamespaceFallback(ctx, event, result)
		if err != nil {
			return bo.RuntimeDecision{}, err
		}
		result.Trace.FallbackReason = reason
		switch action.Type {
		case eo.ActionTypeForward:
			if action.Forward == nil {
				return bo.RuntimeDecision{}, fmt.Errorf("published forward decision requires forward fallback payload")
			}
			return bo.RuntimeDecision{
				Kind:     eo.DecisionKindForward,
				Matched:  false,
				Fallback: true,
				Protocol: event.Protocol,
				Trace:    result.Trace,
				Forward: &bo.ForwardDecision{
					TimeoutMS: effectiveForwardTimeoutMS(*action.Forward),
				},
				Meta: bo.DecisionMeta{
					TraceID: event.Meta.TraceID,
				},
				Diagnostics: decisionDiagnosticsFromSimulation(result),
			}, nil
		case eo.ActionTypeRespond:
			response, err := namespaceStaticResponse(event.Protocol, action)
			if err != nil {
				return bo.RuntimeDecision{}, err
			}
			return bo.RuntimeDecision{
				Kind:     eo.DecisionKindResponse,
				Matched:  false,
				Fallback: true,
				Protocol: event.Protocol,
				Trace:    result.Trace,
				Response: &response,
				Meta: bo.DecisionMeta{
					TraceID: event.Meta.TraceID,
				},
				Diagnostics: decisionDiagnosticsFromSimulation(result),
			}, nil
		default:
			return bo.RuntimeDecision{}, fmt.Errorf("unsupported namespace fallback action type %q", action.Type)
		}
	}
	return bo.RuntimeDecision{
		Kind:     eo.DecisionKindResponse,
		Matched:  true,
		Protocol: event.Protocol,
		Trace:    result.Trace,
		Response: &bo.ProtocolResponse{
			Protocol: result.Response.Protocol,
			Payload:  clonePayload(result.Response.Payload),
		},
		Meta: bo.DecisionMeta{
			TraceID: event.Meta.TraceID,
		},
		Diagnostics: decisionDiagnosticsFromSimulation(result),
	}, nil
}

func decisionDiagnosticsFromSimulation(result bo.SimulationResult) *bo.DecisionDiagnostics {
	if len(result.Explain.RuleSetCandidates) == 0 && len(result.Candidates) == 0 && result.Trace.RuleID == "" {
		return nil
	}
	diagnostics := &bo.DecisionDiagnostics{
		RuleSetSelection: bo.RuleSetSelectionDiagnostics{
			WinnerRuleSetID:  firstNonBlank(result.Trace.RulesetID, result.Explain.WinnerRuleSetID),
			WinnerSnapshotID: result.Trace.SnapshotID,
			Candidates:       append([]bo.RuleSetCandidate(nil), result.Explain.RuleSetCandidates...),
		},
		RuleSelection: bo.RuleSelectionDiagnostics{
			CandidateRuleIDs: append([]string(nil), result.Candidates...),
			WinnerRuleID:     result.Trace.RuleID,
		},
	}
	if diagnostics.RuleSetSelection.WinnerSnapshotID == "" {
		for _, candidate := range diagnostics.RuleSetSelection.Candidates {
			if candidate.Selected {
				diagnostics.RuleSetSelection.WinnerSnapshotID = candidate.SnapshotID
				break
			}
		}
	}
	return diagnostics
}

func (s *runtimeService) resolveNamespaceFallback(ctx context.Context, event bo.Event, result bo.SimulationResult) (string, bo.Action, error) {
	reason := eo.FallbackReasonRulesetMiss
	policy := namespacePolicyForProtocol(bo.DefaultNamespace(event.Namespace), event.Protocol)
	action := policy.RulesetMissAction
	if result.Explain.RuleSetID != "" {
		reason = eo.FallbackReasonRuleMiss
		action = policy.RuleMissAction
	}
	namespaceID := normalizeNamespaceID(event.Namespace)
	namespace, ok, err := s.namespaceRepository.GetNamespace(ctx, namespaceID)
	if err != nil {
		return "", bo.Action{}, err
	}
	if !ok && namespaceID == "default" {
		namespace, err = s.namespaceService.EnsureDefaultNamespace(ctx)
		if err != nil {
			return "", bo.Action{}, err
		}
		ok = true
	}
	if ok {
		policy = namespacePolicyForProtocol(namespace, event.Protocol)
		if reason == eo.FallbackReasonRulesetMiss {
			action = policy.RulesetMissAction
		} else {
			action = policy.RuleMissAction
		}
	}
	return reason, action, nil
}

func effectiveForwardTimeoutMS(action bo.NamespaceForwardFallback) int {
	if action.TimeoutMS <= 0 {
		return bo.DefaultNamespaceForwardTimeoutMS
	}
	return action.TimeoutMS
}

func (s *runtimeService) matchPublished(ctx context.Context, event bo.Event) (bo.SimulationResult, error) {
	snapshots, compiledRuleSets, err := s.cachedPublishedRuleSets(ctx)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	result, err := engine.Match(compiledRuleSets, event)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	attachPublishedSnapshotTrace(&result, snapshots)
	return result, nil
}

func (s *runtimeService) cachedPublishedRuleSets(ctx context.Context) ([]bo.PublishedRuleSetSnapshot, []engine.CompiledRuleSet, error) {
	revision := s.ruleSets.PublishedRevision()
	s.cache.mu.RLock()
	if s.cache.loaded && s.cache.revision == revision {
		snapshots := s.cache.snapshots
		compiled := s.cache.compiled
		s.cache.mu.RUnlock()
		return snapshots, compiled, nil
	}
	s.cache.mu.RUnlock()

	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()
	revision = s.ruleSets.PublishedRevision()
	if s.cache.loaded && s.cache.revision == revision {
		return s.cache.snapshots, s.cache.compiled, nil
	}

	for attempt := 0; attempt < 2; attempt++ {
		before := s.ruleSets.PublishedRevision()
		snapshots, err := s.ruleSets.ListPublished(ctx)
		if err != nil {
			return nil, nil, err
		}
		compiled, err := compilePublishedRuleSets(snapshots)
		if err != nil {
			return nil, nil, err
		}
		after := s.ruleSets.PublishedRevision()
		if before != after && attempt == 0 {
			continue
		}
		s.cache.loaded = true
		s.cache.revision = after
		s.cache.snapshots = snapshots
		s.cache.compiled = compiled
		return snapshots, compiled, nil
	}
	return nil, nil, fmt.Errorf("published ruleset cache reload failed because publish revision kept changing")
}
