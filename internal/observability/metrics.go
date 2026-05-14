package observability

import (
	"strconv"
	"sync"
	"time"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

const maxRecentRuntimeRequests = 80
const RuntimeSourceHTTP = "http_runtime"

type RuntimeObservation struct {
	Source         string
	Protocol       string
	Operation      string
	Namespace      string
	Method         string
	Scheme         string
	Host           string
	Path           string
	RawQuery       string
	TraceID        string
	Matched        bool
	Fallback       bool
	Error          bool
	Status         int
	RulesetID      string
	RuleID         string
	FallbackReason string
	Message        string
	Duration       time.Duration
	Event          *bo.Event
}

type RuntimeRequestRecord struct {
	ID             int64     `json:"id"`
	ObservedAt     time.Time `json:"observed_at"`
	Source         string    `json:"source,omitempty"`
	Protocol       string    `json:"protocol,omitempty"`
	Operation      string    `json:"operation,omitempty"`
	Namespace      string    `json:"namespace,omitempty"`
	Method         string    `json:"method,omitempty"`
	Scheme         string    `json:"scheme,omitempty"`
	Host           string    `json:"host,omitempty"`
	Path           string    `json:"path,omitempty"`
	RawQuery       string    `json:"raw_query,omitempty"`
	TraceID        string    `json:"trace_id,omitempty"`
	Outcome        string    `json:"outcome"`
	Matched        bool      `json:"matched"`
	Fallback       bool      `json:"fallback,omitempty"`
	Error          bool      `json:"error,omitempty"`
	Status         int       `json:"status,omitempty"`
	RulesetID      string    `json:"ruleset_id,omitempty"`
	RuleID         string    `json:"rule_id,omitempty"`
	FallbackReason string    `json:"fallback_reason,omitempty"`
	Message        string    `json:"message,omitempty"`
	DurationMS     int64     `json:"duration_ms"`
	Event          *bo.Event `json:"event,omitempty"`
}

type RuntimeMetricsSnapshot struct {
	TotalRequests       int64                   `json:"total_requests"`
	MatchedRequests     int64                   `json:"matched_requests"`
	UnmatchedRequests   int64                   `json:"unmatched_requests"`
	ErrorRequests       int64                   `json:"error_requests"`
	TotalDurationMS     int64                   `json:"total_duration_ms"`
	AverageDurationMS   float64                 `json:"average_duration_ms"`
	Sources             map[string]int64        `json:"sources,omitempty"`
	Protocols           map[string]int64        `json:"protocols,omitempty"`
	Operations          map[string]int64        `json:"operations,omitempty"`
	RulesetMatches      map[string]int64        `json:"ruleset_matches,omitempty"`
	RuleMatches         map[string]int64        `json:"rule_matches,omitempty"`
	LastMatchedAtByRule map[string]time.Time    `json:"last_matched_at_by_rule,omitempty"`
	FallbackReasons     map[string]int64        `json:"fallback_reasons,omitempty"`
	FallbackStats       map[string]FallbackStat `json:"fallback_stats,omitempty"`
	StatusCodes         map[string]int64        `json:"status_codes,omitempty"`
	RecentRequests      []RuntimeRequestRecord  `json:"recent_requests,omitempty"`
}

type FallbackStat struct {
	Total       int64            `json:"total"`
	ByProtocol  map[string]int64 `json:"by_protocol,omitempty"`
	ByNamespace map[string]int64 `json:"by_namespace,omitempty"`
	ByOperation map[string]int64 `json:"by_operation,omitempty"`
}

type RuntimeMetrics struct {
	mu                  sync.RWMutex
	nextRecordID        int64
	totalRequests       int64
	matchedRequests     int64
	unmatchedRequests   int64
	errorRequests       int64
	totalDuration       time.Duration
	sources             map[string]int64
	protocols           map[string]int64
	operations          map[string]int64
	rulesetMatches      map[string]int64
	ruleMatches         map[string]int64
	lastMatchedAtByRule map[string]time.Time
	fallbackReasons     map[string]int64
	fallbackStats       map[string]FallbackStat
	statusCodes         map[string]int64
	recentRequests      []RuntimeRequestRecord
}

func NewRuntimeMetrics() *RuntimeMetrics {
	return &RuntimeMetrics{
		sources:             make(map[string]int64),
		protocols:           make(map[string]int64),
		operations:          make(map[string]int64),
		rulesetMatches:      make(map[string]int64),
		ruleMatches:         make(map[string]int64),
		lastMatchedAtByRule: make(map[string]time.Time),
		fallbackReasons:     make(map[string]int64),
		fallbackStats:       make(map[string]FallbackStat),
		statusCodes:         make(map[string]int64),
	}
}

func (m *RuntimeMetrics) Observe(observation RuntimeObservation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.totalDuration += observation.Duration
	incrementIfPresent(m.sources, observation.Source)
	incrementIfPresent(m.protocols, observation.Protocol)
	incrementIfPresent(m.operations, observation.Operation)
	switch {
	case observation.Error:
		m.errorRequests++
	case observation.Matched:
		m.matchedRequests++
		if observation.RulesetID != "" {
			m.rulesetMatches[observation.RulesetID]++
		}
		if observation.RuleID != "" {
			m.ruleMatches[observation.RuleID]++
			m.lastMatchedAtByRule[observation.RuleID] = time.Now().UTC()
		}
	default:
		m.unmatchedRequests++
	}
	if observation.Fallback && observation.FallbackReason != "" {
		m.fallbackReasons[observation.FallbackReason]++
		m.incrementFallbackStat(observation)
	}
	if observation.Status > 0 {
		m.statusCodes[statusCodeKey(observation.Status)]++
	}
	m.nextRecordID++
	m.recentRequests = append(m.recentRequests, runtimeRequestRecord(m.nextRecordID, observation))
	if len(m.recentRequests) > maxRecentRuntimeRequests {
		m.recentRequests = append([]RuntimeRequestRecord(nil), m.recentRequests[len(m.recentRequests)-maxRecentRuntimeRequests:]...)
	}
}

func (m *RuntimeMetrics) Snapshot() RuntimeMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := RuntimeMetricsSnapshot{
		TotalRequests:       m.totalRequests,
		MatchedRequests:     m.matchedRequests,
		UnmatchedRequests:   m.unmatchedRequests,
		ErrorRequests:       m.errorRequests,
		TotalDurationMS:     m.totalDuration.Milliseconds(),
		Sources:             cloneInt64Map(m.sources),
		Protocols:           cloneInt64Map(m.protocols),
		Operations:          cloneInt64Map(m.operations),
		RulesetMatches:      cloneInt64Map(m.rulesetMatches),
		RuleMatches:         cloneInt64Map(m.ruleMatches),
		LastMatchedAtByRule: cloneTimeMap(m.lastMatchedAtByRule),
		FallbackReasons:     cloneInt64Map(m.fallbackReasons),
		FallbackStats:       cloneFallbackStats(m.fallbackStats),
		StatusCodes:         cloneInt64Map(m.statusCodes),
		RecentRequests:      cloneRecentRequests(m.recentRequests),
	}
	if m.totalRequests > 0 {
		snapshot.AverageDurationMS = float64(m.totalDuration.Microseconds()) / 1000 / float64(m.totalRequests)
	}
	return snapshot
}

