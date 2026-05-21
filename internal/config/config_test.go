package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

func TestLoadReadsServerYAMLAndEnvOverrides(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "server.yml")
	content := `server:
  address: ":19090"
db:
  database_groups:
    default:
      host: "127.0.0.1"
      port: 3306
      user: "user"
      password: "pass"
      database_name: "mockserver_db"
      max_open_conns: 30
      max_idle_conns: 7
      conn_max_lifetime_seconds: 1800
      debug: false
auth:
  jwt_secret: "jwt-secret"
  debug_login_enabled: false
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	t.Setenv("MOCKSERVER_CONFIG_FILE", configPath)
	t.Setenv("MOCKSERVER_ADDR", ":18080")
	t.Setenv("MOCKSERVER_DB_HOST", "mysql.local")
	t.Setenv("MOCKSERVER_DB_PORT", "3307")
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
	defaultGroup := cfg.DB.DatabaseGroups[dbspi.DefaultDatabaseGroupKey]
	if defaultGroup.Host != "mysql.local" || defaultGroup.Port != 3307 {
		t.Fatalf("unexpected db host/port: %#v", defaultGroup)
	}
	if defaultGroup.User != "user" || defaultGroup.Password != "pass" || defaultGroup.DatabaseName != "mockserver_db" {
		t.Fatalf("unexpected db identity config: %#v", defaultGroup)
	}
	if !defaultGroup.Debug {
		t.Fatalf("expected db debug env override to be true")
	}
	if defaultGroup.MaxOpenConns != 30 || defaultGroup.MaxIdleConns != 7 || defaultGroup.ConnMaxLifetimeSeconds != 1800 {
		t.Fatalf("unexpected db pool config: %#v", cfg)
	}
	if cfg.AuthJWTSecret != "jwt-secret" || cfg.AuthDebugLoginEnabled {
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
