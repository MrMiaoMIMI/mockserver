package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"mockserver/internal/adapter/httpadapter"
	"mockserver/internal/model/bo"
	"mockserver/internal/model/request"
	"mockserver/internal/model/response"
	"mockserver/internal/observability"
	"mockserver/internal/view"
)

type RuntimeController struct {
	view           view.RuntimeView
	runtimeMetrics *observability.RuntimeMetrics
}

func NewRuntimeController(runtimeView view.RuntimeView, runtimeMetrics *observability.RuntimeMetrics) *RuntimeController {
	return &RuntimeController{
		view:           runtimeView,
		runtimeMetrics: runtimeMetrics,
	}
}

func (c *RuntimeController) HandleHTTP(ctx *gin.Context) {
	startedAt := time.Now()
	r := ctx.Request
	namespace, runtimePath, ok := extractRuntimeTargetFromGin(ctx)
	if !ok {
		c.observeRuntime(r, nil, observability.RuntimeObservation{
			Namespace: namespace,
			Error:     true,
			Status:    http.StatusNotFound,
			Message:   "route not found",
		}, startedAt)
		serverresp.NotFoundError(ctx, errNotFound)
		return
	}

	normalizedRequest := r.Clone(r.Context())
	normalizedURL := *r.URL
	normalizedURL.Path = runtimePath
	normalizedRequest.URL = &normalizedURL

	event, err := httpadapter.NormalizeHTTPRequest(normalizedRequest, namespace)
	if err != nil {
		c.observeRuntime(r, nil, observability.RuntimeObservation{
			Namespace: namespace,
			Error:     true,
			Status:    http.StatusBadRequest,
			Message:   err.Error(),
		}, startedAt)
		serverresp.BadRequestError(ctx, err)
		return
	}

	result, err := c.view.MatchPublished(r.Context(), event)
	if err != nil {
		c.observeRuntime(r, &event, observability.RuntimeObservation{
			Error:   true,
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}, startedAt)
		serverresp.InternalServerError(ctx, err)
		return
	}
	if !result.Matched && result.Response.Status == 0 {
		c.observeRuntime(r, &event, observability.RuntimeObservation{
			Status:  http.StatusNotFound,
			Message: "no mock rule matched",
		}, startedAt)
		serverresp.NotFoundError(ctx, errNotFound)
		return
	}

	if result.Trace.RulesetID != "" {
		ctx.Header("X-Mockserver-Ruleset", result.Trace.RulesetID)
	}
	if result.Trace.RuleID != "" {
		ctx.Header("X-Mockserver-Rule", result.Trace.RuleID)
	}
	if result.Fallback {
		ctx.Header("X-Mockserver-Fallback", result.Trace.FallbackReason)
	}
	for key, values := range result.Response.Headers {
		for _, value := range values {
			ctx.Writer.Header().Add(key, value)
		}
	}
	c.observeRuntime(r, &event, observability.RuntimeObservation{
		Matched:        result.Matched,
		Fallback:       result.Fallback,
		Status:         result.Response.Status,
		RulesetID:      result.Trace.RulesetID,
		RuleID:         result.Trace.RuleID,
		FallbackReason: result.Trace.FallbackReason,
	}, startedAt)

	switch body := result.Response.Body.(type) {
	case string:
		ctx.Data(result.Response.Status, "", []byte(body))
	default:
		ctx.JSON(result.Response.Status, body)
	}
}

func (c *RuntimeController) DecidePublished(ctx *gin.Context) {
	var req request.DecidePublishedRequest
	if !bindJSON(ctx, &req) {
		return
	}
	decision, err := c.view.DecidePublished(ctx.Request.Context(), req.Event)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.DecidePublishedResponse{Decision: decision})
}

func (c *RuntimeController) observeRuntime(r *http.Request, event *bo.Event, observation observability.RuntimeObservation, startedAt time.Time) {
	duration := time.Since(startedAt)
	observation.Duration = duration
	if event != nil {
		observation.Namespace = defaultString(observation.Namespace, event.Namespace)
		observation.Method = defaultString(observation.Method, bo.RequestString(event.Request, "method"))
		observation.Scheme = defaultString(observation.Scheme, bo.RequestString(event.Request, "scheme"))
		observation.Host = defaultString(observation.Host, bo.RequestString(event.Request, "host"))
		observation.Path = defaultString(observation.Path, bo.RequestString(event.Request, "path"))
		observation.TraceID = defaultString(observation.TraceID, event.Meta.TraceID)
		observation.Event = event
	}
	observation.Method = defaultString(observation.Method, r.Method)
	observation.Host = defaultString(observation.Host, r.Host)
	observation.Path = defaultString(observation.Path, r.URL.Path)
	observation.RawQuery = defaultString(observation.RawQuery, r.URL.RawQuery)
	observation.TraceID = defaultString(observation.TraceID, r.Header.Get("X-Trace-ID"))
	if c.runtimeMetrics != nil {
		c.runtimeMetrics.Observe(observation)
	}
	fields := []logger.Field{
		logger.String("event", "runtime_request"),
		logger.String("method", observation.Method),
		logger.String("path", observation.Path),
		logger.String("namespace", observation.Namespace),
		logger.Bool("matched", observation.Matched),
		logger.Bool("fallback", observation.Fallback),
		logger.Bool("error", observation.Error),
		logger.Int("status", observation.Status),
		logger.String("ruleset_id", observation.RulesetID),
		logger.String("rule_id", observation.RuleID),
		logger.String("fallback_reason", observation.FallbackReason),
		logger.Int64("duration_ms", duration.Milliseconds()),
	}
	if observation.Message != "" {
		fields = append(fields, logger.String("message", observation.Message))
	}
	if observation.Error {
		logger.Error(r.Context(), "RuntimeRequest completed", fields...)
		return
	}
	logger.Info(r.Context(), "RuntimeRequest completed", fields...)
}

func defaultString(current string, fallback string) string {
	if current != "" {
		return current
	}
	return fallback
}

func extractRuntimeTargetFromGin(ctx *gin.Context) (string, string, bool) {
	namespace := strings.TrimSpace(ctx.Param("namespace"))
	runtimePath := ctx.Param("runtime_path")
	if namespace == "" {
		return "", "", false
	}
	if runtimePath == "" {
		runtimePath = "/"
	}
	return namespace, runtimePath, true
}
