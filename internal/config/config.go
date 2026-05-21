package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/goccy/go-yaml"
)

const DefaultConfigFile = "etc/server.yml"

type Config struct {
	ConfigFile            string
	APIPrefix             string
	DB                    dbspi.DatabaseConfig
	AuthJWTSecret         string
	AuthDebugLoginEnabled bool
}

type fileConfig struct {
	Server struct {
		APIPrefix string `yaml:"api_prefix"`
	} `yaml:"server"`
	DB   dbspi.DatabaseConfig `yaml:"db"`
	Auth struct {
		JWTSecret         string `yaml:"jwt_secret"`
		DebugLoginEnabled *bool  `yaml:"debug_login_enabled"`
	} `yaml:"auth"`
}

func Load() (Config, error) {
	cfg := Default()
	path := strings.TrimSpace(os.Getenv("MOCKSERVER_CONFIG_FILE"))
	explicitConfigFile := path != ""
	if path == "" {
		path = DefaultConfigFile
	}
	cfg.ConfigFile = path

	content, err := os.ReadFile(path)
	if err != nil {
		if explicitConfigFile || !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("load config file %s: %w", path, err)
		}
	} else if len(strings.TrimSpace(string(content))) > 0 {
		var fc fileConfig
		if err := yaml.Unmarshal(content, &fc); err != nil {
			return Config{}, fmt.Errorf("decode config file %s: %w", path, err)
		}
		applyFileConfig(&cfg, fc)
	}

	if err := applyEnvOverrides(&cfg); err != nil {
		return Config{}, err
	}
	if err := normalize(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Default() Config {
	return Config{
		AuthDebugLoginEnabled: true,
	}
}

func applyFileConfig(cfg *Config, fc fileConfig) {
	setString(&cfg.APIPrefix, fc.Server.APIPrefix)
	if fc.DB.DatabaseGroups != nil {
		cfg.DB = fc.DB
	}
	setString(&cfg.AuthJWTSecret, fc.Auth.JWTSecret)
	if fc.Auth.DebugLoginEnabled != nil {
		cfg.AuthDebugLoginEnabled = *fc.Auth.DebugLoginEnabled
	}
}

func applyEnvOverrides(cfg *Config) error {
	setStringFromEnv(&cfg.APIPrefix, "MOCKSERVER_API_PREFIX")
	setDefaultDatabaseGroupStringFromEnv(cfg, "MOCKSERVER_DB_HOST", func(group *dbspi.DatabaseGroupConfig, value string) {
		group.Host = value
	})
	setDefaultDatabaseGroupStringFromEnv(cfg, "MOCKSERVER_DB_USER", func(group *dbspi.DatabaseGroupConfig, value string) {
		group.User = value
	})
	setDefaultDatabaseGroupStringFromEnv(cfg, "MOCKSERVER_DB_PASSWORD", func(group *dbspi.DatabaseGroupConfig, value string) {
		group.Password = value
	})
	setDefaultDatabaseGroupStringFromEnv(cfg, "MOCKSERVER_DB_DATABASE_NAME", func(group *dbspi.DatabaseGroupConfig, value string) {
		group.DatabaseName = value
	})
	if err := setDefaultDatabaseGroupUintFromEnv(cfg, "MOCKSERVER_DB_PORT", func(group *dbspi.DatabaseGroupConfig, value uint) {
		group.Port = value
	}); err != nil {
		return err
	}
	if err := setDefaultDatabaseGroupIntFromEnv(cfg, "MOCKSERVER_DB_MAX_OPEN_CONNS", func(group *dbspi.DatabaseGroupConfig, value int) {
		group.MaxOpenConns = value
	}); err != nil {
		return err
	}
	if err := setDefaultDatabaseGroupIntFromEnv(cfg, "MOCKSERVER_DB_MAX_IDLE_CONNS", func(group *dbspi.DatabaseGroupConfig, value int) {
		group.MaxIdleConns = value
	}); err != nil {
		return err
	}
	if err := setDefaultDatabaseGroupIntFromEnv(cfg, "MOCKSERVER_DB_CONN_MAX_LIFETIME_SECONDS", func(group *dbspi.DatabaseGroupConfig, value int) {
		group.ConnMaxLifetimeSeconds = value
	}); err != nil {
		return err
	}
	if err := setDefaultDatabaseGroupBoolFromEnv(cfg, "MOCKSERVER_DB_DEBUG", func(group *dbspi.DatabaseGroupConfig, value bool) {
		group.Debug = value
	}); err != nil {
		return err
	}
	setStringFromEnv(&cfg.AuthJWTSecret, "MOCKSERVER_AUTH_JWT_SECRET")
	if err := setBoolFromEnv(&cfg.AuthDebugLoginEnabled, "MOCKSERVER_AUTH_DEBUG_LOGIN_ENABLED"); err != nil {
		return err
	}
	return nil
}

func normalize(cfg *Config) error {
	apiPrefix, err := NormalizeAPIPrefix(cfg.APIPrefix)
	if err != nil {
		return err
	}
	cfg.APIPrefix = apiPrefix
	return nil
}

func NormalizeAPIPrefix(value string) (string, error) {
	prefix := strings.TrimSpace(value)
	if prefix == "" || prefix == "/" {
		return "", nil
	}
	if strings.ContainsAny(prefix, " \t\r\n?#:*") {
		return "", fmt.Errorf("invalid api prefix %q: must be a literal URL path prefix", value)
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return strings.TrimRight(prefix, "/"), nil
}

func setString(dst *string, value string) {
	if strings.TrimSpace(value) != "" {
		*dst = value
	}
}

func setStringFromEnv(dst *string, key string) {
	if value, ok := os.LookupEnv(key); ok {
		*dst = value
	}
}

func setDefaultDatabaseGroupStringFromEnv(cfg *Config, key string, apply func(*dbspi.DatabaseGroupConfig, string)) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return
	}
	updateDefaultDatabaseGroup(cfg, func(group *dbspi.DatabaseGroupConfig) {
		apply(group, value)
	})
}

