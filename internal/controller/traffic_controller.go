package controller

import (
	"strconv"
	"strings"

	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
)

type TrafficController struct {
	service service.TrafficService
}

func NewTrafficController(trafficService service.TrafficService) *TrafficController {
	return &TrafficController{service: trafficService}
}

func (c *TrafficController) ListEvents(ctx *gin.Context) {
	query := bo.TrafficQuery{
		Limit:          intQuery(ctx, "limit", 50),
		Offset:         intQuery(ctx, "offset", 0),
		StartTime:      uint64Query(ctx, "start_time", 0),
		EndTime:        uint64Query(ctx, "end_time", 0),
		TraceID:        strings.TrimSpace(ctx.Query("trace_id")),
		ProtocolName:   strings.TrimSpace(ctx.Query("protocol_name")),
		NamespaceID:    strings.TrimSpace(ctx.Query("namespace_id")),
		OperationName:  strings.TrimSpace(ctx.Query("operation_name")),
		Outcome:        strings.TrimSpace(ctx.Query("outcome")),
		DecisionKind:   strings.TrimSpace(ctx.Query("decision_kind")),
		RuleSetID:      strings.TrimSpace(ctx.Query("ruleset_id")),
		RuleID:         strings.TrimSpace(ctx.Query("rule_id")),
		FallbackReason: strings.TrimSpace(ctx.Query("fallback_reason")),
		IndexFilters:   trafficIndexFilters(ctx),
		IncludeIndexes: boolQuery(ctx, "include_indexes", false),
	}
	list, err := c.service.ListTrafficEvents(ctx.Request.Context(), query)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewListTrafficEventsResponse(list))
}

func trafficIndexFilters(ctx *gin.Context) []bo.TrafficIndexFilter {
	paths := ctx.QueryArray("field_path")
	values := ctx.QueryArray("field_value")
	if len(paths) == 0 {
		path := strings.TrimSpace(ctx.Query("field_path"))
		value := strings.TrimSpace(ctx.Query("field_value"))
		if path != "" {
			return []bo.TrafficIndexFilter{{FieldPath: path, FieldValue: value}}
		}
	}
	count := len(paths)
	if len(values) < count {
		count = len(values)
	}
	filters := make([]bo.TrafficIndexFilter, 0, count)
	for i := 0; i < count; i++ {
		path := strings.TrimSpace(paths[i])
		if path == "" {
			continue
		}
		filters = append(filters, bo.TrafficIndexFilter{
			FieldPath:  path,
			FieldValue: strings.TrimSpace(values[i]),
		})
	}
	return filters
}

func intQuery(ctx *gin.Context, key string, defaultValue int) int {
	value := strings.TrimSpace(ctx.Query(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func uint64Query(ctx *gin.Context, key string, defaultValue uint64) uint64 {
	value := strings.TrimSpace(ctx.Query(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func boolQuery(ctx *gin.Context, key string, defaultValue bool) bool {
	value := strings.TrimSpace(ctx.Query(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
