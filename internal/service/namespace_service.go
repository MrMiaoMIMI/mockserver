package service

import (
	"context"
	"fmt"
	"strings"

	"mockserver/internal/dao"
	"mockserver/internal/model/bo"
)

type namespaceService struct {
	namespaces dao.NamespaceRepository
}

func NewNamespaceService(namespaceRepository dao.NamespaceRepository) NamespaceService {
	return &namespaceService{namespaces: namespaceRepository}
}

func (s *namespaceService) UpsertNamespace(ctx context.Context, namespace bo.Namespace) (bo.Namespace, error) {
	namespace.ID = normalizeNamespaceID(namespace.ID)
	if namespace.ID == "" {
		id, err := s.newNamespaceID(ctx, namespace)
		if err != nil {
			return bo.Namespace{}, err
		}
		namespace.ID = id
	}
	normalized, err := normalizeNamespace(namespace)
	if err != nil {
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
		return bo.Namespace{}, fmt.Errorf("namespace id is required")
	}
	namespace, ok, err := s.namespaces.GetNamespace(ctx, normalizedID)
	if err != nil {
		return bo.Namespace{}, err
	}
	if !ok {
		if normalizedID == "default" {
			return s.EnsureDefaultNamespace(ctx)
		}
		return bo.Namespace{}, fmt.Errorf("namespace %s not found", normalizedID)
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
		if item.ID == "default" {
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

func (s *namespaceService) newNamespaceID(ctx context.Context, namespace bo.Namespace) (string, error) {
	base := slugifyNamespaceID(namespace.Name)
	if base == "" {
		base = "namespace"
	}
	const maxIDLength = 128
	const suffixLength = 8
	maxBaseLength := maxIDLength - suffixLength - 1
	if len(base) > maxBaseLength {
		base = strings.Trim(base[:maxBaseLength], "-")
	}
	if base == "" {
		base = "namespace"
	}
	for i := 0; i < 20; i++ {
		candidate := base + "-" + randomIDToken()
		_, exists, err := s.namespaces.GetNamespace(ctx, candidate)
		if err != nil {
			return "", err
		}
		if exists {
			continue
		}
		return candidate, nil
	}
	return "", fmt.Errorf("generate unique namespace id failed")
}
