package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/gin-gonic/gin"
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
