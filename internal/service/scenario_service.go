package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"errors"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/engine"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

const (
	scenarioIDPrefix          = "scn_"
	minScenarioIDLength       = 8
	maxScenarioIDLength       = 128
	defaultScenarioTTLSeconds = 3600
	millisecondsPerSecond     = 1000
)

type scenarioService struct {
	scenarios dao.ScenarioRepository
}

func NewScenarioService(scenarioRepository dao.ScenarioRepository) ScenarioService {
	return &scenarioService{scenarios: scenarioRepository}
}

func (s *scenarioService) CreateScenario(ctx context.Context, scenario bo.Scenario) (bo.Scenario, error) {
	scenario.ID = strings.TrimSpace(scenario.ID)
	if scenario.ID == "" {
		scenario.ID = newScenarioID()
	}
	if err := validateScenarioID(scenario.ID); err != nil {
		return bo.Scenario{}, err
	}
	scenario.Name = strings.TrimSpace(scenario.Name)
	scenario.Description = strings.TrimSpace(scenario.Description)
	if scenario.Name == "" {
		scenario.Name = scenario.ID
	}
	scenario.Status = bo.ScenarioStatusActive
	if scenario.ExpireTime == 0 {
		scenario.ExpireTime = expireTimeFromTTLSeconds(defaultScenarioTTLSeconds)
	}
	scenario.Version = 1
	created, err := s.scenarios.CreateScenario(ctx, scenario)
	if errors.Is(err, dao.ErrConflict) {
		return bo.Scenario{}, conflictErrorf("scenario %s already exists", scenario.ID)
	}
	if err != nil {
		return bo.Scenario{}, err
	}
	return created, nil
}

func (s *scenarioService) GetScenario(ctx context.Context, id string) (bo.Scenario, error) {
	id = strings.TrimSpace(id)
	if err := validateScenarioID(id); err != nil {
		return bo.Scenario{}, err
	}
	scenario, ok, err := s.scenarios.GetScenario(ctx, id)
	if err != nil {
		return bo.Scenario{}, err
	}
	if !ok {
		return bo.Scenario{}, notFoundErrorf("scenario %s not found", id)
	}
	return scenario, nil
}

func (s *scenarioService) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]bo.Scenario, error) {
	query.Status = strings.TrimSpace(query.Status)
	if query.Status != "" && query.Status != bo.ScenarioStatusActive {
		return nil, validationErrorf("unsupported scenario status %q", query.Status)
	}
	if query.Now == 0 {
		query.Now = nowUnixMilli()
	}
	if query.Limit < 0 {
		query.Limit = 0
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	return s.scenarios.ListScenarios(ctx, query)
}

func (s *scenarioService) UpdateScenario(ctx context.Context, id string, update bo.ScenarioUpdate) (bo.Scenario, error) {
	id = strings.TrimSpace(id)
	if err := validateScenarioID(id); err != nil {
		return bo.Scenario{}, err
	}
	current, err := s.GetScenario(ctx, id)
	if err != nil {
		return bo.Scenario{}, err
	}
	if update.Name != nil {
		current.Name = strings.TrimSpace(*update.Name)
		if current.Name == "" {
			current.Name = current.ID
		}
	}
	if update.Description != nil {
		current.Description = strings.TrimSpace(*update.Description)
	}
	if update.ExpireTime != nil {
		if *update.ExpireTime <= nowUnixMilli() {
			return bo.Scenario{}, validationErrorf("scenario expire time must be in the future")
		}
		current.ExpireTime = *update.ExpireTime
	}
	expectedVersion := current.Version
	current.Version++
	updated, err := s.scenarios.UpdateScenario(ctx, current, expectedVersion)
	if errors.Is(err, dao.ErrConflict) {
		return bo.Scenario{}, conflictErrorf("scenario %s already changed", id)
	}
	if errors.Is(err, dao.ErrNotFound) {
		return bo.Scenario{}, notFoundErrorf("scenario %s not found", id)
	}
	if err != nil {
		return bo.Scenario{}, err
	}
	if update.ExpireTime != nil {
		if err := s.extendScenarioRuleTTL(ctx, updated); err != nil {
			return bo.Scenario{}, err
		}
	}
	return updated, nil
}

func (s *scenarioService) DeleteScenario(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if err := validateScenarioID(id); err != nil {
		return err
	}
	return s.scenarios.DeleteScenario(ctx, id)
}

func (s *scenarioService) UpsertScenarioRule(ctx context.Context, rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	rule, err := normalizeScenarioRule(rule)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	scenario, err := s.activeScenario(ctx, rule.ScenarioID)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	if rule.ExpireTime == 0 {
		rule.ExpireTime = scenario.ExpireTime
	}
	if err := validateScenarioRuleWithEngine(rule); err != nil {
		return bo.ScenarioRule{}, err
	}
	saved, err := s.scenarios.UpsertScenarioRule(ctx, rule)
	if errors.Is(err, dao.ErrNotFound) {
		return bo.ScenarioRule{}, notFoundErrorf("scenario %s not found", rule.ScenarioID)
	}
	if errors.Is(err, dao.ErrConflict) {
		return bo.ScenarioRule{}, conflictErrorf("scenario rule %s already changed", rule.RuleID)
	}
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	return saved, nil
}

func (s *scenarioService) UpsertHTTPQuickRule(ctx context.Context, scenarioID string, quick bo.HTTPQuickRule) (bo.ScenarioRule, error) {
	rule, err := httpQuickRuleToScenarioRule(scenarioID, quick)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	return s.UpsertScenarioRule(ctx, rule)
}

func (s *scenarioService) ListScenarioRules(ctx context.Context, scenarioID string) ([]bo.ScenarioRule, error) {
	scenarioID = strings.TrimSpace(scenarioID)
	if err := validateScenarioID(scenarioID); err != nil {
		return nil, err
	}
	if _, err := s.GetScenario(ctx, scenarioID); err != nil {
		return nil, err
	}
	return s.scenarios.ListScenarioRules(ctx, bo.ScenarioRuleQuery{
		ScenarioID: scenarioID,
	})
}

func (s *scenarioService) DeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	scenarioID = strings.TrimSpace(scenarioID)
	ruleID = normalizeRuleID(ruleID)
	if err := validateScenarioID(scenarioID); err != nil {
		return err
	}
	if !isValidBusinessCode(ruleID, maxRuleCodeLength) {
		return validationErrorf("rule id can only contain letters, numbers, underscores and hyphens, and must be at most %d characters", maxRuleCodeLength)
	}
	if _, err := s.activeScenario(ctx, scenarioID); err != nil {
		return err
	}
	return s.scenarios.DeleteScenarioRule(ctx, scenarioID, ruleID)
}

