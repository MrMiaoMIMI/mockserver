package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/internal/trafficutil"
)

const defaultTrafficRetention = 30 * 24 * time.Hour

type trafficService struct {
	repository dao.TrafficRepository
}

func NewTrafficService(repository dao.TrafficRepository) TrafficService {
	return &trafficService{repository: repository}
}

func (s *trafficService) RecordSDKDecision(ctx context.Context, event bo.Event, decision bo.RuntimeDecision, decisionErr error, durationMS uint32) (bo.TrafficEvent, error) {
	if shouldSkipRawSDKTraffic(decision, decisionErr) {
		return bo.TrafficEvent{}, nil
	}
	now := time.Now().UTC()
	eventRaw, err := marshalTrafficJSON(event)
	if err != nil {
		return bo.TrafficEvent{}, fmt.Errorf("marshal traffic event: %w", err)
	}
	decisionRaw, err := marshalTrafficJSON(decision)
	if err != nil {
		return bo.TrafficEvent{}, fmt.Errorf("marshal traffic decision: %w", err)
	}
	explainRaw, err := marshalTrafficJSON(decisionExplainDocument(decision))
	if err != nil {
		return bo.TrafficEvent{}, fmt.Errorf("marshal traffic explain: %w", err)
	}

	errorMessage := ""
	if decisionErr != nil {
		errorMessage = decisionErr.Error()
		decisionRaw, _ = marshalTrafficJSON(map[string]any{
			"error": errorMessage,
		})
		explainRaw = "{}"
	}

	traffic := bo.TrafficEvent{
		EventID:        newTrafficEventID(now),
		TraceID:        firstNonBlank(event.Meta.TraceID, decision.Meta.TraceID),
		TrafficSource:  bo.TrafficSourceSDKDecision,
		ProtocolName:   strings.TrimSpace(event.Protocol),
		NamespaceID:    strings.TrimSpace(event.Namespace),
		OperationName:  operationName(event),
		Outcome:        decisionOutcome(decision, decisionErr),
		DecisionKind:   strings.TrimSpace(decision.Kind),
		RuleSetDBID:    decision.Trace.RulesetDBID,
		RuleSetID:      strings.TrimSpace(decision.Trace.RulesetID),
		RuleID:         strings.TrimSpace(decision.Trace.RuleID),
		SnapshotDBID:   decision.Trace.SnapshotDBID,
		SnapshotID:     strings.TrimSpace(decision.Trace.SnapshotID),
		FallbackReason: strings.TrimSpace(decision.Trace.FallbackReason),
		DurationMS:     durationMS,
		EventTime:      uint64(now.Unix()),
		ExpireTime:     uint64(now.Add(defaultTrafficRetention).Unix()),
		EventJSON:      eventRaw,
		DecisionJSON:   decisionRaw,
		ExplainJSON:    explainRaw,
		ErrorMessage:   errorMessage,
	}
	traffic.Indexes = buildTrafficIndexes(event, decision)
	return s.repository.CreateTrafficEvent(ctx, traffic)
}

func shouldSkipRawSDKTraffic(decision bo.RuntimeDecision, decisionErr error) bool {
	return decisionErr == nil && decision.Fallback && strings.TrimSpace(decision.Trace.FallbackReason) == eo.FallbackReasonRulesetMiss
}

func decisionExplainDocument(decision bo.RuntimeDecision) map[string]any {
	document := map[string]any{
		"trace": decision.Trace,
	}
	if decision.Diagnostics == nil {
		return document
	}
	document["ruleset_selection"] = decision.Diagnostics.RuleSetSelection
	document["rule_selection"] = decision.Diagnostics.RuleSelection
	return document
}

func (s *trafficService) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	query.TrafficSource = bo.TrafficSourceSDKDecision
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	cleanedFilters := make([]bo.TrafficIndexFilter, 0, len(query.IndexFilters))
	for _, filter := range query.IndexFilters {
		filter.FieldPath = strings.TrimSpace(filter.FieldPath)
		if filter.FieldPath == "" {
			continue
		}
		cleanedFilters = append(cleanedFilters, filter)
	}
	query.IndexFilters = cleanedFilters
	return s.repository.ListTrafficEvents(ctx, query)
}

