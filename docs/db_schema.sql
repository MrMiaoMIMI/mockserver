-- Mockserver MySQL schema.
-- The runtime bootstrap executes the table DDL against the database selected by
-- db.dsn. For manual setup, create and use the guide-compliant database first.

-- Open comments if needed to rebuild database
-- DROP DATABASE mockserver_db;

CREATE DATABASE IF NOT EXISTS mockserver_db
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE mockserver_db;

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

CREATE TABLE IF NOT EXISTS mockserver_traffic_event_tab (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Auto increment primary key',
    event_id VARCHAR(128) NOT NULL COMMENT 'Traffic event business id',
    trace_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'SDK trace id',
    traffic_source VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Traffic source',
    protocol_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Protocol name',
    namespace_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Namespace business id',
    operation_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Protocol operation name',
    outcome VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Decision outcome',
    decision_kind VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Decision action kind',
    ruleset_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Matched ruleset business id',
    rule_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Matched rule business id',
    snapshot_id VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'Published snapshot business id',
    fallback_reason VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Fallback reason',
    duration_ms INT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Decision latency in milliseconds',
    event_time BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Event time in UNIX seconds',
    expire_time BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Expire time in UNIX seconds',
    event_json LONGTEXT NOT NULL COMMENT 'Serialized normalized protocol event JSON',
    decision_json LONGTEXT NOT NULL COMMENT 'Serialized decision JSON',
    explain_json LONGTEXT NOT NULL COMMENT 'Serialized explain JSON',
    error_message TEXT NOT NULL COMMENT 'Decision error message',
    creator VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Creator',
    updater VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Updater',
    ctime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Create time in UNIX milliseconds',
    mtime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Update time in UNIX milliseconds',
    deleted TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Soft delete flag',
    PRIMARY KEY (id),
    UNIQUE KEY idx_event_id (event_id),
    KEY idx_ns_protocol_event_time (namespace_id, protocol_name, event_time),
    KEY idx_trace_id (trace_id),
    KEY idx_outcome_event_time (outcome, event_time),
    KEY idx_ruleset_rule_event_time (ruleset_id, rule_id, event_time),
    KEY idx_fallback_event_time (fallback_reason, event_time),
    KEY idx_expire_time (expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mockserver SDK traffic event table';

CREATE TABLE IF NOT EXISTS mockserver_traffic_event_index_tab (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Auto increment primary key',
    traffic_event_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Traffic event primary key',
    event_id VARCHAR(128) NOT NULL COMMENT 'Traffic event business id',
    protocol_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Protocol name',
    field_path VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Protocol field path',
    field_value_preview VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Display preview for field value',
    field_value_hash BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Unsigned 64-bit hash for equality lookup',
    field_value_text TEXT NOT NULL COMMENT 'Full field value',
    event_time BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Event time in UNIX seconds',
    expire_time BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Expire time in UNIX seconds',
    creator VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Creator',
    updater VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Updater',
    ctime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Create time in UNIX milliseconds',
    mtime BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Update time in UNIX milliseconds',
    deleted TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Soft delete flag',
    PRIMARY KEY (id),
    KEY idx_event_id (event_id),
    KEY idx_traffic_event_id (traffic_event_id),
    KEY idx_protocol_path_hash_time (protocol_name, field_path, field_value_hash, event_time),
    KEY idx_expire_time (expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mockserver SDK traffic event index table';
