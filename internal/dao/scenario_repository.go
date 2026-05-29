package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

type scenarioRepository struct {
	tableDAO scenarioTableDAO
}

func newScenarioRepository(tableDAO scenarioTableDAO) ScenarioRepository {
	return &scenarioRepository{tableDAO: tableDAO}
}

func (r *scenarioRepository) CreateScenario(ctx context.Context, scenario bo.Scenario) (bo.Scenario, error) {
	if scenario.Version <= 0 {
		scenario.Version = 1
	}
	record, err := encodeScenarioRecord(scenario)
	if err != nil {
		return bo.Scenario{}, err
	}
	record, err = r.tableDAO.CreateScenario(ctx, record)
	if err != nil {
		return bo.Scenario{}, err
	}
	scenario.DBID = record.Id
	return scenario, nil
}

func (r *scenarioRepository) GetScenario(ctx context.Context, id string) (bo.Scenario, bool, error) {
	record, ok, err := r.tableDAO.GetScenario(ctx, id)
	if err != nil || !ok {
		return bo.Scenario{}, false, err
	}
	scenario, err := decodeScenarioRecord(record)
	if err != nil {
		return bo.Scenario{}, false, err
	}
	return scenario, true, nil
}

func (r *scenarioRepository) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]bo.Scenario, error) {
	records, err := r.tableDAO.ListScenarios(ctx, query)
	if err != nil {
		return nil, err
	}
	items := make([]bo.Scenario, 0, len(records))
	for _, record := range records {
		item, err := decodeScenarioRecord(record)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *scenarioRepository) UpdateScenario(ctx context.Context, scenario bo.Scenario, expectedVersion int) (bo.Scenario, error) {
	record, err := encodeScenarioRecord(scenario)
	if err != nil {
		return bo.Scenario{}, err
	}
	if err := r.tableDAO.UpdateScenario(ctx, record, expectedVersion); err != nil {
		return bo.Scenario{}, err
	}
	saved, ok, err := r.GetScenario(ctx, scenario.ID)
	if err != nil {
		return bo.Scenario{}, err
	}
	if !ok {
		return bo.Scenario{}, ErrNotFound
	}
	return saved, nil
}

func (r *scenarioRepository) DeleteScenario(ctx context.Context, id string) error {
	if err := r.tableDAO.SoftDeleteScenarioRules(ctx, id); err != nil {
		return err
	}
	return r.tableDAO.SoftDeleteScenario(ctx, id)
}

func (r *scenarioRepository) UpsertScenarioRule(ctx context.Context, rule bo.ScenarioRule) (bo.ScenarioRule, error) {
	scenario, ok, err := r.GetScenario(ctx, rule.ScenarioID)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	if !ok {
		return bo.ScenarioRule{}, ErrNotFound
	}
	current, exists, err := r.tableDAO.GetScenarioRule(ctx, rule.ScenarioID, rule.RuleID)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	expectedVersion := 0
	if exists {
		expectedVersion = current.Version
		rule.Version = current.Version + 1
	} else {
		rule.Version = 1
	}
	rule.ScenarioDBID = scenario.DBID
	if rule.ExpireTime == 0 {
		rule.ExpireTime = scenario.ExpireTime
	}
	record, err := encodeScenarioRuleRecord(rule)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	if err := r.tableDAO.UpsertScenarioRule(ctx, record, expectedVersion); err != nil {
		return bo.ScenarioRule{}, err
	}
	saved, ok, err := r.tableDAO.GetScenarioRule(ctx, rule.ScenarioID, rule.RuleID)
	if err != nil {
		return bo.ScenarioRule{}, err
	}
	if ok {
		rule.DBID = saved.Id
	}
	return rule, nil
}

func (r *scenarioRepository) ListScenarioRules(ctx context.Context, query bo.ScenarioRuleQuery) ([]bo.ScenarioRule, error) {
	records, err := r.tableDAO.ListScenarioRules(ctx, query)
	if err != nil {
		return nil, err
	}
	items := make([]bo.ScenarioRule, 0, len(records))
	for _, record := range records {
		item, err := decodeScenarioRuleRecord(record)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *scenarioRepository) DeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	return r.tableDAO.SoftDeleteScenarioRule(ctx, scenarioID, ruleID)
}

func encodeScenarioRecord(scenario bo.Scenario) (modeldo.Scenario, error) {
	raw, err := json.Marshal(scenario)
	if err != nil {
		return modeldo.Scenario{}, fmt.Errorf("marshal scenario %s: %w", scenario.ID, err)
	}
	return modeldo.Scenario{
		ScenarioCode: scenario.ID,
		ScenarioName: scenario.Name,
		Status:       scenario.Status,
		ExpireTime:   scenario.ExpireTime,
		Version:      scenario.Version,
		ScenarioJSON: string(raw),
	}, nil
}

func decodeScenarioRecord(record modeldo.Scenario) (bo.Scenario, error) {
	var scenario bo.Scenario
	if record.ScenarioJSON != "" {
		if err := json.Unmarshal([]byte(record.ScenarioJSON), &scenario); err != nil {
			return bo.Scenario{}, fmt.Errorf("decode scenario %s: %w", record.ScenarioCode, err)
		}
	}
	scenario.DBID = record.Id
	if scenario.ID == "" {
		scenario.ID = record.ScenarioCode
	}
	if scenario.Name == "" {
		scenario.Name = record.ScenarioName
	}
	if scenario.Status == "" {
		scenario.Status = record.Status
	}
	if scenario.ExpireTime == 0 {
		scenario.ExpireTime = record.ExpireTime
	}
	if scenario.Version == 0 {
		scenario.Version = record.Version
	}
	return scenario, nil
}

func encodeScenarioRuleRecord(rule bo.ScenarioRule) (modeldo.ScenarioRule, error) {
	raw, err := json.Marshal(rule.Rule)
	if err != nil {
		return modeldo.ScenarioRule{}, fmt.Errorf("marshal scenario rule %s/%s: %w", rule.ScenarioID, rule.RuleID, err)
	}
	return modeldo.ScenarioRule{
		ScenarioID:    rule.ScenarioDBID,
		ScenarioCode:  rule.ScenarioID,
		RuleCode:      rule.RuleID,
		RuleName:      rule.Name,
		ProtocolName:  rule.Protocol,
		NamespaceName: rule.Namespace,
		Enabled:       rule.Enabled,
		Priority:      rule.Priority,
		ExpireTime:    rule.ExpireTime,
		Version:       rule.Version,
		RuleJSON:      string(raw),
	}, nil
}

func decodeScenarioRuleRecord(record modeldo.ScenarioRule) (bo.ScenarioRule, error) {
	var rule bo.Rule
	if record.RuleJSON != "" {
		if err := json.Unmarshal([]byte(record.RuleJSON), &rule); err != nil {
			return bo.ScenarioRule{}, fmt.Errorf("decode scenario rule %s/%s: %w", record.ScenarioCode, record.RuleCode, err)
		}
	}
	return bo.ScenarioRule{
		DBID:         record.Id,
		ScenarioID:   record.ScenarioCode,
		ScenarioDBID: record.ScenarioID,
		RuleID:       record.RuleCode,
		Name:         record.RuleName,
		Protocol:     record.ProtocolName,
		Namespace:    record.NamespaceName,
		Enabled:      record.Enabled,
		Priority:     record.Priority,
		ExpireTime:   record.ExpireTime,
		Rule:         rule,
		Version:      record.Version,
	}, nil
}
