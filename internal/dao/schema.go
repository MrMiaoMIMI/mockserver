package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	_ "github.com/go-sql-driver/mysql"

	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

const defaultSchemaSQL = `
CREATE TABLE IF NOT EXISTS mockserver_rule_set_draft_tab (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Auto increment primary key',
    ruleset_id VARCHAR(128) NOT NULL COMMENT 'Ruleset business id',
    version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'Draft version',
    ruleset_json LONGTEXT NOT NULL COMMENT 'Serialized ruleset JSON',
    creator VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Creator',
    updater VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Updater',
    ctime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Create time in UNIX milliseconds',
    mtime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Update time in UNIX milliseconds',
    deleted TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Soft delete flag',
    PRIMARY KEY (id),
    UNIQUE KEY idx_ruleset_id (ruleset_id),
    KEY idx_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mockserver ruleset draft table';

CREATE TABLE IF NOT EXISTS mockserver_published_snapshot_tab (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Auto increment primary key',
    snapshot_id VARCHAR(256) NOT NULL COMMENT 'Snapshot business id',
    ruleset_id VARCHAR(128) NOT NULL COMMENT 'Ruleset business id',
    ruleset_version INT UNSIGNED NOT NULL COMMENT 'Ruleset version',
    ruleset_json LONGTEXT NOT NULL COMMENT 'Serialized ruleset JSON',
    audit_json LONGTEXT NOT NULL COMMENT 'Serialized audit JSON',
    publish_time BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Publish time in UNIX milliseconds',
    creator VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Creator',
    updater VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Updater',
    ctime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Create time in UNIX milliseconds',
    mtime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Update time in UNIX milliseconds',
    deleted TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Soft delete flag',
    PRIMARY KEY (id),
    UNIQUE KEY idx_snapshot_id (snapshot_id),
    KEY idx_ruleset_id_publish_time (ruleset_id, publish_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mockserver published snapshot table';

CREATE TABLE IF NOT EXISTS mockserver_published_rule_set_tab (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Auto increment primary key',
    ruleset_id VARCHAR(128) NOT NULL COMMENT 'Ruleset business id',
    current_snapshot_id VARCHAR(256) NOT NULL COMMENT 'Current snapshot business id',
    creator VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Creator',
    updater VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Updater',
    ctime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Create time in UNIX milliseconds',
    mtime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Update time in UNIX milliseconds',
    deleted TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Soft delete flag',
    PRIMARY KEY (id),
    UNIQUE KEY idx_ruleset_id (ruleset_id),
    KEY idx_current_snapshot_id (current_snapshot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mockserver current published ruleset table';

CREATE TABLE IF NOT EXISTS mockserver_namespace_tab (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Auto increment primary key',
    namespace_id VARCHAR(128) NOT NULL COMMENT 'Namespace business id',
    version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'Namespace version',
    namespace_json LONGTEXT NOT NULL COMMENT 'Serialized namespace JSON',
    creator VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Creator',
    updater VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Updater',
    ctime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Create time in UNIX milliseconds',
    mtime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Update time in UNIX milliseconds',
    deleted TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Soft delete flag',
    PRIMARY KEY (id),
    UNIQUE KEY idx_namespace_id (namespace_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mockserver namespace table';
`

func applySchema(ctx context.Context, manager dbspi.Manager, schema string) error {
	store := dbhelper.NewTableStore(&modeldo.RuleSetDraft{}, dbhelper.WithManager(manager))
	sqlStore, ok := dbhelper.AsSQLTableStore(store)
	if !ok {
		return fmt.Errorf("rule set draft table store does not support raw SQL")
	}
	for _, statement := range strings.Split(schema, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if err := sqlStore.Exec(ctx, statement); err != nil && !isIgnorableSchemaError(err) {
			return fmt.Errorf("apply schema statement %q: %w", statement, err)
		}
	}
	return nil
}

func isIgnorableSchemaError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "Error 1060") ||
		strings.Contains(message, "Duplicate column name")
}