func (s *scenarioService) extendScenarioRuleTTL(ctx context.Context, scenario bo.Scenario) error {
	rules, err := s.scenarios.ListScenarioRules(ctx, bo.ScenarioRuleQuery{ScenarioID: scenario.ID})
	if err != nil {
		return err
	}
	for _, rule := range rules {
		rule.ExpireTime = scenario.ExpireTime
		if _, err := s.scenarios.UpsertScenarioRule(ctx, rule); err != nil {
			return err
		}
	}
	return nil
}

func (s *scenarioService) ListActiveScenarioRules(ctx context.Context, scenarioID, protocol, namespace string) ([]bo.ScenarioRule, error) {
	_ = namespace
	scenarioID = strings.TrimSpace(scenarioID)
	if scenarioID == "" {
		return nil, nil
	}
	if err := validateScenarioID(scenarioID); err != nil {
		return nil, err
	}
	if _, err := s.activeScenario(ctx, scenarioID); err != nil {
		return nil, err
	}
	return s.scenarios.ListScenarioRules(ctx, bo.ScenarioRuleQuery{
		ScenarioID:  scenarioID,
		Protocol:    strings.ToLower(strings.TrimSpace(protocol)),
		Now:         nowUnixMilli(),
		EnabledOnly: true,
	})
}

func (s *scenarioService) SimulateScenario(ctx context.Context, scenarioID string, event bo.Event, explainOnly bool, explainMaxDepth int, explainCompact bool, explainSummary bool) (bo.SimulationResult, error) {
	event.Meta.ScenarioID = scenarioID
	rules, err := s.ListActiveScenarioRules(ctx, scenarioID, event.Protocol, event.Namespace)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	return matchScenarioRules(rules, event, engine.MatchOptions{
		ExplainOnly:     explainOnly,
		ExplainMaxDepth: explainMaxDepth,
		ExplainCompact:  explainCompact,
		ExplainSummary:  explainSummary,
	})
}

func (s *scenarioService) activeScenario(ctx context.Context, scenarioID string) (bo.Scenario, error) {
	scenario, ok, err := s.scenarios.GetScenario(ctx, scenarioID)
	if err != nil {
		return bo.Scenario{}, err
	}
	if !ok {
		return bo.Scenario{}, notFoundErrorf("scenario %s not found", scenarioID)
	}
	if scenario.Status != bo.ScenarioStatusActive {
		return bo.Scenario{}, validationErrorf("scenario %s is not active", scenarioID)
	}
	if scenario.ExpireTime <= nowUnixMilli() {
		return bo.Scenario{}, validationErrorf("scenario %s is expired", scenarioID)
	}
	return scenario, nil
}

