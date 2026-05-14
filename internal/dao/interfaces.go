package dao

import (
	"context"
	"errors"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

var (
	ErrNotFound = errors.New("dao record not found")
	ErrConflict = errors.New("dao optimistic lock conflict")
)

type RuleSetRepository interface {
	UpsertDraft(ctx context.Context, ruleSet bo.RuleSet) (bo.RuleSet, error)
	GetDraft(ctx context.Context, id string) (bo.RuleSet, bool, error)
	ListDrafts(ctx context.Context) ([]bo.RuleSet, error)
	Publish(ctx context.Context, id string, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error)
	GetPublished(ctx context.Context, id string) (bo.PublishedRuleSetSnapshot, bool, error)
	GetPublishedSnapshot(ctx context.Context, id, snapshotID string) (bo.PublishedRuleSetSnapshot, bool, error)
	Rollback(ctx context.Context, id, snapshotID string, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error)
	ListPublished(ctx context.Context) ([]bo.PublishedRuleSetSnapshot, error)
	ListPublishedSnapshots(ctx context.Context, id string) ([]bo.PublishedRuleSetSnapshot, error)
}

type NamespaceRepository interface {
	UpsertNamespace(ctx context.Context, namespace bo.Namespace) (bo.Namespace, error)
	GetNamespace(ctx context.Context, id string) (bo.Namespace, bool, error)
	ListNamespaces(ctx context.Context) ([]bo.Namespace, error)
}

type TrafficRepository interface {
	CreateTrafficEvent(ctx context.Context, event bo.TrafficEvent) (bo.TrafficEvent, error)
	GetTrafficEvent(ctx context.Context, id uint64) (bo.TrafficEvent, bool, error)
	ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error)
}

type ruleSetTableDAO interface {
	UpsertDraft(ctx context.Context, draft modeldo.RuleSetDraft, expectedVersion int) error
	GetDraft(ctx context.Context, id string) (modeldo.RuleSetDraft, bool, error)
	ListDrafts(ctx context.Context) ([]modeldo.RuleSetDraft, error)
	PublishSnapshot(ctx context.Context, snapshot modeldo.PublishedRuleSetSnapshot) (modeldo.PublishedRuleSetSnapshot, error)
	GetCurrentPublished(ctx context.Context, id string) (modeldo.PublishedRuleSetSnapshot, bool, error)
	GetPublishedSnapshot(ctx context.Context, id string, snapshotID string) (modeldo.PublishedRuleSetSnapshot, bool, error)
	ListPublished(ctx context.Context) ([]modeldo.PublishedRuleSetSnapshot, error)
	ListPublishedSnapshots(ctx context.Context, id string) ([]modeldo.PublishedRuleSetSnapshot, error)
}

type namespaceTableDAO interface {
	UpsertNamespace(ctx context.Context, namespace modeldo.NamespaceConfig) error
	GetNamespace(ctx context.Context, id string) (modeldo.NamespaceConfig, bool, error)
	ListNamespaces(ctx context.Context) ([]modeldo.NamespaceConfig, error)
}

type trafficTableDAO interface {
	CreateTrafficEvent(ctx context.Context, event modeldo.TrafficEvent, indexes []modeldo.TrafficEventIndex) (modeldo.TrafficEvent, error)
	GetTrafficEvent(ctx context.Context, id uint64) (modeldo.TrafficEvent, bool, error)
	ListTrafficEvents(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64) ([]modeldo.TrafficEvent, uint64, error)
	ListTrafficEventIndexesByEventIDs(ctx context.Context, eventIDs []uint64) (map[uint64][]modeldo.TrafficEventIndex, error)
	ListTrafficEventIDsByIndexFilter(ctx context.Context, query bo.TrafficQuery, filter bo.TrafficIndexFilter, valueHash uint64) ([]uint64, error)
	TrafficStats(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64) (bo.TrafficStats, error)
}

type DB interface {
	GetRuleSetRepository() RuleSetRepository
	GetNamespaceRepository() NamespaceRepository
	GetTrafficRepository() TrafficRepository
	GetManager() dbspi.Manager
}
