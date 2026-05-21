package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateTokenContainsOnlyEmailUserClaim(t *testing.T) {
	cfg := Config{JWTSecret: "test-secret", TokenTTLSeconds: 3600}
	token, err := GenerateToken(cfg, " user@example.com ")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected jwt segment count: %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("decode claims: %v", err)
	}
	if claims["email"] != "user@example.com" {
		t.Fatalf("email claim = %v", claims["email"])
	}
	if _, ok := claims["name"]; ok {
		t.Fatalf("unexpected name claim: %v", claims["name"])
	}
}
