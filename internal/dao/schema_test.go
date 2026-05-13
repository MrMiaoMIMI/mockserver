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
		"mockserver_traffic_event_tab",
		"mockserver_traffic_event_index_tab",
		"id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
		"ENGINE=InnoDB",
		"COLLATE=utf8mb4_unicode_ci",
		"ctime BIGINT UNSIGNED",
		"mtime BIGINT UNSIGNED",
		"publish_time BIGINT UNSIGNED",
		"deleted TINYINT UNSIGNED",
		"ruleset_code VARCHAR(96)",
		"snapshot_code VARCHAR(40)",
		"namespace_code VARCHAR(64)",
		"event_code VARCHAR(48)",
		"UNIQUE KEY idx_ruleset_code",
		"UNIQUE KEY idx_snapshot_code",
		"UNIQUE KEY idx_namespace_code",
		"UNIQUE KEY idx_event_code",
		"KEY idx_mtime",
		"KEY idx_ruleset_id_publish_time",
		"KEY idx_namespace_id_protocol_event_time",
		"KEY idx_ruleset_id_rule_code_event_time",
		"event_time BIGINT UNSIGNED",
		"expire_time BIGINT UNSIGNED",
		"field_value_text TEXT NOT NULL",
		"field_value_hash BIGINT UNSIGNED",
		"KEY idx_fallback_event_time",
		"KEY idx_protocol_path_hash_time",
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
		"status_code",
		"field_value VARCHAR",
		"ruleset_id VARCHAR",
		"namespace_id VARCHAR",
		"snapshot_id VARCHAR",
		"event_id VARCHAR",
		"rule_id VARCHAR",
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