func nowUnixMilli() uint64 {
	return uint64(time.Now().UnixMilli())
}

func expireTimeFromTTLSeconds(ttlSeconds uint64) uint64 {
	return nowUnixMilli() + ttlSeconds*millisecondsPerSecond
}

func validateScenarioID(id string) error {
	if len(id) < minScenarioIDLength || len(id) > maxScenarioIDLength {
		return validationErrorf("scenario id length must be between %d and %d characters", minScenarioIDLength, maxScenarioIDLength)
	}
	if !strings.HasPrefix(id, scenarioIDPrefix) {
		return validationErrorf("scenario id must start with %q", scenarioIDPrefix)
	}
	for _, item := range id {
		if item >= 'a' && item <= 'z' || item >= 'A' && item <= 'Z' || item >= '0' && item <= '9' || item == '_' || item == '-' {
			continue
		}
		return validationErrorf("scenario id may only contain letters, numbers, underscores, and hyphens")
	}
	return nil
}

func newScenarioID() string {
	var data [6]byte
	if _, err := rand.Read(data[:]); err == nil {
		return scenarioIDPrefix + hex.EncodeToString(data[:])
	}
	return scenarioIDPrefix + strings.ToLower(time.Now().UTC().Format("150405000000"))
}

func normalizeScenarioRule(rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	rule.ScenarioID = strings.TrimSpace(rule.ScenarioID)
	if err := validateScenarioID(rule.ScenarioID); err != nil {
		return bo.ScenarioRule{}, err
	}
	rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
	rule.Namespace = ""
	rule.RuleID = normalizeRuleID(firstNonBlank(rule.RuleID, rule.Rule.ID))
	rule.Name = strings.TrimSpace(firstNonBlank(rule.Name, rule.Rule.Name, rule.RuleID))
	if rule.Protocol == "" {
		return bo.ScenarioRule{}, validationErrorf("protocol is required")
	}
	if !isValidBusinessCode(rule.RuleID, maxRuleCodeLength) {
		return bo.ScenarioRule{}, validationErrorf("rule id can only contain letters, numbers, underscores and hyphens, and must be at most %d characters", maxRuleCodeLength)
	}
	rule.Rule.ID = rule.RuleID
	rule.Rule.Name = rule.Name
	rule.Rule.Enabled = rule.Enabled
	rule.Rule.Priority = rule.Priority
	if rule.Rule.Action.Type == "" {
		rule.Rule.Action.Type = eo.ActionTypeRespond
	}
	return rule, nil
}

func validateScenarioRuleWithEngine(rule bo.ScenarioRule) error {
	ruleSet := scenarioRuleSet(rule.Protocol, "default", []bo.Rule{rule.Rule})
	validation := engine.ValidateRuleSet(ruleSet)
	if !validation.Valid {
		return validationErrorf("scenario rule validation failed: %+v", validation.Issues)
	}
	return nil
}

func matchScenarioRules(rules []bo.ScenarioRule, event bo.Event, options engine.MatchOptions) (bo.SimulationResult, error) {
	if len(rules) == 0 {
		return bo.SimulationResult{}, nil
	}
	protocol := strings.ToLower(strings.TrimSpace(event.Protocol))
	namespace := normalizeNamespaceID(event.Namespace)
	boRules := make([]bo.Rule, 0, len(rules))
	for _, rule := range rules {
		if rule.Protocol != protocol {
			continue
		}
		boRules = append(boRules, rule.Rule)
	}
	sort.SliceStable(boRules, func(i, j int) bool {
		if boRules[i].Priority != boRules[j].Priority {
			return boRules[i].Priority > boRules[j].Priority
		}
		return boRules[i].ID < boRules[j].ID
	})
	if len(boRules) == 0 {
		return bo.SimulationResult{}, nil
	}
	compiled, err := engine.CompileRuleSet(scenarioRuleSet(protocol, namespace, boRules))
	if err != nil {
		return bo.SimulationResult{}, err
	}
	result, err := engine.MatchWithOptions([]engine.CompiledRuleSet{compiled}, event, options)
	if err != nil {
		return bo.SimulationResult{}, err
	}
	result.Explain.RuleSetID = event.Meta.ScenarioID
	result.Explain.WinnerRuleSetID = event.Meta.ScenarioID
	if result.Matched {
		result.Trace.RulesetID = event.Meta.ScenarioID
	}
	return result, nil
}