func setDefaultDatabaseGroupUintFromEnv(cfg *Config, key string, apply func(*dbspi.DatabaseGroupConfig, uint)) error {
	value, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	updateDefaultDatabaseGroup(cfg, func(group *dbspi.DatabaseGroupConfig) {
		apply(group, uint(parsed))
	})
	return nil
}

func setDefaultDatabaseGroupBoolFromEnv(cfg *Config, key string, apply func(*dbspi.DatabaseGroupConfig, bool)) error {
	value, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	updateDefaultDatabaseGroup(cfg, func(group *dbspi.DatabaseGroupConfig) {
		apply(group, parsed)
	})
	return nil
}

func setDefaultDatabaseGroupIntFromEnv(cfg *Config, key string, apply func(*dbspi.DatabaseGroupConfig, int)) error {
	value, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	updateDefaultDatabaseGroup(cfg, func(group *dbspi.DatabaseGroupConfig) {
		apply(group, parsed)
	})
	return nil
}

func updateDefaultDatabaseGroup(cfg *Config, update func(*dbspi.DatabaseGroupConfig)) {
	if cfg.DB.DatabaseGroups == nil {
		cfg.DB.DatabaseGroups = map[string]dbspi.DatabaseGroupConfig{}
	}
	group := cfg.DB.DatabaseGroups[dbspi.DefaultDatabaseGroupKey]
	update(&group)
	cfg.DB.DatabaseGroups[dbspi.DefaultDatabaseGroupKey] = group
}

func setBoolFromEnv(dst *bool, key string) error {
	value, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	*dst = parsed
	return nil
}

func setIntFromEnv(dst *int, key string) error {
	value, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	*dst = parsed
	return nil
}
