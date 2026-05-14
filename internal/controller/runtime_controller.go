package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/adapter/httpadapter"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/request"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
	"github.com/MrMiaoMIMI/mockserver/internal/observability"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
	"github.com/MrMiaoMIMI/mockserver/internal/view"
)

type RuntimeController struct {
	view           view.RuntimeView
	runtimeMetrics *observability.RuntimeMetrics
	trafficService service.TrafficService
}

func NewRuntimeController(runtimeView view.RuntimeView, runtimeMetrics *observability.RuntimeMetrics) *RuntimeController {
	return &RuntimeController{
		view:           runtimeView,
		runtimeMetrics: runtimeMetrics,
	}
}

func NewRuntimeControllerWithTraffic(runtimeView view.RuntimeView, runtimeMetrics *observability.RuntimeMetrics, trafficService service.TrafficService) *RuntimeController {
	controller := NewRuntimeController(runtimeView, runtimeMetrics)
	controller.trafficService = trafficService
	return controller
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
	startedAt := time.Now()
	var req request.DecidePublishedRequest
	if !bindJSON(ctx, &req) {
		return
	}
	decision, err := c.view.DecidePublished(ctx.Request.Context(), req.Event)
	c.observeSuppressedSDKRulesetMiss(ctx, req.Event, decision, err, startedAt)
	c.recordSDKDecision(ctx, req.Event, decision, err, startedAt)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.DecidePublishedResponse{Decision: decision})
}

func (c *RuntimeController) recordSDKDecision(ctx *gin.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, startedAt time.Time) {
	if c.trafficService == nil {
		return
	}
	durationMS := time.Since(startedAt).Milliseconds()
	if durationMS < 0 {
		durationMS = 0
	}
	if durationMS > int64(^uint32(0)) {
		durationMS = int64(^uint32(0))
	}
	if _, err := c.trafficService.RecordSDKDecision(ctx.Request.Context(), event, decision, decisionErr, uint32(durationMS)); err != nil {
		logger.Error(ctx.Request.Context(), "Record SDK traffic event failed",
			logger.String("event", "sdk_traffic_record_failed"),
			logger.String("protocol", event.Protocol),
			logger.String("namespace", event.Namespace),
			logger.String("message", err.Error()),
		)
	}
}

func (c *RuntimeController) observeSuppressedSDKRulesetMiss(ctx *gin.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, startedAt time.Time) {
	if c.runtimeMetrics == nil || decisionErr != nil || strings.TrimSpace(decision.Trace.FallbackReason) != eo.FallbackReasonRulesetMiss {
		return
	}
	method := defaultString(bo.RequestString(event.Request, "method"), bo.RequestString(event.Request, "operation"))
	path := defaultString(bo.RequestString(event.Request, "path"), bo.RequestString(event.Request, "cmd"))
	path = defaultString(path, bo.RequestString(event.Request, "key"))
	c.runtimeMetrics.Observe(observability.RuntimeObservation{
		Source:         bo.TrafficSourceSDKDecision,
		Protocol:       strings.TrimSpace(event.Protocol),
		Operation:      sdkRuntimeOperation(event),
		Namespace:      event.Namespace,
		Method:         method,
		Host:           bo.RequestString(event.Request, "host"),
		Path:           path,
		TraceID:        firstNonEmpty(event.Meta.TraceID, decision.Meta.TraceID),
		Fallback:       true,
		FallbackReason: decision.Trace.FallbackReason,
		Message:        "sdk traffic ruleset_miss skipped raw persistence",
		Duration:       time.Since(startedAt),
	})
}

func (c *RuntimeController) observeRuntime(r *http.Request, event *bo.Event, observation observability.RuntimeObservation, startedAt time.Time) {
	duration := time.Since(startedAt)
	observation.Duration = duration
	if event != nil {
		observation.Protocol = defaultString(observation.Protocol, event.Protocol)
		observation.Operation = defaultString(observation.Operation, event.Operation)
		observation.Namespace = defaultString(observation.Namespace, event.Namespace)
		observation.Method = defaultString(observation.Method, bo.RequestString(event.Request, "method"))
		observation.Operation = defaultString(observation.Operation, bo.RequestString(event.Request, "operation"))
		observation.Scheme = defaultString(observation.Scheme, bo.RequestString(event.Request, "scheme"))
		observation.Host = defaultString(observation.Host, bo.RequestString(event.Request, "host"))
		observation.Path = defaultString(observation.Path, bo.RequestString(event.Request, "path"))
		observation.TraceID = defaultString(observation.TraceID, event.Meta.TraceID)
		observation.Event = event
	}
	observation.Source = defaultString(observation.Source, observability.RuntimeSourceHTTP)
	observation.Protocol = defaultString(observation.Protocol, eo.ProtocolHTTP)
	observation.Method = defaultString(observation.Method, r.Method)
	observation.Operation = defaultString(observation.Operation, observation.Method)
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func sdkRuntimeOperation(event bo.Event) string {
	if value := strings.TrimSpace(event.Operation); value != "" {
		return value
	}
	if value := bo.RequestString(event.Request, "operation"); value != "" {
		return value
	}
	if value := bo.RequestString(event.Request, "method"); value != "" {
		return strings.ToUpper(value)
	}
	return ""
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
