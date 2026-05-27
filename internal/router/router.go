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

	"github.com/MrMiaoMIMI/mockserver/internal/controller"
)

type traceKey string

const traceIDKey traceKey = "trace_id"

const (
	mockserverOperatorHeader = "X-Mockserver-Operator"
	operatorHeader           = "X-Operator"
)

type RouteConfig struct {
	APIPrefix string
}

func New(adminController *controller.AdminController, runtimeController *controller.RuntimeController, authConfig AuthConfig, metricsController *controller.MetricsController) *gin.Engine {
	return NewWithTraffic(adminController, runtimeController, authConfig, metricsController, nil)
}

func NewWithTraffic(adminController *controller.AdminController, runtimeController *controller.RuntimeController, authConfig AuthConfig, metricsController *controller.MetricsController, trafficController *controller.TrafficController) *gin.Engine {
	return NewWithConfig(adminController, runtimeController, authConfig, RouteConfig{}, metricsController, trafficController)
}

func NewWithConfig(adminController *controller.AdminController, runtimeController *controller.RuntimeController, authConfig AuthConfig, routeConfig RouteConfig, metricsController *controller.MetricsController, trafficController *controller.TrafficController) *gin.Engine {
	return NewWithConfigAndAgent(adminController, runtimeController, authConfig, routeConfig, metricsController, trafficController, nil)
}

func NewWithConfigAndAgent(adminController *controller.AdminController, runtimeController *controller.RuntimeController, authConfig AuthConfig, routeConfig RouteConfig, metricsController *controller.MetricsController, trafficController *controller.TrafficController, agentController *controller.AgentController) *gin.Engine {
	return NewWithConfigAndAgentMCP(adminController, runtimeController, authConfig, routeConfig, metricsController, trafficController, agentController, nil)
}

func NewWithConfigAndAgentMCP(adminController *controller.AdminController, runtimeController *controller.RuntimeController, authConfig AuthConfig, routeConfig RouteConfig, metricsController *controller.MetricsController, trafficController *controller.TrafficController, agentController *controller.AgentController, agentMCPController *controller.AgentMCPController) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.Use(traceMiddleware(), operatorMiddleware(), accessLogMiddleware(), recoveryMiddleware())

	authController := controller.NewAuthController(authConfig.JWT)
	root := engine.Group(normalizeRoutePrefix(routeConfig.APIPrefix))
	authRoutes := root.Group("/mockserver/api/v1/auth")
	authRoutes.POST("/debug/login", authController.DebugLogin)
	authRoutes.GET("/me", jwtAuthMiddleware(authConfig), authController.CurrentUser)

	admin := root.Group("/mockserver/api/v1/admin")
	admin.Use(jwtAuthMiddleware(authConfig))
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
	if trafficController != nil {
		admin.GET("/traffic/events", trafficController.ListEvents)
		admin.GET("/traffic/events/:traffic_event_id", trafficController.GetEvent)
	}
	if agentController != nil {
		agent := root.Group("/mockserver/api/v1/agent")
		agent.Use(jwtAuthMiddleware(authConfig))
		agent.POST("/scenarios", agentController.CreateScenario)
		agent.GET("/scenarios", agentController.ListScenarios)
		agent.GET("/scenarios/:scenario_id", agentController.GetScenario)
		agent.PATCH("/scenarios/:scenario_id", agentController.UpdateScenario)
		agent.GET("/scenarios/:scenario_id/rules", agentController.ListScenarioRules)
		agent.POST("/scenarios/:scenario_id/rules/quick", agentController.UpsertHTTPQuickRule)
		agent.PUT("/scenarios/:scenario_id/rules/:rule_id", agentController.UpsertScenarioRule)
		agent.DELETE("/scenarios/:scenario_id/rules/:rule_id", agentController.DeleteScenarioRule)
		agent.POST("/scenarios/:scenario_id/simulate", agentController.SimulateScenario)
		agent.GET("/scenarios/:scenario_id/traffic", agentController.ListScenarioTraffic)
		agent.DELETE("/scenarios/:scenario_id", agentController.DeleteScenario)
		if agentMCPController != nil {
			agent.Any("/mcp", agentMCPController.Handle)
		}
	}
	root.POST("/mockserver/api/v1/sdk/decision", runtimeController.DecidePublished)
	root.Any("/mockserver/runtime/:namespace/http", runtimeController.HandleHTTP)
	root.Any("/mockserver/runtime/:namespace/http/*runtime_path", runtimeController.HandleHTTP)
	return engine
}

func normalizeRoutePrefix(prefix string) string {
	normalized := strings.TrimSpace(prefix)
	if normalized == "" || normalized == "/" {
		return ""
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return strings.TrimRight(normalized, "/")
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
