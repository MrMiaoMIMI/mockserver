package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	modelfmo "github.com/MrMiaoMIMI/mockserver/internal/model/fmo"
)

const maxTrafficIndexFilterEventIDs = 5000

type gosharedTrafficTableDAO struct {
	manager     dbspi.Manager
	eventStore  dbspi.SoftDeleteTableStore[*modeldo.TrafficEvent]
	indexStore  dbspi.SoftDeleteTableStore[*modeldo.TrafficEventIndex]
	eventFields modelfmo.TrafficEventFields
	indexFields modelfmo.TrafficEventIndexFields
}

func newGosharedTrafficTableDAO(manager dbspi.Manager) trafficTableDAO {
	return &gosharedTrafficTableDAO{
		manager:     manager,
		eventStore:  dbhelper.NewSoftDeleteTableStore(&modeldo.TrafficEvent{}, dbhelper.WithManager(manager)),
		indexStore:  dbhelper.NewSoftDeleteTableStore(&modeldo.TrafficEventIndex{}, dbhelper.WithManager(manager)),
		eventFields: modelfmo.NewTrafficEventFields(),
		indexFields: modelfmo.NewTrafficEventIndexFields(),
	}
}

func (d *gosharedTrafficTableDAO) CreateTrafficEvent(ctx context.Context, event modeldo.TrafficEvent, indexes []modeldo.TrafficEventIndex) (modeldo.TrafficEvent, error) {
	if err := dbhelper.Transaction(ctx, func(tx *dbhelper.Tx) error {
		eventStore := dbhelper.NewSoftDeleteTableStore(&modeldo.TrafficEvent{}, dbhelper.WithTx(tx))
		indexStore := dbhelper.NewSoftDeleteTableStore(&modeldo.TrafficEventIndex{}, dbhelper.WithTx(tx))
		if err := eventStore.Create(ctx, &event); err != nil {
			return fmt.Errorf("insert traffic event %s: %w", event.EventCode, err)
		}
		if len(indexes) == 0 {
			return nil
		}
		rows := make([]*modeldo.TrafficEventIndex, 0, len(indexes))
		for i := range indexes {
			indexes[i].TrafficEventID = event.Id
			indexes[i].ProtocolName = event.ProtocolName
			indexes[i].EventTime = event.EventTime
			indexes[i].ExpireTime = event.ExpireTime
			rows = append(rows, &indexes[i])
		}
		if err := indexStore.BatchCreate(ctx, rows, 100); err != nil {
			return fmt.Errorf("insert traffic event indexes %s: %w", event.EventCode, err)
		}
		return nil
	}, dbhelper.WithManager(d.manager)); err != nil {
		return modeldo.TrafficEvent{}, err
	}
	return event, nil
}

func (d *gosharedTrafficTableDAO) GetTrafficEvent(ctx context.Context, id uint64) (modeldo.TrafficEvent, bool, error) {
	limit := 1
	items, err := d.eventStore.FindNotDeleted(ctx, dbhelper.Q(d.eventFields.ID.Eq(&id)), dbhelper.NewPagination().WithLimit(&limit))
	if err != nil {
		return modeldo.TrafficEvent{}, false, fmt.Errorf("get traffic event %d: %w", id, err)
	}
	if len(items) == 0 || items[0] == nil {
		return modeldo.TrafficEvent{}, false, nil
	}
	return *items[0], true, nil
}

