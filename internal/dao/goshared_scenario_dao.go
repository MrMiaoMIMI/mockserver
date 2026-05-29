package dao

import (
	"context"
	"fmt"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	modelfmo "github.com/MrMiaoMIMI/mockserver/internal/model/fmo"
)

type gosharedScenarioTableDAO struct {
	scenarioStore  dbspi.SoftDeleteTableStore[*modeldo.Scenario]
	ruleStore      dbspi.SoftDeleteTableStore[*modeldo.ScenarioRule]
	scenarioFields modelfmo.ScenarioFields
	ruleFields     modelfmo.ScenarioRuleFields
}

func newGosharedScenarioTableDAO(manager dbspi.Manager) scenarioTableDAO {
	return &gosharedScenarioTableDAO{
		scenarioStore:  dbhelper.NewSoftDeleteTableStore(&modeldo.Scenario{}, dbhelper.WithManager(manager)),
		ruleStore:      dbhelper.NewSoftDeleteTableStore(&modeldo.ScenarioRule{}, dbhelper.WithManager(manager)),
		scenarioFields: modelfmo.NewScenarioFields(),
		ruleFields:     modelfmo.NewScenarioRuleFields(),
	}
}

func (d *gosharedScenarioTableDAO) CreateScenario(ctx context.Context, scenario modeldo.Scenario) (modeldo.Scenario, error) {
	if err := d.scenarioStore.Create(ctx, &scenario); err != nil {
		if isDuplicateEntryError(err) {
			return modeldo.Scenario{}, ErrConflict
		}
		return modeldo.Scenario{}, fmt.Errorf("insert scenario %s: %w", scenario.ScenarioCode, err)
	}
	return scenario, nil
}

func (d *gosharedScenarioTableDAO) GetScenario(ctx context.Context, id string) (modeldo.Scenario, bool, error) {
	query := dbhelper.Q(d.scenarioFields.ScenarioCode.Eq(&id))
	exists, scenario, err := d.scenarioStore.ExistsNotDeleted(ctx, query)
	if err != nil {
		return modeldo.Scenario{}, false, fmt.Errorf("get scenario %s: %w", id, err)
	}
	if !exists || scenario == nil {
		return modeldo.Scenario{}, false, nil
	}
	return *scenario, true, nil
}

