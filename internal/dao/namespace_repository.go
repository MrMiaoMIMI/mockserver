package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

type namespaceRepository struct {
	tableDAO namespaceTableDAO
}

func newNamespaceRepository(tableDAO namespaceTableDAO) NamespaceRepository {
	return &namespaceRepository{tableDAO: tableDAO}
}

func (r *namespaceRepository) UpsertNamespace(ctx context.Context, namespace bo.Namespace) (bo.Namespace, error) {
	raw, err := json.Marshal(namespace)
	if err != nil {
		return bo.Namespace{}, fmt.Errorf("marshal namespace %s: %w", namespace.ID, err)
	}
	if err := r.tableDAO.UpsertNamespace(ctx, modeldo.NamespaceConfig{
		NamespaceCode: namespace.ID,
		Version:       1,
		NamespaceJSON: string(raw),
	}); err != nil {
		return bo.Namespace{}, err
	}
	if record, ok, err := r.tableDAO.GetNamespace(ctx, namespace.ID); err != nil {
		return bo.Namespace{}, err
	} else if ok {
		namespace.DBID = record.Id
	}
	return namespace, nil
}

func (r *namespaceRepository) GetNamespace(ctx context.Context, id string) (bo.Namespace, bool, error) {
	record, ok, err := r.tableDAO.GetNamespace(ctx, id)
	if err != nil || !ok {
		return bo.Namespace{}, false, err
	}
	namespace, err := decodeNamespaceRecord(record.NamespaceJSON)
	if err != nil {
		return bo.Namespace{}, false, err
	}
	namespace.DBID = record.Id
	if namespace.ID == "" {
		namespace.ID = record.NamespaceCode
	}
	return namespace, true, nil
}

func (r *namespaceRepository) ListNamespaces(ctx context.Context) ([]bo.Namespace, error) {
	records, err := r.tableDAO.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]bo.Namespace, 0, len(records))
	for _, record := range records {
		namespace, err := decodeNamespaceRecord(record.NamespaceJSON)
		if err != nil {
			return nil, err
		}
		namespace.DBID = record.Id
		if namespace.ID == "" {
			namespace.ID = record.NamespaceCode
		}
		items = append(items, namespace)
	}
	return items, nil
}

func decodeNamespaceRecord(raw string) (bo.Namespace, error) {
	var namespace bo.Namespace
	if err := json.Unmarshal([]byte(raw), &namespace); err != nil {
		return bo.Namespace{}, fmt.Errorf("decode namespace json: %w", err)
	}
	return namespace, nil
}
