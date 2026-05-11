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
}

func NewDB(ctx context.Context, cfg config.Config) (DB, error) {
	if cfg.DBDriver != "" && cfg.DBDriver != "mysql" {
		return nil, fmt.Errorf("unsupported db.driver %q; current db backend supports mysql", cfg.DBDriver)
	}
	if cfg.DBDSN == "" {
		return nil, fmt.Errorf("db.dsn is required")
	}

	manager := newManagerFromConfig(cfg)
	if cfg.DBInitSchema {
		if err := applySchema(ctx, manager, defaultSchemaSQL); err != nil {
			return nil, fmt.Errorf("apply db schema: %w", err)
		}
	}

	ruleSetTableDAO := newGosharedRuleSetTableDAO(manager)
	namespaceTableDAO := newGosharedNamespaceTableDAO(manager)
	return &dbImpl{
		manager:             manager,
		ruleSetRepository:   newRuleSetRepository(ruleSetTableDAO),
		namespaceRepository: newNamespaceRepository(namespaceTableDAO),
	}, nil
}

func newManagerFromConfig(cfg config.Config) dbspi.Manager {
	return dbhelper.NewManager(dbspi.DatabaseConfig{
		DatabaseGroups: map[string]dbspi.DatabaseGroupConfig{
			dbspi.DefaultDatabaseGroupKey: {
				DSN:                    cfg.DBDSN,
				Debug:                  cfg.DBDebug,
				MaxOpenConns:           cfg.DBMaxOpenConns,
				MaxIdleConns:           cfg.DBMaxIdleConns,
				ConnMaxLifetimeSeconds: cfg.DBConnMaxLifetimeSeconds,
			},
		},
	}, dbhelper.WithCommonFieldAutoFill(true), dbhelper.WithCommonFieldTimeProvider(func(context.Context) uint64 {
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

func (d *dbImpl) GetManager() dbspi.Manager {
	return d.manager
}
