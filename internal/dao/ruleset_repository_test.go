package dao

import (
	"context"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

func TestRuleSetRepositoryPublishesAuditSnapshot(t *testing.T) {
	ruleSetDAO := newFakeRuleSetTableDAO()
	repository := newRuleSetRepository(ruleSetDAO)
	ctx := context.WithValue(context.Background(), fakeCtxKey("trace_id"), "trace-from-test")

	ruleSet := testRuleSet("db-repository", "db-rule", "/api/v1/db", 1)
	saved, err := repository.UpsertDraft(ctx, ruleSet)
	if err != nil {
		t.Fatalf("UpsertDraft() error = %v", err)
	}
	if saved.Version != 1 {
		t.Fatalf("unexpected saved version: %d", saved.Version)
	}

	updated, err := repository.UpsertDraft(ctx, saved)
	if err != nil {
		t.Fatalf("UpsertDraft(update) error = %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("unexpected updated version: %d", updated.Version)
	}

	snapshot, err := repository.Publish(ctx, updated.ID, &bo.AuditInfo{
		Action:   "publish",
		Operator: "db-tester",
		Reason:   "db repository test",
		TraceID:  "trace-db",
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if snapshot.Audit == nil || snapshot.Audit.Operator != "db-tester" || snapshot.Audit.TraceID != "trace-db" {
		t.Fatalf("unexpected snapshot audit: %#v", snapshot.Audit)
	}
	if len(ruleSetDAO.observedCtxValues) == 0 {
		t.Fatalf("expected dao to receive request context")
	}
	for _, value := range ruleSetDAO.observedCtxValues {
		if value != "trace-from-test" {
			t.Fatalf("unexpected observed context value: %#v", value)
		}
	}

	published, ok, err := repository.GetPublished(ctx, updated.ID)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}
	if !ok {
		t.Fatalf("expected published snapshot")
	}
	if published.SnapshotID != snapshot.SnapshotID {
		t.Fatalf("unexpected current snapshot id: got=%s want=%s", published.SnapshotID, snapshot.SnapshotID)
	}
	publishedItems, err := repository.ListPublished(ctx)
	if err != nil {
		t.Fatalf("ListPublished() error = %v", err)
	}
	if len(publishedItems) != 1 {
		t.Fatalf("unexpected published count: %d", len(publishedItems))
	}
}

func TestNamespaceRepositoryUpsertAndList(t *testing.T) {
	namespaceDAO := newFakeRuleSetTableDAO()
	repository := newNamespaceRepository(namespaceDAO)
	ctx := context.WithValue(context.Background(), fakeCtxKey("trace_id"), "trace-from-test")

	saved, err := repository.UpsertNamespace(ctx, bo.Namespace{
		ID:          "tenant-a",
		Name:        "Tenant A",
		Description: "integration tenant",
	})
	if err != nil {
		t.Fatalf("UpsertNamespace() error = %v", err)
	}
	if saved.ID != "tenant-a" {
		t.Fatalf("unexpected namespace id: %s", saved.ID)
	}

	loaded, ok, err := repository.GetNamespace(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("GetNamespace() error = %v", err)
	}
	if !ok || loaded.Name != "Tenant A" {
		t.Fatalf("unexpected loaded namespace: ok=%v namespace=%#v", ok, loaded)
	}

	items, err := repository.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListNamespaces() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected namespace count: %d", len(items))
	}
}

type fakeCtxKey string

type fakeRuleSetTableDAO struct {
	drafts            map[string]modeldo.RuleSetDraft
	current           map[string]string
	snapshots         map[string]modeldo.PublishedRuleSetSnapshot
	namespaces        map[string]modeldo.NamespaceConfig
	observedCtxValues []any
}

func newFakeRuleSetTableDAO() *fakeRuleSetTableDAO {
	return &fakeRuleSetTableDAO{
		drafts:     make(map[string]modeldo.RuleSetDraft),
		current:    make(map[string]string),
		snapshots:  make(map[string]modeldo.PublishedRuleSetSnapshot),
		namespaces: make(map[string]modeldo.NamespaceConfig),
	}
}

func (d *fakeRuleSetTableDAO) UpsertDraft(ctx context.Context, draft modeldo.RuleSetDraft, expectedVersion int) error {
	d.observeContext(ctx)
	current, exists := d.drafts[draft.RuleSetID]
	if expectedVersion == 0 && exists {
		return ErrConflict
	}
	if expectedVersion > 0 && (!exists || current.Version != expectedVersion) {
		return ErrConflict
	}
	d.drafts[draft.RuleSetID] = draft
	return nil
}

func (d *fakeRuleSetTableDAO) GetDraft(ctx context.Context, id string) (modeldo.RuleSetDraft, bool, error) {
	d.observeContext(ctx)
	item, ok := d.drafts[id]
	return item, ok, nil
}

func (d *fakeRuleSetTableDAO) ListDrafts(ctx context.Context) ([]modeldo.RuleSetDraft, error) {
	d.observeContext(ctx)
	items := make([]modeldo.RuleSetDraft, 0, len(d.drafts))
	for _, item := range d.drafts {
		items = append(items, item)
	}
	return items, nil
}

func (d *fakeRuleSetTableDAO) PublishSnapshot(ctx context.Context, snapshot modeldo.PublishedRuleSetSnapshot) error {
	d.observeContext(ctx)
	d.snapshots[snapshot.SnapshotID] = snapshot
	d.current[snapshot.RuleSetID] = snapshot.SnapshotID
	return nil
}

func (d *fakeRuleSetTableDAO) GetCurrentPublished(ctx context.Context, id string) (modeldo.PublishedRuleSetSnapshot, bool, error) {
	d.observeContext(ctx)
	snapshotID, ok := d.current[id]
	if !ok {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	item, ok := d.snapshots[snapshotID]
	return item, ok, nil
}

func (d *fakeRuleSetTableDAO) GetPublishedSnapshot(ctx context.Context, id string, snapshotID string) (modeldo.PublishedRuleSetSnapshot, bool, error) {
	d.observeContext(ctx)
	item, ok := d.snapshots[snapshotID]
	if !ok || item.RuleSetID != id {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	return item, true, nil
}

func (d *fakeRuleSetTableDAO) ListPublished(ctx context.Context) ([]modeldo.PublishedRuleSetSnapshot, error) {
	d.observeContext(ctx)
	items := make([]modeldo.PublishedRuleSetSnapshot, 0, len(d.current))
	for _, snapshotID := range d.current {
		items = append(items, d.snapshots[snapshotID])
	}
	return items, nil
}

func (d *fakeRuleSetTableDAO) ListPublishedSnapshots(ctx context.Context, id string) ([]modeldo.PublishedRuleSetSnapshot, error) {
	d.observeContext(ctx)
	items := make([]modeldo.PublishedRuleSetSnapshot, 0)
	for _, item := range d.snapshots {
		if item.RuleSetID == id {
			items = append(items, item)
		}
	}
	return items, nil
}

func (d *fakeRuleSetTableDAO) UpsertNamespace(ctx context.Context, namespace modeldo.NamespaceConfig) error {
	d.observeContext(ctx)
	d.namespaces[namespace.NamespaceID] = namespace
	return nil
}

func (d *fakeRuleSetTableDAO) GetNamespace(ctx context.Context, id string) (modeldo.NamespaceConfig, bool, error) {
	d.observeContext(ctx)
	item, ok := d.namespaces[id]
	return item, ok, nil
}

func (d *fakeRuleSetTableDAO) ListNamespaces(ctx context.Context) ([]modeldo.NamespaceConfig, error) {
	d.observeContext(ctx)
	items := make([]modeldo.NamespaceConfig, 0, len(d.namespaces))
	for _, item := range d.namespaces {
		items = append(items, item)
	}
	return items, nil
}

func (d *fakeRuleSetTableDAO) observeContext(ctx context.Context) {
	d.observedCtxValues = append(d.observedCtxValues, ctx.Value(fakeCtxKey("trace_id")))
}
