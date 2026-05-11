package dao

import (
	"strings"
	"testing"
)

func TestDefaultSchemaSQLFollowsMySQLDesignGuide(t *testing.T) {
	required := []string{
		"mockserver_rule_set_draft_tab",
		"mockserver_published_snapshot_tab",
		"mockserver_published_rule_set_tab",
		"mockserver_namespace_tab",
		"id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
		"ENGINE=InnoDB",
		"COLLATE=utf8mb4_unicode_ci",
		"ctime BIGINT UNSIGNED",
		"mtime BIGINT UNSIGNED",
		"publish_time BIGINT UNSIGNED",
		"deleted TINYINT UNSIGNED",
		"UNIQUE KEY idx_ruleset_id",
		"KEY idx_mtime",
		"KEY idx_ruleset_id_publish_time",
	}
	for _, item := range required {
		if !strings.Contains(defaultSchemaSQL, item) {
			t.Fatalf("defaultSchemaSQL missing %q", item)
		}
	}

	forbidden := []string{
		"mockserver_rule_set_drafts",
		"mockserver_published_snapshots",
		"mockserver_published_rulesets",
		"mockserver_namespaces",
		" TIMESTAMP",
		" BOOLEAN",
		" DATETIME",
		" FOREIGN KEY",
		" ADD COLUMN ",
		"create_time",
		"update_time",
		"is_deleted",
	}
	upperSchema := strings.ToUpper(defaultSchemaSQL)
	for _, item := range forbidden {
		if strings.Contains(upperSchema, strings.ToUpper(item)) {
			t.Fatalf("defaultSchemaSQL contains forbidden pattern %q", item)
		}
	}
}
