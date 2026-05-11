package service

import (
	"context"
	"fmt"

	"mockserver/internal/dao"
	"mockserver/internal/engine"
	"mockserver/internal/model/bo"
	"mockserver/internal/model/eo"
)

type runtimeService struct {
	ruleSets            dao.RuleSetRepository
	namespaceRepository dao.NamespaceRepository
	namespaceService    NamespaceService
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
		case eo.NamespaceFallbackTypeForward:
			if action.Forward == nil {
				return bo.RuntimeDecision{}, fmt.Errorf("published forward decision requires forward fallback payload")
			}
			return bo.RuntimeDecision{
				Kind:     eo.DecisionKindForward,
				Matched:  false,
				Fallback: true,
				Trace:    result.Trace,
				Forward: &bo.ForwardDecision{
					TimeoutMS: effectiveForwardTimeoutMS(*action.Forward),
				},
				Meta: bo.DecisionMeta{
					TraceID: event.Meta.TraceID,
				},
			}, nil
		case eo.NamespaceFallbackTypeResponse:
			if action.Response == nil {
				return bo.RuntimeDecision{}, fmt.Errorf("published response decision requires response fallback payload")
			}
			return bo.RuntimeDecision{
				Kind:     eo.DecisionKindResponse,
				Matched:  false,
				Fallback: true,
				Trace:    result.Trace,
				Response: &bo.ActionExecution{
					Status:  action.Response.Status,
					Headers: cloneHeaders(action.Response.Headers),
					Body:    action.Response.Body,
				},
				Meta: bo.DecisionMeta{
					TraceID: event.Meta.TraceID,
				},
			}, nil
		default:
			return bo.RuntimeDecision{}, fmt.Errorf("unsupported namespace fallback action type %q", action.Type)
		}
	}
	return bo.RuntimeDecision{
		Kind:    eo.DecisionKindResponse,
		Matched: true,
		Trace:   result.Trace,
		Response: &bo.ActionExecution{
			Status:  result.Response.Status,
			Headers: cloneHeaders(result.Response.Headers),
			Body:    result.Response.Body,
		},
		Meta: bo.DecisionMeta{
			TraceID: event.Meta.TraceID,
		},
	}, nil
}

func (s *runtimeService) resolveNamespaceFallback(ctx context.Context, event bo.Event, result bo.SimulationResult) (string, bo.NamespaceFallbackAction, error) {
	reason := eo.FallbackReasonRulesetMiss
	action := bo.DefaultNamespace(event.Namespace).RulesetMissAction
	if result.Explain.RuleSetID != "" {
		reason = eo.FallbackReasonRuleMiss
		action = bo.DefaultNamespace(event.Namespace).RuleMissAction
	}
	namespaceID := normalizeNamespaceID(event.Namespace)
	namespace, ok, err := s.namespaceRepository.GetNamespace(ctx, namespaceID)
	if err != nil {
		return "", bo.NamespaceFallbackAction{}, err
	}
	if !ok && namespaceID == "default" {
		namespace, err = s.namespaceService.EnsureDefaultNamespace(ctx)
		if err != nil {
			return "", bo.NamespaceFallbackAction{}, err
		}
		ok = true
	}
	if ok {
		if reason == eo.FallbackReasonRulesetMiss {
			action = namespace.RulesetMissAction
		} else {
			action = namespace.RuleMissAction
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
	snapshots, err := s.ruleSets.ListPublished(ctx)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	compiledRuleSets, err := compilePublishedRuleSets(snapshots)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	return engine.Match(compiledRuleSets, event)
}
