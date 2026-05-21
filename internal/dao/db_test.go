package dao

import (
	"os"
	"strconv"
	"testing"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/config"
)

func TestNewDBRequiresDatabaseConfig(t *testing.T) {
	_, err := NewDB(config.Config{})
	if err == nil {
		t.Fatalf("expected missing db config error")
	}
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
