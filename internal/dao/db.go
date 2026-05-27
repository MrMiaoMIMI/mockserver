package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/config"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

type dbImpl struct {
	manager             dbspi.Manager
	ruleSetRepository   RuleSetRepository
	namespaceRepository NamespaceRepository
	trafficRepository   TrafficRepository
	scenarioRepository  ScenarioRepository
}

func NewDB(cfg config.Config) (DB, error) {
	manager, err := newManagerFromConfig(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("create db manager: %w", err)
	}

	ruleSetTableDAO := newGosharedRuleSetTableDAO(manager)
	namespaceTableDAO := newGosharedNamespaceTableDAO(manager)
	trafficTableDAO := newGosharedTrafficTableDAO(manager)
	scenarioTableDAO := newGosharedScenarioTableDAO(manager)
	return &dbImpl{
		manager:             manager,
		ruleSetRepository:   newRuleSetRepository(ruleSetTableDAO),
		namespaceRepository: newNamespaceRepository(namespaceTableDAO),
		trafficRepository:   newTrafficRepository(trafficTableDAO, namespaceTableDAO, ruleSetTableDAO),
		scenarioRepository:  newScenarioRepository(scenarioTableDAO),
	}, nil
}

func newManagerFromConfig(dbConfig dbspi.DatabaseConfig) (dbspi.Manager, error) {
	return dbhelper.NewManager(dbConfig, dbhelper.WithCommonFieldAutoFill(true), dbhelper.WithCommonFieldTimeProvider(func(context.Context) uint64 {
		return uint64(time.Now().UnixMilli())
	}), dbhelper.WithCommonFieldOperatorProvider(func(ctx context.Context) (string, bool) {
		if operator, ok := dbspi.OperatorFromContext(ctx); ok {
			operator = strings.TrimSpace(operator)
			if operator != "" {
				return operator, true
			}
		}
		return modeldo.SystemOperator, true
	}))
}

func (d *dbImpl) GetRuleSetRepository() RuleSetRepository {
	return d.ruleSetRepository
}

func (d *dbImpl) GetNamespaceRepository() NamespaceRepository {
	return d.namespaceRepository
}

func (d *dbImpl) GetTrafficRepository() TrafficRepository {
	return d.trafficRepository
}

func (d *dbImpl) GetScenarioRepository() ScenarioRepository {
	return d.scenarioRepository
}

func (d *dbImpl) GetManager() dbspi.Manager {
	return d.manager
}