func (d *gosharedTrafficTableDAO) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64) ([]modeldo.TrafficEvent, uint64, error) {
	dbQuery := d.buildEventQuery(query, eventIDs)
	total, err := d.eventStore.CountNotDeleted(ctx, dbQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("count traffic events: %w", err)
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	pagination := dbhelper.NewPagination().
		WithLimit(&limit).
		WithOffset(&offset).
		AppendOrder(dbhelper.Desc(d.eventFields.EventTime)).
		AppendOrder(dbhelper.Desc(d.eventFields.ID))
	items, err := d.eventStore.FindNotDeleted(ctx, dbQuery, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("list traffic events: %w", err)
	}
	result := make([]modeldo.TrafficEvent, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, total, nil
}

func (d *gosharedTrafficTableDAO) TrafficStats(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64) (bo.TrafficStats, error) {
	stats := emptyTrafficStats()
	total, err := d.eventStore.CountNotDeleted(ctx, d.buildEventQuery(query, eventIDs))
	if err != nil {
		return bo.TrafficStats{}, fmt.Errorf("count traffic stats: %w", err)
	}
	stats.Total = total
	outcomes, err := d.groupTrafficEvents(ctx, query, eventIDs, "outcome")
	if err != nil {
		return bo.TrafficStats{}, err
	}
	protocols, err := d.groupTrafficEvents(ctx, query, eventIDs, "protocol_name")
	if err != nil {
		return bo.TrafficStats{}, err
	}
	namespaces, err := d.groupTrafficEvents(ctx, query, eventIDs, "namespace_code")
	if err != nil {
		return bo.TrafficStats{}, err
	}
	stats.ByOutcome = outcomes
	stats.ByProtocol = protocols
	stats.ByNamespace = namespaces
	return stats, nil
}

func (d *gosharedTrafficTableDAO) groupTrafficEvents(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64, column string) (map[string]uint64, error) {
	column, ok := trafficGroupColumn(column)
	if !ok {
		return nil, fmt.Errorf("unsupported traffic stats group column %q", column)
	}
	sqlStore, ok := d.eventStore.(dbspi.SQLTableStore[*modeldo.TrafficEvent])
	if !ok {
		return d.groupTrafficEventsFallback(ctx, query, eventIDs, column)
	}
	whereSQL, args := d.buildEventWhereSQL(query, eventIDs)
	sql := fmt.Sprintf("SELECT %s, COUNT(*) AS id FROM mockserver_traffic_event_tab %s GROUP BY %s", column, whereSQL, column)
	rows, err := sqlStore.Raw(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("group traffic events by %s: %w", column, err)
	}
	result := make(map[string]uint64, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		key := trafficGroupKey(*row, column)
		if key != "" {
			result[key] = row.Id
		}
	}
	return result, nil
}

func (d *gosharedTrafficTableDAO) groupTrafficEventsFallback(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64, column string) (map[string]uint64, error) {
	items, err := d.eventStore.FindNotDeleted(ctx, d.buildEventQuery(query, eventIDs), nil)
	if err != nil {
		return nil, fmt.Errorf("group traffic events fallback by %s: %w", column, err)
	}
	result := map[string]uint64{}
	for _, item := range items {
		if item == nil {
			continue
		}
		if key := trafficGroupKey(*item, column); key != "" {
			result[key]++
		}
	}
	return result, nil
}

func trafficGroupColumn(column string) (string, bool) {
	switch column {
	case "outcome", "protocol_name", "namespace_code":
		return column, true
	default:
		return "", false
	}
}

func trafficGroupKey(record modeldo.TrafficEvent, column string) string {
	switch column {
	case "outcome":
		return record.Outcome
	case "protocol_name":
		return record.ProtocolName
	case "namespace_code":
		return record.NamespaceCode
	default:
		return ""
	}
}

func (d *gosharedTrafficTableDAO) ListTrafficEventIndexesByEventIDs(ctx context.Context, eventIDs []uint64) (map[uint64][]modeldo.TrafficEventIndex, error) {
	result := make(map[uint64][]modeldo.TrafficEventIndex)
	if len(eventIDs) == 0 {
		return result, nil
	}
	query := dbhelper.Q(d.indexFields.TrafficEventID.In(eventIDs))
	items, err := d.indexStore.FindNotDeleted(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("list traffic event indexes: %w", err)
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		result[item.TrafficEventID] = append(result[item.TrafficEventID], *item)
	}
	return result, nil
}

func (d *gosharedTrafficTableDAO) ListTrafficEventIDsByIndexFilter(ctx context.Context, query bo.TrafficQuery, filter bo.TrafficIndexFilter, valueHash uint64) ([]uint64, error) {
	limit := maxTrafficIndexFilterEventIDs + 1
	conditions := []dbspi.Condition{
		d.indexFields.FieldPath.Eq(&filter.FieldPath),
		d.indexFields.FieldValueHash.Eq(&valueHash),
		d.indexFields.FieldValueText.Eq(&filter.FieldValue),
	}
	if query.ProtocolName != "" {
		conditions = append(conditions, d.indexFields.ProtocolName.Eq(&query.ProtocolName))
	}
	if query.StartTime > 0 {
		conditions = append(conditions, d.indexFields.EventTime.GtEq(&query.StartTime))
	}
	if query.EndTime > 0 {
		conditions = append(conditions, d.indexFields.EventTime.LtEq(&query.EndTime))
	}
	pagination := dbhelper.NewPagination().
		WithLimit(&limit).
		AppendOrder(dbhelper.Desc(d.indexFields.EventTime)).
		AppendOrder(dbhelper.Desc(d.indexFields.TrafficEventID))
	items, err := d.indexStore.FindNotDeleted(ctx, dbhelper.Q(conditions...), pagination)
	if err != nil {
		return nil, fmt.Errorf("list traffic index filter %s: %w", filter.FieldPath, err)
	}
	if len(items) > maxTrafficIndexFilterEventIDs {
		return nil, fmt.Errorf("traffic index filter %s is too broad; narrow the time range or add filters", filter.FieldPath)
	}
	ids := make([]uint64, 0, len(items))
	seen := make(map[uint64]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := seen[item.TrafficEventID]; ok {
			continue
		}
		seen[item.TrafficEventID] = struct{}{}
		ids = append(ids, item.TrafficEventID)
	}
	return ids, nil
}

func (d *gosharedTrafficTableDAO) buildEventQuery(query bo.TrafficQuery, eventIDs []uint64) dbspi.Query {
	conditions := make([]dbspi.Condition, 0, 12)
	if len(eventIDs) > 0 {
		conditions = append(conditions, d.eventFields.ID.In(eventIDs))
	}
	if query.StartTime > 0 {
		conditions = append(conditions, d.eventFields.EventTime.GtEq(&query.StartTime))
	}
	if query.EndTime > 0 {
		conditions = append(conditions, d.eventFields.EventTime.LtEq(&query.EndTime))
	}
	if query.EventID != "" {
		conditions = append(conditions, d.eventFields.EventCode.Eq(&query.EventID))
	}
	if query.TraceID != "" {
		conditions = append(conditions, d.eventFields.TraceID.Eq(&query.TraceID))
	}
	if query.TrafficSource != "" {
		conditions = append(conditions, d.eventFields.TrafficSource.Eq(&query.TrafficSource))
	}
	if query.ProtocolName != "" {
		conditions = append(conditions, d.eventFields.ProtocolName.Eq(&query.ProtocolName))
	}
	if query.NamespaceDBID > 0 && query.NamespaceID != "" {
		conditions = append(conditions, dbhelper.Or(
			d.eventFields.NamespaceID.Eq(&query.NamespaceDBID),
			d.eventFields.NamespaceCode.Eq(&query.NamespaceID),
		))
	} else if query.NamespaceDBID > 0 {
		conditions = append(conditions, d.eventFields.NamespaceID.Eq(&query.NamespaceDBID))
	} else if query.NamespaceID != "" {
		conditions = append(conditions, d.eventFields.NamespaceCode.Eq(&query.NamespaceID))
	}
	if query.OperationName != "" {
		conditions = append(conditions, d.eventFields.OperationName.Eq(&query.OperationName))
	}
	if query.Outcome != "" {
		conditions = append(conditions, d.eventFields.Outcome.Eq(&query.Outcome))
	}
	if query.DecisionKind != "" {
		conditions = append(conditions, d.eventFields.DecisionKind.Eq(&query.DecisionKind))
	}
	if query.RuleSetDBID > 0 && query.RuleSetID != "" {
		conditions = append(conditions, dbhelper.Or(
			d.eventFields.RuleSetID.Eq(&query.RuleSetDBID),
			d.eventFields.RuleSetCode.Eq(&query.RuleSetID),
		))
	} else if query.RuleSetDBID > 0 {
		conditions = append(conditions, d.eventFields.RuleSetID.Eq(&query.RuleSetDBID))
	} else if query.RuleSetID != "" {
		conditions = append(conditions, d.eventFields.RuleSetCode.Eq(&query.RuleSetID))
	}
	if query.RuleID != "" {
		conditions = append(conditions, d.eventFields.RuleCode.Eq(&query.RuleID))
	}
	if query.FallbackReason != "" {
		conditions = append(conditions, d.eventFields.FallbackReason.Eq(&query.FallbackReason))
	}
	if len(conditions) == 0 {
		return nil
	}
	return dbhelper.Q(conditions...)
}

func (d *gosharedTrafficTableDAO) buildEventWhereSQL(query bo.TrafficQuery, eventIDs []uint64) (string, []any) {
	conditions := []string{"deleted = 0"}
	args := []any{}
	appendInCondition := func(column string, ids []uint64) {
		if len(ids) == 0 {
			return
		}
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ",")))
	}
	appendEqCondition := func(column string, value any, present bool) {
		if !present {
			return
		}
		conditions = append(conditions, column+" = ?")
		args = append(args, value)
	}
	appendInCondition("id", eventIDs)
	if query.StartTime > 0 {
		conditions = append(conditions, "event_time >= ?")
		args = append(args, query.StartTime)
	}
	if query.EndTime > 0 {
		conditions = append(conditions, "event_time <= ?")
		args = append(args, query.EndTime)
	}
	appendEqCondition("event_code", query.EventID, query.EventID != "")
	appendEqCondition("trace_id", query.TraceID, query.TraceID != "")
	appendEqCondition("traffic_source", query.TrafficSource, query.TrafficSource != "")
	appendEqCondition("protocol_name", query.ProtocolName, query.ProtocolName != "")
	if query.NamespaceDBID > 0 && query.NamespaceID != "" {
		conditions = append(conditions, "(namespace_id = ? OR namespace_code = ?)")
		args = append(args, query.NamespaceDBID, query.NamespaceID)
	} else {
		appendEqCondition("namespace_id", query.NamespaceDBID, query.NamespaceDBID > 0)
		appendEqCondition("namespace_code", query.NamespaceID, query.NamespaceID != "")
	}
	appendEqCondition("operation_name", query.OperationName, query.OperationName != "")
	appendEqCondition("outcome", query.Outcome, query.Outcome != "")
	appendEqCondition("decision_kind", query.DecisionKind, query.DecisionKind != "")
	if query.RuleSetDBID > 0 && query.RuleSetID != "" {
		conditions = append(conditions, "(ruleset_id = ? OR ruleset_code = ?)")
		args = append(args, query.RuleSetDBID, query.RuleSetID)
	} else {
		appendEqCondition("ruleset_id", query.RuleSetDBID, query.RuleSetDBID > 0)
		appendEqCondition("ruleset_code", query.RuleSetID, query.RuleSetID != "")
	}
	appendEqCondition("rule_code", query.RuleID, query.RuleID != "")
	appendEqCondition("fallback_reason", query.FallbackReason, query.FallbackReason != "")
	return "WHERE " + strings.Join(conditions, " AND "), args
}