func (d *gosharedScenarioTableDAO) ListScenarios(ctx context.Context, query bo.ScenarioQuery) ([]modeldo.Scenario, error) {
	conditions := []dbspi.Condition{}
	if query.Status != "" {
		conditions = append(conditions, d.scenarioFields.Status.Eq(&query.Status))
	}
	if !query.IncludeExpired && query.Now > 0 {
		conditions = append(conditions, d.scenarioFields.ExpireTime.Gt(&query.Now))
	}
	pagination := dbhelper.NewPagination().
		AppendOrder(dbhelper.Asc(d.scenarioFields.ExpireTime)).
		AppendOrder(dbhelper.Desc(d.scenarioFields.ID))
	if query.Limit > 0 {
		pagination = pagination.WithLimit(&query.Limit)
	}
	if query.Offset > 0 {
		pagination = pagination.WithOffset(&query.Offset)
	}
	items, err := d.scenarioStore.FindNotDeleted(ctx, dbhelper.Q(conditions...), pagination)
	if err != nil {
		return nil, fmt.Errorf("list scenarios: %w", err)
	}
	result := make([]modeldo.Scenario, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (d *gosharedScenarioTableDAO) UpdateScenario(ctx context.Context, scenario modeldo.Scenario, expectedVersion int) error {
	updater := dbhelper.NewUpdater().
		Set(d.scenarioFields.ScenarioName, scenario.ScenarioName).
		Set(d.scenarioFields.Status, scenario.Status).
		Set(d.scenarioFields.ExpireTime, scenario.ExpireTime).
		Set(d.scenarioFields.Version, scenario.Version).
		Set(d.scenarioFields.ScenarioJSON, scenario.ScenarioJSON)
	query := dbhelper.Q(
		d.scenarioFields.ScenarioCode.Eq(&scenario.ScenarioCode),
		d.scenarioFields.Version.Eq(&expectedVersion),
	)
	if err := d.scenarioStore.UpdateByQuery(ctx, query, updater); err != nil {
		if isDuplicateEntryError(err) {
			return ErrConflict
		}
		return fmt.Errorf("update scenario %s: %w", scenario.ScenarioCode, err)
	}
	updated, ok, err := d.GetScenario(ctx, scenario.ScenarioCode)
	if err != nil {
		return err
	}
	if !ok || updated.Version != scenario.Version {
		return ErrConflict
	}
	return nil
}

func (d *gosharedScenarioTableDAO) SoftDeleteScenario(ctx context.Context, id string) error {
	query := dbhelper.Q(d.scenarioFields.ScenarioCode.Eq(&id))
	if err := d.scenarioStore.SoftDeleteByQuery(ctx, query); err != nil {
		return fmt.Errorf("delete scenario %s: %w", id, err)
	}
	return nil
}

func (d *gosharedScenarioTableDAO) UpsertScenarioRule(ctx context.Context, rule modeldo.ScenarioRule, expectedVersion int) error {
	if expectedVersion <= 0 {
		if err := d.ruleStore.Create(ctx, &rule); err != nil {
			if isDuplicateEntryError(err) {
				return ErrConflict
			}
			return fmt.Errorf("insert scenario rule %s/%s: %w", rule.ScenarioCode, rule.RuleCode, err)
		}
		return nil
	}
	updater := dbhelper.NewUpdater().
		Set(d.ruleFields.RuleName, rule.RuleName).
		Set(d.ruleFields.ProtocolName, rule.ProtocolName).
		Set(d.ruleFields.NamespaceName, rule.NamespaceName).
		Set(d.ruleFields.Enabled, rule.Enabled).
		Set(d.ruleFields.Priority, rule.Priority).
		Set(d.ruleFields.ExpireTime, rule.ExpireTime).
		Set(d.ruleFields.Version, rule.Version).
		Set(d.ruleFields.RuleJSON, rule.RuleJSON)
	query := dbhelper.Q(
		d.ruleFields.ScenarioCode.Eq(&rule.ScenarioCode),
		d.ruleFields.RuleCode.Eq(&rule.RuleCode),
		d.ruleFields.Version.Eq(&expectedVersion),
	)
	if err := d.ruleStore.UpdateByQuery(ctx, query, updater); err != nil {
		if isDuplicateEntryError(err) {
			return ErrConflict
		}
		return fmt.Errorf("update scenario rule %s/%s: %w", rule.ScenarioCode, rule.RuleCode, err)
	}
	updated, ok, err := d.GetScenarioRule(ctx, rule.ScenarioCode, rule.RuleCode)
	if err != nil {
		return err
	}
	if !ok || updated.Version != rule.Version {
		return ErrConflict
	}
	return nil
}

func (d *gosharedScenarioTableDAO) GetScenarioRule(ctx context.Context, scenarioID, ruleID string) (modeldo.ScenarioRule, bool, error) {
	query := dbhelper.Q(
		d.ruleFields.ScenarioCode.Eq(&scenarioID),
		d.ruleFields.RuleCode.Eq(&ruleID),
	)
	exists, rule, err := d.ruleStore.ExistsNotDeleted(ctx, query)
	if err != nil {
		return modeldo.ScenarioRule{}, false, fmt.Errorf("get scenario rule %s/%s: %w", scenarioID, ruleID, err)
	}
	if !exists || rule == nil {
		return modeldo.ScenarioRule{}, false, nil
	}
	return *rule, true, nil
}

func (d *gosharedScenarioTableDAO) ListScenarioRules(ctx context.Context, query bo.ScenarioRuleQuery) ([]modeldo.ScenarioRule, error) {
	conditions := []dbspi.Condition{}
	if query.ScenarioID != "" {
		conditions = append(conditions, d.ruleFields.ScenarioCode.Eq(&query.ScenarioID))
	}
	if query.Protocol != "" {
		conditions = append(conditions, d.ruleFields.ProtocolName.Eq(&query.Protocol))
	}
	if query.Namespace != "" {
		conditions = append(conditions, d.ruleFields.NamespaceName.Eq(&query.Namespace))
	}
	if query.EnabledOnly {
		enabled := true
		conditions = append(conditions, d.ruleFields.Enabled.Eq(&enabled))
	}
	if query.Now > 0 {
		conditions = append(conditions, d.ruleFields.ExpireTime.Gt(&query.Now))
	}
	pagination := dbhelper.NewPagination().
		AppendOrder(dbhelper.Desc(d.ruleFields.Priority))
	items, err := d.ruleStore.FindNotDeleted(ctx, dbhelper.Q(conditions...), pagination)
	if err != nil {
		return nil, fmt.Errorf("list scenario rules: %w", err)
	}
	result := make([]modeldo.ScenarioRule, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (d *gosharedScenarioTableDAO) SoftDeleteScenarioRule(ctx context.Context, scenarioID, ruleID string) error {
	query := dbhelper.Q(
		d.ruleFields.ScenarioCode.Eq(&scenarioID),
		d.ruleFields.RuleCode.Eq(&ruleID),
	)
	if err := d.ruleStore.SoftDeleteByQuery(ctx, query); err != nil {
		return fmt.Errorf("delete scenario rule %s/%s: %w", scenarioID, ruleID, err)
	}
	return nil
}

func (d *gosharedScenarioTableDAO) SoftDeleteScenarioRules(ctx context.Context, scenarioID string) error {
	query := dbhelper.Q(d.ruleFields.ScenarioCode.Eq(&scenarioID))
	if err := d.ruleStore.SoftDeleteByQuery(ctx, query); err != nil {
		return fmt.Errorf("delete scenario rules %s: %w", scenarioID, err)
	}
	return nil
}
