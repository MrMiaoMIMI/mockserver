package server

import "testing"

func TestListenAddressFromEnvDefaultsTo8080(t *testing.T) {
	t.Setenv("PORT", "")

	address, err := listenAddressFromEnv()
	if err != nil {
		t.Fatalf("listenAddressFromEnv() error = %v", err)
	}
	if address != ":8080" {
		t.Fatalf("unexpected listen address: %s", address)
	}
}

func TestListenAddressFromEnvUsesPortEnv(t *testing.T) {
	t.Setenv("PORT", "18080")

	address, err := listenAddressFromEnv()
	if err != nil {
		t.Fatalf("listenAddressFromEnv() error = %v", err)
	}
	if address != ":18080" {
		t.Fatalf("unexpected listen address: %s", address)
	}
}

func TestListenAddressFromEnvRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "bad-port")

	if _, err := listenAddressFromEnv(); err == nil {
		t.Fatalf("expected invalid PORT error")
	}
}
