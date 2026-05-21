package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsServerYAMLAndEnvOverrides(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "server.yml")
	content := `server:
  address: ":19090"
bootstrap:
  ruleset_file: "./examples"
db:
  driver: mysql
  dsn: "user:pass@tcp(127.0.0.1:3306)/mockserver_db?charset=utf8mb4&parseTime=True&loc=Local"
  init_schema: false
  max_open_conns: 30
  max_idle_conns: 7
  conn_max_lifetime_seconds: 1800
  debug: false
admin:
  token: "all"
  read_token: "read"
  write_token: "write"
  publish_token: "publish"
auth:
  jwt_secret: "jwt-secret"
  debug_login_enabled: false
  token_ttl_seconds: 3600
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	t.Setenv("MOCKSERVER_CONFIG_FILE", configPath)
	t.Setenv("MOCKSERVER_ADDR", ":18080")
	t.Setenv("MOCKSERVER_DB_DEBUG", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ConfigFile != configPath {
		t.Fatalf("unexpected config file: %s", cfg.ConfigFile)
	}
	if cfg.Address != ":18080" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}
	if cfg.RuleSetFile != "./examples" {
		t.Fatalf("unexpected ruleset file: %s", cfg.RuleSetFile)
	}
	if cfg.DBInitSchema {
		t.Fatalf("expected db init schema to be false")
	}
	if !cfg.DBDebug {
		t.Fatalf("expected db debug env override to be true")
	}
	if cfg.DBMaxOpenConns != 30 || cfg.DBMaxIdleConns != 7 || cfg.DBConnMaxLifetimeSeconds != 1800 {
		t.Fatalf("unexpected db pool config: %#v", cfg)
	}
	if cfg.AdminToken != "all" || cfg.AdminReadToken != "read" || cfg.AdminWriteToken != "write" || cfg.AdminPublishToken != "publish" {
		t.Fatalf("unexpected admin tokens: %#v", cfg)
	}
	if cfg.AuthJWTSecret != "jwt-secret" || cfg.AuthDebugLoginEnabled || cfg.AuthTokenTTLSeconds != 3600 {
		t.Fatalf("unexpected auth config: %#v", cfg)
	}
}

func TestLoadReturnsErrorForExplicitMissingConfigFile(t *testing.T) {
	t.Setenv("MOCKSERVER_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.yml"))

	if _, err := Load(); err == nil {
		t.Fatalf("expected missing explicit config file error")
	}
}

func TestLoadRejectsInvalidEnvValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "server.yml")
	if err := os.WriteFile(configPath, []byte("server:\n  address: ':8081'\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	t.Setenv("MOCKSERVER_CONFIG_FILE", configPath)
	t.Setenv("MOCKSERVER_DB_MAX_OPEN_CONNS", "bad-int")

	if _, err := Load(); err == nil {
		t.Fatalf("expected invalid env value error")
	}
}
