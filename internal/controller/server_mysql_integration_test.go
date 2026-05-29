package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/config"
	"github.com/MrMiaoMIMI/mockserver/internal/controller"
	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	"github.com/MrMiaoMIMI/mockserver/internal/observability"
	"github.com/MrMiaoMIMI/mockserver/internal/router"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
	"github.com/MrMiaoMIMI/mockserver/internal/view"
)

func TestAdminRuntimeFlowWithMySQLRepository(t *testing.T) {
	mysqlConfig, ok := mysqlTestConfigFromEnv(t)
	if !ok {
		t.Skip("set MOCKSERVER_MYSQL_TEST_HOST, MOCKSERVER_MYSQL_TEST_USER, and MOCKSERVER_MYSQL_TEST_DATABASE_NAME to run MySQL integration test")
	}

	ctx := context.Background()
	db, err := dao.NewDB(mysqlConfig)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	manager := db.GetManager()

	id := fmt.Sprintf("mysql-e2e-%d", time.Now().UnixMilli())
	cleanupMySQLRuleset(t, ctx, manager, id)
	namespaceID := ""
	t.Cleanup(func() {
		cleanupMySQLRuleset(t, ctx, manager, id)
		if namespaceID != "" {
			cleanupMySQLRuleset(t, ctx, manager, namespaceID)
		}
	})

	handler := newMySQLBackedHandler(db.GetRuleSetRepository(), db.GetNamespaceRepository())
	namespaceResp := doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/namespaces", map[string]any{
		"name": id + "-namespace",
		"policies": httpNamespacePolicies(
			httpStaticActionPayload(404, map[string]any{"message": "no mock ruleset matched"}),
			httpStaticActionPayload(404, map[string]any{"message": "no mock rule matched"}),
		),
	}, http.StatusOK)
	var namespaceEnvelope struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(readBody(t, namespaceResp), &namespaceEnvelope); err != nil {
		t.Fatalf("decode namespace response: %v", err)
	}
	namespaceID = namespaceEnvelope.Data.Name
	if namespaceID != id+"-namespace" {
		t.Fatalf("unexpected namespace name: %s", namespaceID)
	}

	upsertBody := map[string]any{
		"id":        id,
		"name":      "mysql e2e",
		"enabled":   true,
		"protocol":  "http",
		"namespace": namespaceID,
		"selector":  httpPathSelectorBody("/api/v1/mysql-e2e"),
		"rules": []map[string]any{
			{
				"id":       "mysql-e2e-rule",
				"name":     "Mysql E2e Rule",
				"enabled":  true,
				"priority": 100,
				"when": map[string]any{
					"all": []map[string]any{
						{"field": "request.method", "op": "eq", "value": "GET"},
						{"field": "request.path", "op": "eq", "value": "/api/v1/mysql-e2e"},
					},
				},
				"action": httpStaticActionPayload(202, map[string]any{"store": "mysql"}),
			},
		},
	}

	doJSON(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets", upsertBody, http.StatusOK)
	publishResp := doJSONWithHeaders(t, handler, http.MethodPost, "/mockserver/api/v1/admin/rulesets/"+id+"/publish", map[string]any{
		"reason": "mysql e2e publish",
	}, map[string]string{
		"X-Mockserver-Operator": "mysql-e2e@example.com",
		"X-Trace-ID":            "trace-mysql-e2e",
	}, http.StatusOK)
	publishBody := readBody(t, publishResp)
	assertBytesContain(t, publishBody, `"action":"publish"`)
	assertBytesContain(t, publishBody, `"operator":"mysql-e2e@example.com"`)
	assertBytesContain(t, publishBody, `"reason":"mysql e2e publish"`)

	runtimePath := "/mockserver/runtime/" + namespaceID + "/http/api/v1/mysql-e2e"
	runtimeResp := doJSON(t, handler, http.MethodGet, runtimePath, nil, http.StatusAccepted)
	assertBytesContain(t, readBody(t, runtimeResp), `"store":"mysql"`)

	reloadedDB, err := dao.NewDB(mysqlConfig)
	if err != nil {
		t.Fatalf("reload NewDB() error = %v", err)
	}
	reloadedHandler := newMySQLBackedHandler(reloadedDB.GetRuleSetRepository(), reloadedDB.GetNamespaceRepository())
	reloadedRuntimeResp := doJSON(t, reloadedHandler, http.MethodGet, runtimePath, nil, http.StatusAccepted)
	assertBytesContain(t, readBody(t, reloadedRuntimeResp), `"store":"mysql"`)

	publishedResp := doJSON(t, reloadedHandler, http.MethodGet, "/mockserver/api/v1/admin/published/rulesets/"+id, nil, http.StatusOK)
	publishedBody := readBody(t, publishedResp)
	assertBytesContain(t, publishedBody, `"id":"`+id+`"`)
	assertBytesContain(t, publishedBody, `"trace_id":"trace-mysql-e2e"`)
}

