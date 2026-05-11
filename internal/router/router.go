package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"mockserver/internal/controller"
)

type traceKey string

const traceIDKey traceKey = "trace_id"

const (
	mockserverOperatorHeader = "X-Mockserver-Operator"
	operatorHeader           = "X-Operator"
)

func New(adminController *controller.AdminController, runtimeController *controller.RuntimeController, authConfig AdminAuthConfig, metricsController *controller.MetricsController) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.Use(traceMiddleware(), operatorMiddleware(), accessLogMiddleware(), recoveryMiddleware(), adminAuthMiddleware(authConfig))

	admin := engine.Group("/mockserver/api/v1/admin")
	admin.POST("/rulesets", adminController.CreateOrUpdateDraft)
	admin.GET("/rulesets", adminController.ListDrafts)
	admin.GET("/rulesets/:ruleset_id", adminController.GetDraft)
	admin.POST("/rulesets/:ruleset_id/rules", adminController.AddDraftRule)
	admin.PUT("/rulesets/:ruleset_id/rules/:rule_id", adminController.UpdateDraftRule)
	admin.DELETE("/rulesets/:ruleset_id/rules/:rule_id", adminController.DeleteDraftRule)
	admin.POST("/rulesets/:ruleset_id/rules/:rule_id/enable", adminController.EnableDraftRule)
	admin.POST("/rulesets/:ruleset_id/rules/:rule_id/disable", adminController.DisableDraftRule)
	admin.POST("/rulesets/:ruleset_id/rules/:rule_id/priority", adminController.SetDraftRulePriority)
	admin.POST("/rulesets/:ruleset_id/validate", adminController.ValidateDraft)
	admin.POST("/rulesets/:ruleset_id/publish", adminController.PublishDraft)
	admin.POST("/rulesets/:ruleset_id/simulate", adminController.SimulateDraft)
	admin.GET("/protocols", adminController.ListProtocols)

	admin.POST("/namespaces", adminController.CreateNamespace)
	admin.GET("/namespaces", adminController.ListNamespaces)
	admin.GET("/namespaces/:namespace_id", adminController.GetNamespace)
	admin.PUT("/namespaces/:namespace_id", adminController.UpdateNamespace)

	admin.POST("/published/simulate", adminController.SimulatePublished)
	admin.GET("/published/rulesets", adminController.ListPublished)
	admin.GET("/published/rulesets/:ruleset_id", adminController.GetPublished)
	admin.GET("/published/rulesets/:ruleset_id/snapshots", adminController.ListPublishedSnapshots)
	admin.POST("/published/rulesets/:ruleset_id/rollback/preview", adminController.RollbackPreview)
	admin.POST("/published/rulesets/:ruleset_id/rollback", adminController.Rollback)
	if metricsController != nil {
		admin.GET("/metrics/runtime", metricsController.RuntimeMetrics)
	}
	engine.POST("/mockserver/api/v1/sdk/decision", runtimeController.DecidePublished)
	engine.Any("/mockserver/runtime/:namespace/http", runtimeController.HandleHTTP)
	engine.Any("/mockserver/runtime/:namespace/http/*runtime_path", runtimeController.HandleHTTP)
	return engine
}

func traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		traceID := r.Header.Get("X-Trace-ID")
		if strings.TrimSpace(traceID) == "" {
			traceID = newTraceID()
		}
		ctx := logger.SetTraceID(r.Context(), traceID)
		ctx = withTraceID(ctx, traceID)
		r = r.WithContext(ctx)
		r.Header.Set("X-Trace-ID", traceID)
		c.Request = r
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

func operatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		operator := operatorFromHeaders(c)
		if operator != "" {
			r := c.Request
			r = r.WithContext(dbspi.WithOperator(r.Context(), operator))
			c.Request = r
		}
		c.Next()
	}
}

func operatorFromHeaders(c *gin.Context) string {
	operator := strings.TrimSpace(c.GetHeader(mockserverOperatorHeader))
	if operator == "" {
		operator = strings.TrimSpace(c.GetHeader(operatorHeader))
	}
	return operator
}

func accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		status := c.Writer.Status()
		fields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("raw_query", c.Request.URL.RawQuery),
			logger.String("route", c.FullPath()),
			logger.Int("status", status),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
			logger.Duration("latency", time.Since(startedAt)),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("errors", c.Errors.String()))
		}

		ctx := c.Request.Context()
		switch {
		case status >= http.StatusInternalServerError:
			logger.Error(ctx, "HTTP request completed", fields...)
		case status >= http.StatusBadRequest:
			logger.Warn(ctx, "HTTP request completed", fields...)
		default:
			logger.Info(ctx, "HTTP request completed", fields...)
		}
	}
}

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			err := recoveredError(recovered)
			_ = c.Error(err)
			logger.Error(c.Request.Context(), "HTTP panic recovered",
				logger.Any("panic", recovered),
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("route", c.FullPath()),
				logger.String("client_ip", c.ClientIP()),
				logger.String("stack", string(debug.Stack())),
			)
			if !c.Writer.Written() {
				serverresp.InternalServerError(c, errors.New("internal server error"))
			}
			c.Abort()
		}()
		c.Next()
	}
}

func recoveredError(recovered any) error {
	if err, ok := recovered.(error); ok {
		return err
	}
	return fmt.Errorf("%v", recovered)
}

func newTraceID() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err == nil {
		return hex.EncodeToString(data[:])
	}
	return "trace-" + time.Now().UTC().Format("20060102150405.000000000")
}

func withTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}
