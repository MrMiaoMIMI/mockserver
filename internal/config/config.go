package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

const DefaultConfigFile = "etc/server.yml"

type Config struct {
	ConfigFile               string
	Address                  string
	RuleSetFile              string
	DBDriver                 string
	DBDSN                    string
	DBInitSchema             bool
	DBMaxOpenConns           int
	DBMaxIdleConns           int
	DBConnMaxLifetimeSeconds int
	DBDebug                  bool
	AdminToken               string
	AdminReadToken           string
	AdminWriteToken          string
	AdminPublishToken        string
}

type fileConfig struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
	Bootstrap struct {
		RuleSetFile string `yaml:"ruleset_file"`
	} `yaml:"bootstrap"`
	DB struct {
		Driver                 string `yaml:"driver"`
		DSN                    string `yaml:"dsn"`
		InitSchema             *bool  `yaml:"init_schema"`
		MaxOpenConns           int    `yaml:"max_open_conns"`
		MaxIdleConns           int    `yaml:"max_idle_conns"`
		ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime_seconds"`
		Debug                  *bool  `yaml:"debug"`
	} `yaml:"db"`
	Admin struct {
		Token        string `yaml:"token"`
		ReadToken    string `yaml:"read_token"`
		WriteToken   string `yaml:"write_token"`
		PublishToken string `yaml:"publish_token"`
	} `yaml:"admin"`
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
	return cfg, nil
}

func Default() Config {
	return Config{
		Address:      ":8080",
		DBDriver:     "mysql",
		DBInitSchema: true,
	}
}

func applyFileConfig(cfg *Config, fc fileConfig) {
	setString(&cfg.Address, fc.Server.Address)
	setString(&cfg.RuleSetFile, fc.Bootstrap.RuleSetFile)
	setString(&cfg.DBDriver, fc.DB.Driver)
	setString(&cfg.DBDSN, fc.DB.DSN)
	if fc.DB.InitSchema != nil {
		cfg.DBInitSchema = *fc.DB.InitSchema
	}
	if fc.DB.MaxOpenConns > 0 {
		cfg.DBMaxOpenConns = fc.DB.MaxOpenConns
	}
	if fc.DB.MaxIdleConns > 0 {
		cfg.DBMaxIdleConns = fc.DB.MaxIdleConns
	}
	if fc.DB.ConnMaxLifetimeSeconds > 0 {
		cfg.DBConnMaxLifetimeSeconds = fc.DB.ConnMaxLifetimeSeconds
	}
	if fc.DB.Debug != nil {
		cfg.DBDebug = *fc.DB.Debug
	}
	setString(&cfg.AdminToken, fc.Admin.Token)
	setString(&cfg.AdminReadToken, fc.Admin.ReadToken)
	setString(&cfg.AdminWriteToken, fc.Admin.WriteToken)
	setString(&cfg.AdminPublishToken, fc.Admin.PublishToken)
}

func applyEnvOverrides(cfg *Config) error {
	setStringFromEnv(&cfg.Address, "MOCKSERVER_ADDR")
	setStringFromEnv(&cfg.RuleSetFile, "MOCKSERVER_RULESET_FILE")
	setStringFromEnv(&cfg.DBDriver, "MOCKSERVER_DB_DRIVER")
	setStringFromEnv(&cfg.DBDSN, "MOCKSERVER_DB_DSN")
	if err := setBoolFromEnv(&cfg.DBInitSchema, "MOCKSERVER_DB_INIT_SCHEMA"); err != nil {
		return err
	}
	if err := setIntFromEnv(&cfg.DBMaxOpenConns, "MOCKSERVER_DB_MAX_OPEN_CONNS"); err != nil {
		return err
	}
	if err := setIntFromEnv(&cfg.DBMaxIdleConns, "MOCKSERVER_DB_MAX_IDLE_CONNS"); err != nil {
		return err
	}
	if err := setIntFromEnv(&cfg.DBConnMaxLifetimeSeconds, "MOCKSERVER_DB_CONN_MAX_LIFETIME_SECONDS"); err != nil {
		return err
	}
	if err := setBoolFromEnv(&cfg.DBDebug, "MOCKSERVER_DB_DEBUG"); err != nil {
		return err
	}
	setStringFromEnv(&cfg.AdminToken, "MOCKSERVER_ADMIN_TOKEN")
	setStringFromEnv(&cfg.AdminReadToken, "MOCKSERVER_ADMIN_READ_TOKEN")
	setStringFromEnv(&cfg.AdminWriteToken, "MOCKSERVER_ADMIN_WRITE_TOKEN")
	setStringFromEnv(&cfg.AdminPublishToken, "MOCKSERVER_ADMIN_PUBLISH_TOKEN")
	return nil
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
