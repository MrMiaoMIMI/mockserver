package service

import (
	"context"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type namespaceService struct {
	namespaces dao.NamespaceRepository
}

func NewNamespaceService(namespaceRepository dao.NamespaceRepository) NamespaceService {
	return &namespaceService{namespaces: namespaceRepository}
}

func (s *namespaceService) UpsertNamespace(ctx context.Context, namespace bo.Namespace) (bo.Namespace, error) {
	namespace.Name = normalizeNamespaceID(namespace.Name)
	if namespace.Name == "" {
		return bo.Namespace{}, validationErrorf("namespace name is required")
	}
	normalized, err := normalizeNamespace(namespace)
	if err != nil {
		return bo.Namespace{}, err
	}
	if err := s.validateNamespaceWrite(ctx, normalized); err != nil {
		return bo.Namespace{}, err
	}
	return s.namespaces.UpsertNamespace(ctx, normalized)
}

func (s *namespaceService) EnsureDefaultNamespace(ctx context.Context) (bo.Namespace, error) {
	namespace, ok, err := s.namespaces.GetNamespace(ctx, "default")
	if err != nil {
		return bo.Namespace{}, err
	}
	if ok {
		return namespace, nil
	}
	return s.namespaces.UpsertNamespace(ctx, bo.DefaultNamespace("default"))
}

func (s *namespaceService) GetNamespace(ctx context.Context, id string) (bo.Namespace, error) {
	normalizedID := normalizeNamespaceID(id)
	if normalizedID == "" {
		return bo.Namespace{}, validationErrorf("namespace name is required")
	}
	namespace, ok, err := s.namespaces.GetNamespace(ctx, normalizedID)
	if err != nil {
		return bo.Namespace{}, err
	}
	if !ok {
		if normalizedID == "default" {
			return s.EnsureDefaultNamespace(ctx)
		}
		return bo.Namespace{}, notFoundErrorf("namespace %s not found", normalizedID)
	}
	return namespace, nil
}

func (s *namespaceService) ListNamespaces(ctx context.Context) ([]bo.Namespace, error) {
	items, err := s.namespaces.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	hasDefault := false
	for _, item := range items {
		if item.Name == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		defaultNamespace, err := s.EnsureDefaultNamespace(ctx)
		if err != nil {
			return nil, err
		}
		items = append([]bo.Namespace{defaultNamespace}, items...)
	}
	return items, nil
}

func (s *namespaceService) validateNamespaceWrite(ctx context.Context, namespace bo.Namespace) error {
	current, exists, err := s.namespaces.GetNamespace(ctx, namespace.Name)
	if err != nil {
		return err
	}
	if exists {
		if namespace.Version <= 0 {
			return conflictErrorf("namespace %s already exists; version is required for updates", namespace.Name)
		}
		if namespace.Version != current.Version {
			return conflictErrorf("namespace %s version conflict: got %d, current %d", namespace.Name, namespace.Version, current.Version)
		}
	} else if namespace.Version != 0 {
		return conflictErrorf("namespace %s does not exist; create requests must omit version", namespace.Name)
	}
	return nil
}
