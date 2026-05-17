package controller_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type testRuleSetRepository struct {
	mu         sync.RWMutex
	drafts     map[string]bo.RuleSet
	published  map[string]testPublishedRecord
	snapshots  map[string][]bo.PublishedRuleSetSnapshot
	namespaces map[string]bo.Namespace
	revision   uint64
}

type testPublishedRecord struct {
	snapshot bo.PublishedRuleSetSnapshot
}

func newTestRuleSetRepository() *testRuleSetRepository {
	return &testRuleSetRepository{
		drafts:    make(map[string]bo.RuleSet),
		published: make(map[string]testPublishedRecord),
		snapshots: make(map[string][]bo.PublishedRuleSetSnapshot),
		namespaces: map[string]bo.Namespace{
			"default": bo.DefaultNamespace("default"),
		},
	}
}

func (s *testRuleSetRepository) UpsertDraft(ctx context.Context, ruleSet bo.RuleSet) (bo.RuleSet, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.drafts[ruleSet.ID]
	if exists {
		ruleSet.Version = current.Version + 1
	} else {
		ruleSet.Version = 1
	}
	s.drafts[ruleSet.ID] = ruleSet
	return ruleSet, nil
}

func (s *testRuleSetRepository) GetDraft(ctx context.Context, id string) (bo.RuleSet, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	ruleSet, ok := s.drafts[id]
	return ruleSet, ok, nil
}

func (s *testRuleSetRepository) ListDrafts(ctx context.Context) ([]bo.RuleSet, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]bo.RuleSet, 0, len(s.drafts))
	for _, item := range s.drafts {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (s *testRuleSetRepository) Publish(ctx context.Context, id string, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	ruleSet, ok := s.drafts[id]
	if !ok {
		return bo.PublishedRuleSetSnapshot{}, fmt.Errorf("ruleset %s not found", id)
	}
	return s.publishSnapshot(ruleSet, audit)
}

func (s *testRuleSetRepository) GetPublished(ctx context.Context, id string) (bo.PublishedRuleSetSnapshot, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.published[id]
	if !ok {
		return bo.PublishedRuleSetSnapshot{}, false, nil
	}
	return record.snapshot, true, nil
}

func (s *testRuleSetRepository) GetPublishedSnapshot(ctx context.Context, id, snapshotID string) (bo.PublishedRuleSetSnapshot, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, snapshot := range s.snapshots[id] {
		if snapshot.SnapshotID == snapshotID {
			return snapshot, true, nil
		}
	}
	return bo.PublishedRuleSetSnapshot{}, false, nil
}

func (s *testRuleSetRepository) Rollback(ctx context.Context, id, snapshotID string, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, snapshot := range s.snapshots[id] {
		if snapshot.SnapshotID == snapshotID {
			return s.publishSnapshot(snapshot.RuleSet, audit)
		}
	}
	return bo.PublishedRuleSetSnapshot{}, fmt.Errorf("snapshot %s for ruleset %s not found", snapshotID, id)
}

func (s *testRuleSetRepository) ListPublished(ctx context.Context) ([]bo.PublishedRuleSetSnapshot, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]bo.PublishedRuleSetSnapshot, 0, len(s.published))
	for _, item := range s.published {
		items = append(items, item.snapshot)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].RuleSet.ID < items[j].RuleSet.ID
	})
	return items, nil
}

func (s *testRuleSetRepository) ListPublishedSnapshots(ctx context.Context, id string) ([]bo.PublishedRuleSetSnapshot, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]bo.PublishedRuleSetSnapshot(nil), s.snapshots[id]...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].PublishedAt.After(items[j].PublishedAt)
	})
	return items, nil
}

func (s *testRuleSetRepository) UpsertNamespace(ctx context.Context, namespace bo.Namespace) (bo.Namespace, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.namespaces[namespace.ID]
	if exists {
		namespace.Version = current.Version + 1
	} else {
		namespace.Version = 1
	}
	s.namespaces[namespace.ID] = namespace
	return namespace, nil
}

func (s *testRuleSetRepository) GetNamespace(ctx context.Context, id string) (bo.Namespace, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	namespace, ok := s.namespaces[id]
	return namespace, ok, nil
}

func (s *testRuleSetRepository) ListNamespaces(ctx context.Context) ([]bo.Namespace, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]bo.Namespace, 0, len(s.namespaces))
	for _, item := range s.namespaces {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (s *testRuleSetRepository) PublishedRevision() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.revision
}

func (s *testRuleSetRepository) publishSnapshot(ruleSet bo.RuleSet, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	publishedAt := time.Now().UTC()
	snapshot := bo.PublishedRuleSetSnapshot{
		SnapshotID:  fmt.Sprintf("snap_%d", publishedAt.UnixNano()),
		PublishedAt: publishedAt,
		RuleSet:     ruleSet,
		Audit:       cloneTestAuditInfo(audit),
	}
	s.published[ruleSet.ID] = testPublishedRecord{
		snapshot: snapshot,
	}
	s.snapshots[ruleSet.ID] = append(s.snapshots[ruleSet.ID], snapshot)
	s.revision++
	return snapshot, nil
}

func cloneTestAuditInfo(audit *bo.AuditInfo) *bo.AuditInfo {
	if audit == nil {
		return nil
	}
	cloned := *audit
	return &cloned
}
