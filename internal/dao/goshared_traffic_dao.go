package dao

import (
	"context"
	"fmt"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	modelfmo "github.com/MrMiaoMIMI/mockserver/internal/model/fmo"
)

const trafficStatsLimit = 5000

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

func (d *gosharedTrafficTableDAO) ListTrafficEventsForStats(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64, limit int) ([]modeldo.TrafficEvent, error) {
	if limit <= 0 || limit > trafficStatsLimit {
		limit = trafficStatsLimit
	}
	dbQuery := d.buildEventQuery(query, eventIDs)
	pagination := dbhelper.NewPagination().
		WithLimit(&limit).
		AppendOrder(dbhelper.Desc(d.eventFields.EventTime))
	items, err := d.eventStore.FindNotDeleted(ctx, dbQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("list traffic events for stats: %w", err)
	}
	result := make([]modeldo.TrafficEvent, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
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
	items, err := d.indexStore.FindNotDeleted(ctx, dbhelper.Q(conditions...), nil)
	if err != nil {
		return nil, fmt.Errorf("list traffic index filter %s: %w", filter.FieldPath, err)
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
	if query.TraceID != "" {
		conditions = append(conditions, d.eventFields.TraceID.Eq(&query.TraceID))
	}
	if query.TrafficSource != "" {
		conditions = append(conditions, d.eventFields.TrafficSource.Eq(&query.TrafficSource))
	}
	if query.ProtocolName != "" {
		conditions = append(conditions, d.eventFields.ProtocolName.Eq(&query.ProtocolName))
	}
	if query.NamespaceDBID > 0 {
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
	if query.RuleSetDBID > 0 {
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