func (s *trafficService) GetTrafficEvent(ctx context.Context, id uint64) (bo.TrafficEvent, error) {
	if id == 0 {
		return bo.TrafficEvent{}, fmt.Errorf("traffic event %d not found", id)
	}
	event, ok, err := s.repository.GetTrafficEvent(ctx, id)
	if err != nil {
		return bo.TrafficEvent{}, err
	}
	if !ok {
		return bo.TrafficEvent{}, fmt.Errorf("traffic event %d not found", id)
	}
	return event, nil
}

func marshalTrafficJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	if string(raw) == "null" {
		return "{}", nil
	}
	return string(raw), nil
}

func newTrafficEventID(now time.Time) string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("te_%d", now.UnixNano())
	}
	return fmt.Sprintf("te_%d_%s", now.UnixNano(), hex.EncodeToString(buf[:]))
}

func decisionOutcome(decision bo.RuntimeDecision, decisionErr error) string {
	if decisionErr != nil {
		return bo.TrafficOutcomeError
	}
	if decision.Matched {
		return bo.TrafficOutcomeMatched
	}
	if decision.Fallback {
		return bo.TrafficOutcomeFallback
	}
	return bo.TrafficOutcomeUnmatched
}

func operationName(event bo.Event) string {
	if value := stringFromAny(event.Request["operation"]); value != "" {
		return value
	}
	if value := strings.TrimSpace(event.Operation); value != "" {
		return value
	}
	if value := stringFromAny(event.Request["method"]); value != "" {
		return strings.ToUpper(value)
	}
	return ""
}

func buildTrafficIndexes(event bo.Event, decision bo.RuntimeDecision) []bo.TrafficEventIndex {
	fields := map[string]string{}
	for _, key := range sortedRequestKeys(event.Request) {
		if !isQueryableRequestField(event.Protocol, key) {
			continue
		}
		if value := stringFromAny(event.Request[key]); value != "" {
			addField(fields, "event.request."+key, value)
		}
	}
	if decision.Response != nil {
		if decision.Response.Protocol != "" {
			addField(fields, "decision.response.protocol", decision.Response.Protocol)
		}
		for path, value := range responseIndexValues(*decision.Response) {
			addField(fields, "decision.response.payload."+path, value)
		}
	}

	indexes := make([]bo.TrafficEventIndex, 0, len(fields))
	paths := make([]string, 0, len(fields))
	for path := range fields {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		value := fields[path]
		indexes = append(indexes, bo.TrafficEventIndex{
			FieldPath:         path,
			FieldValuePreview: trafficutil.FieldValuePreview(value),
			FieldValueHash:    trafficutil.FieldValueHash(value),
			FieldValueText:    value,
		})
	}
	return indexes
}

func responseIndexValues(response bo.ProtocolResponse) map[string]string {
	values := map[string]string{}
	switch strings.ToLower(strings.TrimSpace(response.Protocol)) {
	case eo.ProtocolHTTP:
		if value := stringFromAny(response.Payload["status"]); value != "" {
			values["status"] = value
		}
	case eo.ProtocolSPEX:
		if value := stringFromAny(response.Payload["code"]); value != "" {
			values["code"] = value
		}
	case eo.ProtocolCache:
		if value := stringFromAny(response.Payload["hit"]); value != "" {
			values["hit"] = value
		}
	}
	return values
}

func sortedRequestKeys(request bo.EventRequest) []string {
	keys := make([]string, 0, len(request))
	for key := range request {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isQueryableRequestField(protocol, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "http":
		switch key {
		case "method", "host", "path":
			return true
		default:
			return false
		}
	case "cache":
		return key == "operation" || key == "key"
	case "spex":
		return key == "cmd"
	default:
		return false
	}
}

func addField(fields map[string]string, path, value string) {
	value = strings.TrimSpace(value)
	if path == "" || len([]rune(path)) > 128 || value == "" {
		return
	}
	fields[path] = value
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []string:
		return strings.Join(v, ",")
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if value := stringFromAny(item); value != "" {
				parts = append(parts, value)
			}
		}
		return strings.Join(parts, ",")
	default:
		return ""
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