func newMySQLBackedHandler(ruleSetRepository dao.RuleSetRepository, namespaceRepository dao.NamespaceRepository) http.Handler {
	namespaceService := service.NewNamespaceService(namespaceRepository)
	ruleSetService := service.NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := service.NewRuntimeService(ruleSetRepository, namespaceRepository, namespaceService)
	ruleSetView := view.NewRuleSetView(ruleSetService)
	namespaceView := view.NewNamespaceView(namespaceService)
	runtimeView := view.NewRuntimeView(runtimeService)
	runtimeMetrics := observability.NewRuntimeMetrics()
	adminController := controller.NewAdminController(ruleSetView, namespaceView)
	runtimeController := controller.NewRuntimeController(runtimeView, runtimeMetrics)
	metricsController := controller.NewMetricsController(runtimeMetrics)
	return router.New(adminController, runtimeController, router.AuthConfig{}, metricsController)
}

func mysqlTestConfigFromEnv(t testing.TB) (config.Config, bool) {
	t.Helper()
	host := os.Getenv("MOCKSERVER_MYSQL_TEST_HOST")
	user := os.Getenv("MOCKSERVER_MYSQL_TEST_USER")
	databaseName := os.Getenv("MOCKSERVER_MYSQL_TEST_DATABASE_NAME")
	if host == "" || user == "" || databaseName == "" {
		return config.Config{}, false
	}
	portText := os.Getenv("MOCKSERVER_MYSQL_TEST_PORT")
	if portText == "" {
		portText = "3306"
	}
	port, err := strconv.ParseUint(portText, 10, 0)
	if err != nil {
		t.Fatalf("parse mysql port %q: %v", portText, err)
	}
	return config.Config{
		DB: dbspi.DatabaseConfig{
			DatabaseGroups: map[string]dbspi.DatabaseGroupConfig{
				dbspi.DefaultDatabaseGroupKey: {
					Host:         host,
					Port:         uint(port),
					User:         user,
					Password:     os.Getenv("MOCKSERVER_MYSQL_TEST_PASSWORD"),
					DatabaseName: databaseName,
				},
			},
		},
	}, true
}

func cleanupMySQLRuleset(t *testing.T, ctx context.Context, manager dbspi.Manager, id string) {
	t.Helper()
	tableStore := dbhelper.NewTableStore(&modeldo.RuleSetDraft{}, dbhelper.WithManager(manager))
	sqlStore, ok := dbhelper.AsSQLTableStore(tableStore)
	if !ok {
		t.Fatalf("table store does not support raw SQL")
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_published_rule_set_tab WHERE ruleset_id IN (SELECT id FROM mockserver_rule_set_draft_tab WHERE ruleset_code = ?)", id); err != nil {
		t.Fatalf("cleanup current published %s: %v", id, err)
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_published_snapshot_tab WHERE ruleset_id IN (SELECT id FROM mockserver_rule_set_draft_tab WHERE ruleset_code = ?)", id); err != nil {
		t.Fatalf("cleanup published snapshots %s: %v", id, err)
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_rule_set_draft_tab WHERE ruleset_code = ?", id); err != nil {
		t.Fatalf("cleanup draft %s: %v", id, err)
	}
	if err := sqlStore.Exec(ctx, "DELETE FROM mockserver_namespace_tab WHERE namespace_name = ?", id); err != nil {
		t.Fatalf("cleanup namespace %s: %v", id, err)
	}
}
