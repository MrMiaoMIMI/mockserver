package service

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

func TestLoadRuleSetFilesPublishesRuleSet(t *testing.T) {
	tempDir := t.TempDir()
	ruleSetPath := filepath.Join(tempDir, "ruleset.json")

	content := `{
  "id": "http-default",
  "name": "http default",
  "enabled": true,
  "protocol": "http",
  "namespace": "default",
  "selector": {
    "all": [
      {
        "field": "request.host",
        "op": "eq",
        "value": "demo.com"
      },
      {
        "field": "request.path",
        "op": "prefix",
        "value": "/api/"
      }
    ]
  },
  "rules": [
    {
      "id": "debug-api",
      "name": "Debug API response",
      "enabled": true,
      "priority": 100,
      "when": {
        "all": [
          {
            "field": "request.method",
            "op": "eq",
            "value": "GET"
          },
          {
            "field": "request.path",
            "op": "eq",
            "value": "/api/v1/debug"
          }
        ]
      },
      "action": {
        "type": "respond",
        "renderer": "static",
        "response": {
          "payload": {
            "status": 200,
            "body": {
              "message": "ok"
            }
          }
        }
      }
    }
  ]
}`
	if err := os.WriteFile(ruleSetPath, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	ruleSetService := NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)

	if err := LoadRuleSetFiles(context.Background(), ruleSetService, ruleSetPath); err != nil {
		t.Fatalf("LoadRuleSetFiles() error = %v", err)
	}

	result, err := runtimeService.MatchPublished(context.Background(), bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"host":   "demo.com",
			"path":   "/api/v1/debug",
		},
	})
	if err != nil {
		t.Fatalf("MatchPublished() error = %v", err)
	}
	if !result.Matched {
		t.Fatalf("expected published ruleset to match")
	}
	if result.Trace.RuleID != "debug-api" {
		t.Fatalf("unexpected rule id: %s", result.Trace.RuleID)
	}
}

func TestLoadRuleSetFilesSupportsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	mustWriteRuleSetFile(t, filepath.Join(tempDir, "a-ruleset.json"), "dir-rule-a", "/api/v1/a", 1)
	mustWriteRuleSetFile(t, filepath.Join(tempDir, "b-ruleset.json"), "dir-rule-b", "/api/v1/b", 2)
	if err := os.WriteFile(filepath.Join(tempDir, "ignore.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	ruleSetService := NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)

	if err := LoadRuleSetFiles(context.Background(), ruleSetService, tempDir); err != nil {
		t.Fatalf("LoadRuleSetFiles() error = %v", err)
	}

	assertPublishedRuleMatch(t, runtimeService, "/api/v1/a", "dir-rule-a")
	assertPublishedRuleMatch(t, runtimeService, "/api/v1/b", "dir-rule-b")
}

func TestLoadRuleSetFilesSupportsMultiplePaths(t *testing.T) {
	tempDir := t.TempDir()
	fileA := filepath.Join(tempDir, "a.json")
	fileB := filepath.Join(tempDir, "b.json")
	mustWriteRuleSetFile(t, fileA, "multi-rule-a", "/api/v1/a", 1)
	mustWriteRuleSetFile(t, fileB, "multi-rule-b", "/api/v1/b", 2)

	ruleSetRepository := newTestRuleSetRepository()
	namespaceService := NewNamespaceService(ruleSetRepository)
	ruleSetService := NewRuleSetService(ruleSetRepository, namespaceService)
	runtimeService := NewRuntimeService(ruleSetRepository, ruleSetRepository, namespaceService)

	if err := LoadRuleSetFiles(context.Background(), ruleSetService, fileA+","+fileB); err != nil {
		t.Fatalf("LoadRuleSetFiles() error = %v", err)
	}

	assertPublishedRuleMatch(t, runtimeService, "/api/v1/a", "multi-rule-a")
	assertPublishedRuleMatch(t, runtimeService, "/api/v1/b", "multi-rule-b")
}

func mustWriteRuleSetFile(t *testing.T, path string, ruleID string, requestPath string, version int) {
	t.Helper()

	content := `{
  "id": "` + ruleID + `-ruleset",
  "name": "` + ruleID + ` ruleset",
  "enabled": true,
  "protocol": "http",
  "namespace": "default",
  "selector": {
    "all": [
      {
        "field": "request.host",
        "op": "eq",
        "value": "demo.com"
      },
      {
        "field": "request.path",
        "op": "prefix",
        "value": "` + requestPath + `"
      }
    ]
  },
  "rules": [
    {
      "id": "` + ruleID + `",
      "name": "` + ruleID + ` response",
      "enabled": true,
      "priority": 100,
      "when": {
        "all": [
          {
            "field": "request.method",
            "op": "eq",
            "value": "GET"
          },
          {
            "field": "request.path",
            "op": "eq",
            "value": "` + requestPath + `"
          }
        ]
      },
      "action": {
        "type": "respond",
        "renderer": "static",
        "response": {
          "payload": {
            "status": 200,
            "body": {
              "version": ` + strconv.Itoa(version) + `
            }
          }
        }
      }
    }
  ]
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}

func assertPublishedRuleMatch(t *testing.T, runtimeService RuntimeService, path string, wantRuleID string) {
	t.Helper()

	result, err := runtimeService.MatchPublished(context.Background(), bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"host":   "demo.com",
			"path":   path,
		},
	})
	if err != nil {
		t.Fatalf("MatchPublished() error = %v", err)
	}
	if !result.Matched {
		t.Fatalf("expected published ruleset to match path %s", path)
	}
	if result.Trace.RuleID != wantRuleID {
		t.Fatalf("unexpected rule id for path %s: got=%s want=%s", path, result.Trace.RuleID, wantRuleID)
	}
}
