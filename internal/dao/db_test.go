package dao

import (
	"context"
	"testing"

	"mockserver/internal/config"
)

func TestNewDBRequiresDSN(t *testing.T) {
	_, err := NewDB(context.Background(), config.Config{})
	if err == nil {
		t.Fatalf("expected missing db dsn error")
	}
}

func TestNewDBRejectsUnsupportedDriver(t *testing.T) {
	_, err := NewDB(context.Background(), config.Config{
		DBDriver: "sqlite",
		DBDSN:    "sqlite.db",
	})
	if err == nil {
		t.Fatalf("expected unsupported db driver error")
	}
}
