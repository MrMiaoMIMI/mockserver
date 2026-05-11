package dao

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"mockserver/internal/config"
	"mockserver/internal/model/bo"
	modeldo "mockserver/internal/model/do"
)

func TestRuleSetRepositoryWithMySQL(t *testing.T) {
	dsn := os.Getenv("MOCKSERVER_MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MOCKSERVER_MYSQL_TEST_DSN to run MySQL integration test")
	}

	ctx := context.Background()
	db, err := NewDB(ctx, config.Config{
		DBDriver:     "mysql",
		DBDSN:        dsn,
		DBInitSchema: true,
	})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	manager := db.GetManager()

	id := fmt.Sprintf("mysql-repository-%d", time.Now().UnixMilli())
	namespaceID := id + "-namespace"
	cleanupMySQLRuleSet(t, ctx, manager, id)
	cleanupMySQLNamespace(t, ctx, manager, namespaceID)
	t.Cleanup(func() {
		cleanupMySQLRuleSet(t, ctx, manager, id)
		cleanupMySQLNamespace(t, ctx, manager, namespaceID)
	})

	repository := db.GetRuleSetRepository()
	namespaceRepository := db.GetNamespaceRepository()
	if _, err := namespaceRepository.UpsertNamespace(ctx, bo.Namespace{
		ID:          namespaceID,
		Name:        "MySQL namespace",
		Description: "namespace repository integration test",
	}); err != nil {
		t.Fatalf("UpsertNamespace() error = %v", err)
	}
	loadedNamespace, ok, err := namespaceRepository.GetNamespace(ctx, namespaceID)
	if err != nil {
		t.Fatalf("GetNamespace() error = %v", err)
	}
	if !ok || loadedNamespace.ID != namespaceID {
		t.Fatalf("unexpected namespace: ok=%v namespace=%#v", ok, loadedNamespace)
	}

	ruleSet := testRuleSet(id, "mysql-rule", "/api/v1/mysql", 1)
	saved, err := repository.UpsertDraft(ctx, ruleSet)
	if err != nil {
		t.Fatalf("UpsertDraft() error = %v", err)
	}
	if saved.Version != 1 {
		t.Fatalf("unexpected initial version: %d", saved.Version)
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
		Operator: "mysql-tester",
		Reason:   "mysql integration test",
		TraceID:  "trace-mysql",
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	reloadedDB, err := NewDB(ctx, config.Config{
		DBDriver: "mysql",
		DBDSN:    dsn,
	})
	if err != nil {
		t.Fatalf("reload NewDB() error = %v", err)
	}
	reloadedRepository := reloadedDB.GetRuleSetRepository()
	published, ok, err := reloadedRepository.GetPublished(ctx, updated.ID)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}
	if !ok {
		t.Fatalf("expected current published snapshot")
	}
	if published.SnapshotID != snapshot.SnapshotID {
		t.Fatalf("unexpected snapshot id: got=%s want=%s", published.SnapshotID, snapshot.SnapshotID)
	}
	if published.Audit == nil || published.Audit.Operator != "mysql-tester" || published.Audit.TraceID != "trace-mysql" {
		t.Fatalf("unexpected published audit: %#v", published.Audit)
	}
	snapshots, err := reloadedRepository.ListPublishedSnapshots(ctx, updated.ID)
	if err != nil {
		t.Fatalf("ListPublishedSnapshots() error = %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("unexpected snapshot count: %d", len(snapshots))
	}
	publishedItems, err := reloadedRepository.ListPublished(ctx)
	if err != nil {
		t.Fatalf("ListPublished() error = %v", err)
	}
	if len(publishedItems) == 0 {
		t.Fatalf("expected published rulesets")
	}
}

func cleanupMySQLRuleSet(t *testing.T, ctx context.Context, manager dbspi.Manager, id string) {
	t.Helper()
	tableStore := dbhelper.NewTableStore(&modeldo.RuleSetDraft{}, dbhelper.WithManager(manager))
	sqlStore, ok := dbhelper.AsSQLTableStore(tableStore)
	if !ok {
		t.Fatalf("table store does not support raw SQL")
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_published_rule_set_tab WHERE ruleset_id = ?", id); err != nil {
		t.Fatalf("cleanup current published %s: %v", id, err)
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_published_snapshot_tab WHERE ruleset_id = ?", id); err != nil {
		t.Fatalf("cleanup published snapshots %s: %v", id, err)
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_rule_set_draft_tab WHERE ruleset_id = ?", id); err != nil {
		t.Fatalf("cleanup draft %s: %v", id, err)
	}
}

func cleanupMySQLNamespace(t *testing.T, ctx context.Context, manager dbspi.Manager, id string) {
	t.Helper()
	tableStore := dbhelper.NewTableStore(&modeldo.RuleSetDraft{}, dbhelper.WithManager(manager))
	sqlStore, ok := dbhelper.AsSQLTableStore(tableStore)
	if !ok {
		t.Fatalf("table store does not support raw SQL")
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_namespace_tab WHERE namespace_id = ?", id); err != nil {
		t.Fatalf("cleanup namespace %s: %v", id, err)
	}
}