func scenarioRuleSet(protocol, namespace string, rules []bo.Rule) bo.RuleSet {
	return bo.RuleSet{
		ID:        "scenario-overlay",
		Name:      "Scenario Overlay",
		Enabled:   true,
		Protocol:  protocol,
		Namespace: namespace,
		Selector:  scenarioSelector(protocol),
		Rules:     rules,
		Version:   1,
	}
}

func scenarioSelector(protocol string) bo.Selector {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case eo.ProtocolCache:
		return bo.Selector{All: []bo.Condition{{Field: "request.operation", Op: eo.OperatorExists}}}
	case eo.ProtocolSPEX:
		return bo.Selector{All: []bo.Condition{{Field: "request.cmd", Op: eo.OperatorPrefix, Value: ""}}}
	default:
		return bo.Selector{All: []bo.Condition{{Field: "request.path", Op: eo.OperatorPrefix, Value: "/"}}}
	}
}

func httpQuickRuleToScenarioRule(scenarioID string, quick bo.HTTPQuickRule) (bo.ScenarioRule, error) {
	_ = quick.Namespace
	ruleID := normalizeRuleID(quick.RuleID)
	if ruleID == "" {
		ruleID = "http-" + randomIDToken()
	}
	name := strings.TrimSpace(quick.Name)
	if name == "" {
		name = ruleID
	}
	enabled := true
	if quick.Enabled != nil {
		enabled = *quick.Enabled
	}
	conditions := make([]bo.Condition, 0)
	appendStringCondition := func(field, op, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		conditions = append(conditions, bo.Condition{Field: field, Op: op, Value: value})
	}
	appendStringCondition("request.method", eo.OperatorEQ, strings.ToUpper(quick.Match.Method))
	appendStringCondition("request.host", eo.OperatorEQ, strings.ToLower(quick.Match.Host))
	if path := strings.TrimSpace(quick.Match.Path); path != "" {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		appendStringCondition("request.path", eo.OperatorEQ, path)
	}
	appendMapConditions(&conditions, "request.query", quick.Match.Query, false)
	appendMapConditions(&conditions, "request.headers", quick.Match.Headers, true)
	appendMapConditions(&conditions, "request.body", quick.Match.Body, false)
	if len(conditions) == 0 {
		return bo.ScenarioRule{}, validationErrorf("quick rule match must contain at least one condition")
	}
	status := quick.Respond.Status
	if status == 0 {
		status = 200
	}
	action := bo.Action{
		Type:     eo.ActionTypeRespond,
		Renderer: eo.ActionRendererStatic,
		Response: &bo.ProtocolResponse{
			Protocol: eo.ProtocolHTTP,
			Payload: map[string]any{
				"status":  status,
				"headers": quick.Respond.Headers,
				"body":    quick.Respond.Body,
			},
		},
	}
	return bo.ScenarioRule{
		ScenarioID: scenarioID,
		RuleID:     ruleID,
		Name:       name,
		Protocol:   eo.ProtocolHTTP,
		Enabled:    enabled,
		Priority:   quick.Priority,
		Rule: bo.Rule{
			ID:       ruleID,
			Name:     name,
			Enabled:  enabled,
			Priority: quick.Priority,
			When:     bo.Condition{All: conditions},
			Action:   action,
		},
	}, nil
}

func appendMapConditions(conditions *[]bo.Condition, root string, values map[string]any, lowerKey bool) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := values[key]
		fieldKey := strings.TrimSpace(key)
		if fieldKey == "" {
			continue
		}
		if lowerKey {
			fieldKey = strings.ToLower(fieldKey)
		}
		field := root + "." + fieldKey
		if root == "request.query" || root == "request.headers" {
			field += "[*]"
			appendCollectionConditions(conditions, field, value)
			continue
		}
		*conditions = append(*conditions, bo.Condition{Field: field, Op: eo.OperatorEQ, Value: value})
	}
}

func appendCollectionConditions(conditions *[]bo.Condition, field string, value any) {
	switch typed := value.(type) {
	case []string:
		for _, item := range typed {
			if strings.TrimSpace(item) != "" {
				*conditions = append(*conditions, bo.Condition{Field: field, Op: eo.OperatorContains, Value: item})
			}
		}
	case []any:
		for _, item := range typed {
			*conditions = append(*conditions, bo.Condition{Field: field, Op: eo.OperatorContains, Value: item})
		}
	default:
		*conditions = append(*conditions, bo.Condition{Field: field, Op: eo.OperatorContains, Value: value})
	}
}
