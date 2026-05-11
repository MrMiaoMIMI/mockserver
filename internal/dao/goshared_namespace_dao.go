package dao

import (
	"context"
	"fmt"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	modelfmo "github.com/MrMiaoMIMI/mockserver/internal/model/fmo"
)

type gosharedNamespaceTableDAO struct {
	namespaceStore  dbspi.SoftDeleteTableStore[*modeldo.NamespaceConfig]
	namespaceFields modelfmo.NamespaceConfigFields
}

func newGosharedNamespaceTableDAO(manager dbspi.Manager) namespaceTableDAO {
	return &gosharedNamespaceTableDAO{
		namespaceStore:  dbhelper.NewSoftDeleteTableStore(&modeldo.NamespaceConfig{}, dbhelper.WithManager(manager)),
		namespaceFields: modelfmo.NewNamespaceConfigFields(),
	}
}

func (d *gosharedNamespaceTableDAO) UpsertNamespace(ctx context.Context, namespace modeldo.NamespaceConfig) error {
	exists, _, err := d.namespaceStore.ExistsById(ctx, namespace.NamespaceID)
	if err != nil {
		return fmt.Errorf("get namespace %s: %w", namespace.NamespaceID, err)
	}
	if !exists {
		if err := d.namespaceStore.Create(ctx, &namespace); err != nil {
			return fmt.Errorf("insert namespace %s: %w", namespace.NamespaceID, err)
		}
		return nil
	}

	updater := dbhelper.NewUpdater().
		Set(d.namespaceFields.NamespaceJSON, namespace.NamespaceJSON)
	if err := d.namespaceStore.UpdateById(ctx, namespace.NamespaceID, updater); err != nil {
		return fmt.Errorf("update namespace %s: %w", namespace.NamespaceID, err)
	}
	return nil
}

func (d *gosharedNamespaceTableDAO) GetNamespace(ctx context.Context, id string) (modeldo.NamespaceConfig, bool, error) {
	exists, namespace, err := d.namespaceStore.ExistsByIdNotDeleted(ctx, id)
	if err != nil {
		return modeldo.NamespaceConfig{}, false, fmt.Errorf("get namespace %s: %w", id, err)
	}
	if !exists || namespace == nil {
		return modeldo.NamespaceConfig{}, false, nil
	}
	return *namespace, true, nil
}

func (d *gosharedNamespaceTableDAO) ListNamespaces(ctx context.Context) ([]modeldo.NamespaceConfig, error) {
	items, err := d.namespaceStore.FindNotDeleted(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	result := make([]modeldo.NamespaceConfig, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}