func runtimeRequestRecord(id int64, observation RuntimeObservation) RuntimeRequestRecord {
	return RuntimeRequestRecord{
		ID:             id,
		ObservedAt:     time.Now().UTC(),
		Source:         observation.Source,
		Protocol:       observation.Protocol,
		Operation:      observation.Operation,
		Namespace:      observation.Namespace,
		Method:         observation.Method,
		Scheme:         observation.Scheme,
		Host:           observation.Host,
		Path:           observation.Path,
		RawQuery:       observation.RawQuery,
		TraceID:        observation.TraceID,
		Outcome:        runtimeOutcome(observation),
		Matched:        observation.Matched,
		Fallback:       observation.Fallback,
		Error:          observation.Error,
		Status:         observation.Status,
		RulesetID:      observation.RulesetID,
		RuleID:         observation.RuleID,
		FallbackReason: observation.FallbackReason,
		Message:        observation.Message,
		DurationMS:     observation.Duration.Milliseconds(),
		Event:          cloneEvent(observation.Event),
	}
}

func (m *RuntimeMetrics) incrementFallbackStat(observation RuntimeObservation) {
	reason := observation.FallbackReason
	stat := m.fallbackStats[reason]
	stat.Total++
	if observation.Protocol != "" {
		if stat.ByProtocol == nil {
			stat.ByProtocol = make(map[string]int64)
		}
		stat.ByProtocol[observation.Protocol]++
	}
	if observation.Namespace != "" {
		if stat.ByNamespace == nil {
			stat.ByNamespace = make(map[string]int64)
		}
		stat.ByNamespace[observation.Namespace]++
	}
	if observation.Operation != "" {
		if stat.ByOperation == nil {
			stat.ByOperation = make(map[string]int64)
		}
		stat.ByOperation[observation.Operation]++
	}
	m.fallbackStats[reason] = stat
}

func incrementIfPresent(values map[string]int64, key string) {
	if key != "" {
		values[key]++
	}
}

func runtimeOutcome(observation RuntimeObservation) string {
	switch {
	case observation.Error:
		return "error"
	case observation.Matched:
		return "matched"
	case observation.Fallback:
		return "fallback"
	default:
		return "unmatched"
	}
}

func statusCodeKey(status int) string {
	return strconv.Itoa(status/100) + "xx"
}

func cloneInt64Map(src map[string]int64) map[string]int64 {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]int64, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneTimeMap(src map[string]time.Time) map[string]time.Time {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]time.Time, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneFallbackStats(src map[string]FallbackStat) map[string]FallbackStat {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]FallbackStat, len(src))
	for key, value := range src {
		dst[key] = FallbackStat{
			Total:       value.Total,
			ByProtocol:  cloneInt64Map(value.ByProtocol),
			ByNamespace: cloneInt64Map(value.ByNamespace),
			ByOperation: cloneInt64Map(value.ByOperation),
		}
	}
	return dst
}

func cloneRecentRequests(src []RuntimeRequestRecord) []RuntimeRequestRecord {
	if len(src) == 0 {
		return nil
	}
	dst := make([]RuntimeRequestRecord, 0, len(src))
	for i := len(src) - 1; i >= 0; i-- {
		item := src[i]
		item.Event = cloneEvent(item.Event)
		dst = append(dst, item)
	}
	return dst
}

func cloneEvent(src *bo.Event) *bo.Event {
	if src == nil {
		return nil
	}
	dst := *src
	if src.Request != nil {
		dst.Request = make(bo.EventRequest, len(src.Request))
		for key, value := range src.Request {
			switch key {
			case "query", "headers":
				if cloned := cloneStringSliceMap(bo.RequestStringMap(src.Request, key)); cloned != nil {
					dst.Request[key] = cloned
					continue
				}
			}
			dst.Request[key] = value
		}
	}
	return &dst
}

func cloneStringSliceMap(src map[string][]string) map[string][]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string][]string, len(src))
	for key, values := range src {
		dst[key] = append([]string(nil), values...)
	}
	return dst
}
