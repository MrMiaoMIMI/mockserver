package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/gin-gonic/gin"

	authlib "github.com/MrMiaoMIMI/mockserver/internal/auth"
)

func TestTraceMiddlewareSetsLoggerTraceID(t *testing.T) {
	var observedTraceID string
	engine := gin.New()
	engine.Use(traceMiddleware())
	engine.GET("/mockserver/runtime/default/http/api", func(c *gin.Context) {
		observedTraceID = logger.GetTraceID(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/runtime/default/http/api", nil)
	req.Header.Set("X-Trace-ID", "trace-router-test")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if observedTraceID != "trace-router-test" {
		t.Fatalf("unexpected context trace id: %q", observedTraceID)
	}
	if got := recorder.Header().Get("X-Trace-ID"); got != "trace-router-test" {
		t.Fatalf("unexpected response trace id: %q", got)
	}
}

func TestOperatorMiddlewareSetsDBOperator(t *testing.T) {
	var observedOperator string
	engine := gin.New()
	engine.Use(operatorMiddleware())
	engine.POST("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		observedOperator, _ = dbspi.OperatorFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("X-Mockserver-Operator", " admin@example.com ")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if observedOperator != "admin@example.com" {
		t.Fatalf("unexpected context operator: %q", observedOperator)
	}
}

func TestJWTAuthMiddlewareSetsEmailAsDefaultOperator(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	token, err := authlib.GenerateToken(config.JWT, "jwt-user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var observedEmail string
	var observedOperator string
	engine := gin.New()
	engine.Use(operatorMiddleware(), jwtAuthMiddleware(config))
	engine.GET("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		email, _ := c.Get(authlib.UserEmailContextKey)
		observedEmail, _ = email.(string)
		observedOperator, _ = dbspi.OperatorFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if observedEmail != "jwt-user@example.com" {
		t.Fatalf("unexpected user email: %q", observedEmail)
	}
	if observedOperator != "jwt-user@example.com" {
		t.Fatalf("unexpected context operator: %q", observedOperator)
	}
}

func TestJWTAuthMiddlewarePreservesExplicitOperatorHeader(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	token, err := authlib.GenerateToken(config.JWT, "jwt-user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var observedOperator string
	engine := gin.New()
	engine.Use(operatorMiddleware(), jwtAuthMiddleware(config))
	engine.POST("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		observedOperator, _ = dbspi.OperatorFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Mockserver-Operator", "manual@example.com")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if observedOperator != "manual@example.com" {
		t.Fatalf("unexpected context operator: %q", observedOperator)
	}
}

func TestJWTAuthMiddlewareRejectsMissingToken(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	engine := gin.New()
	engine.Use(jwtAuthMiddleware(config))
	engine.GET("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["message"] != "jwt token is required" {
		t.Fatalf("unexpected response body: %v", body)
	}
}

func TestJWTAuthMiddlewareRejectsLegacyAdminTokenHeader(t *testing.T) {
	config := AuthConfig{JWT: authlib.Config{JWTSecret: "test-secret", DebugLoginEnabled: true}}
	engine := gin.New()
	engine.Use(jwtAuthMiddleware(config))
	engine.GET("/mockserver/api/v1/admin/rulesets", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/mockserver/api/v1/admin/rulesets", nil)
	req.Header.Set("X-Mockserver-Admin-Token", "legacy-token")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestRecoveryMiddlewareCapturesPanic(t *testing.T) {
	engine := gin.New()
	engine.Use(traceMiddleware(), accessLogMiddleware(), recoveryMiddleware())
	engine.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Trace-ID", "trace-panic-test")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if got := recorder.Header().Get("X-Trace-ID"); got != "trace-panic-test" {
		t.Fatalf("unexpected response trace id: %q", got)
	}
}
