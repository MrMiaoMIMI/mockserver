package dao

import (
	"context"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	"github.com/MrMiaoMIMI/mockserver/internal/trafficutil"
)

type trafficRepository struct {
	tableDAO          trafficTableDAO
	namespaceTableDAO namespaceTableDAO
	ruleSetTableDAO   ruleSetTableDAO
}

func newTrafficRepository(tableDAO trafficTableDAO, namespaceTableDAO namespaceTableDAO, ruleSetTableDAO ruleSetTableDAO) TrafficRepository {
	return &trafficRepository{
		tableDAO:          tableDAO,
		namespaceTableDAO: namespaceTableDAO,
		ruleSetTableDAO:   ruleSetTableDAO,
	}
}

func (r *trafficRepository) CreateTrafficEvent(ctx context.Context, event bo.TrafficEvent) (bo.TrafficEvent, error) {
	event, err := r.resolveEventReferences(ctx, event)
	if err != nil {
		return bo.TrafficEvent{}, err
	}
	record, err := r.tableDAO.CreateTrafficEvent(ctx, encodeTrafficEventRecord(event), encodeTrafficEventIndexRecords(event.Indexes))
	if err != nil {
		return bo.TrafficEvent{}, err
	}
	event.ID = record.Id
	for i := range event.Indexes {
		event.Indexes[i].TrafficEventID = record.Id
		event.Indexes[i].ProtocolName = event.ProtocolName
		event.Indexes[i].EventTime = event.EventTime
		event.Indexes[i].ExpireTime = event.ExpireTime
	}
	return event, nil
}

func (r *trafficRepository) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
	var ok bool
	var err error
	query, ok, err = r.resolveQueryReferences(ctx, query)
	if err != nil {
		return bo.TrafficEventList{}, err
	}
	if !ok {
		return bo.TrafficEventList{Stats: emptyTrafficStats()}, nil
	}
	eventIDs, ok, err := r.resolveIndexFilterEventIDs(ctx, query)
	if err != nil {
		return bo.TrafficEventList{}, err
	}
	if !ok {
		return bo.TrafficEventList{Stats: emptyTrafficStats()}, nil
	}

	records, total, err := r.tableDAO.ListTrafficEvents(ctx, query, eventIDs)
	if err != nil {
		return bo.TrafficEventList{}, err
	}
	items := decodeTrafficEventRecords(records)
	stats, err := r.tableDAO.TrafficStats(ctx, query, eventIDs)
	if err != nil {
		return bo.TrafficEventList{}, err
	}
	stats.Total = total
	return bo.TrafficEventList{
		Items: items,
		Total: total,
		Stats: stats,
	}, nil
}

func (r *trafficRepository) GetTrafficEvent(ctx context.Context, id uint64) (bo.TrafficEvent, bool, error) {
	record, ok, err := r.tableDAO.GetTrafficEvent(ctx, id)
	if err != nil {
		return bo.TrafficEvent{}, false, err
	}
	if !ok {
		return bo.TrafficEvent{}, false, nil
	}
	items := []bo.TrafficEvent{decodeTrafficEventRecord(record)}
	if err := r.attachTrafficIndexes(ctx, items); err != nil {
		return bo.TrafficEvent{}, false, err
	}
	return items[0], true, nil
}

func (r *trafficRepository) resolveEventReferences(ctx context.Context, event bo.TrafficEvent) (bo.TrafficEvent, error) {
	if event.NamespaceID != "" && r.namespaceTableDAO != nil {
		namespace, ok, err := r.namespaceTableDAO.GetNamespace(ctx, event.NamespaceID)
		if err != nil {
			return bo.TrafficEvent{}, err
		}
		if ok {
			event.NamespaceDBID = namespace.Id
		}
	}
	if event.RuleSetID != "" && r.ruleSetTableDAO != nil {
		ruleSet, ok, err := r.ruleSetTableDAO.GetDraft(ctx, event.RuleSetID)
		if err != nil {
			return bo.TrafficEvent{}, err
		}
		if ok {
			event.RuleSetDBID = ruleSet.Id
		}
		if event.SnapshotID != "" {
			snapshot, ok, err := r.ruleSetTableDAO.GetPublishedSnapshot(ctx, event.RuleSetID, event.SnapshotID)
			if err != nil {
				return bo.TrafficEvent{}, err
			}
			if ok {
				event.SnapshotDBID = snapshot.Id
			}
		}
	}
	return event, nil
}

func (r *trafficRepository) resolveQueryReferences(ctx context.Context, query bo.TrafficQuery) (bo.TrafficQuery, bool, error) {
	if query.NamespaceID != "" && r.namespaceTableDAO != nil {
		namespace, ok, err := r.namespaceTableDAO.GetNamespace(ctx, query.NamespaceID)
		if err != nil {
			return bo.TrafficQuery{}, false, err
		}
		if ok {
			query.NamespaceDBID = namespace.Id
		}
	}
	if query.RuleSetID != "" && r.ruleSetTableDAO != nil {
		ruleSet, ok, err := r.ruleSetTableDAO.GetDraft(ctx, query.RuleSetID)
		if err != nil {
			return bo.TrafficQuery{}, false, err
		}
		if ok {
			query.RuleSetDBID = ruleSet.Id
		}
	}
	return query, true, nil
}

func (r *trafficRepository) resolveIndexFilterEventIDs(ctx context.Context, query bo.TrafficQuery) ([]uint64, bool, error) {
	if len(query.IndexFilters) == 0 {
		return nil, true, nil
	}
	var current map[uint64]struct{}
	for _, filter := range query.IndexFilters {
		valueHash := trafficutil.FieldValueHash(filter.FieldValue)
		ids, err := r.tableDAO.ListTrafficEventIDsByIndexFilter(ctx, query, filter, valueHash)
		if err != nil {
			return nil, false, err
		}
		next := make(map[uint64]struct{}, len(ids))
		for _, id := range ids {
			if current == nil {
				next[id] = struct{}{}
				continue
			}
			if _, ok := current[id]; ok {
				next[id] = struct{}{}
			}
		}
		current = next
		if len(current) == 0 {
			return nil, false, nil
		}
	}
	result := make([]uint64, 0, len(current))
	for id := range current {
		result = append(result, id)
	}
	return result, true, nil
}

func (r *trafficRepository) attachTrafficIndexes(ctx context.Context, items []bo.TrafficEvent) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	indexes, err := r.tableDAO.ListTrafficEventIndexesByEventIDs(ctx, ids)
	if err != nil {
		return err
	}
	for i := range items {
		items[i].Indexes = decodeTrafficEventIndexRecords(indexes[items[i].ID])
	}
	return nil
}

func encodeTrafficEventRecord(event bo.TrafficEvent) modeldo.TrafficEvent {
	return modeldo.TrafficEvent{
		EventCode:      event.EventID,
		TraceID:        event.TraceID,
		ScenarioCode:   event.ScenarioID,
		TrafficSource:  event.TrafficSource,
		ProtocolName:   event.ProtocolName,
		NamespaceID:    event.NamespaceDBID,
		NamespaceName:  event.NamespaceID,
		OperationName:  event.OperationName,
		Outcome:        event.Outcome,
		DecisionKind:   event.DecisionKind,
		RuleSetID:      event.RuleSetDBID,
		RuleSetCode:    event.RuleSetID,
		RuleCode:       event.RuleID,
		SnapshotID:     event.SnapshotDBID,
		SnapshotCode:   event.SnapshotID,
		FallbackReason: event.FallbackReason,
		DurationMS:     event.DurationMS,
		EventTime:      event.EventTime,
		ExpireTime:     event.ExpireTime,
		EventJSON:      event.EventJSON,
		DecisionJSON:   event.DecisionJSON,
		ExplainJSON:    event.ExplainJSON,
		ErrorMessage:   event.ErrorMessage,
	}
}

func decodeTrafficEventRecords(records []modeldo.TrafficEvent) []bo.TrafficEvent {
	items := make([]bo.TrafficEvent, 0, len(records))
	for _, record := range records {
		items = append(items, decodeTrafficEventRecord(record))
	}
	return items
}

func decodeTrafficEventRecord(record modeldo.TrafficEvent) bo.TrafficEvent {
	return bo.TrafficEvent{
		ID:             record.Id,
		EventID:        record.EventCode,
		TraceID:        record.TraceID,
		ScenarioID:     record.ScenarioCode,
		TrafficSource:  record.TrafficSource,
		ProtocolName:   record.ProtocolName,
		NamespaceDBID:  record.NamespaceID,
		NamespaceID:    record.NamespaceName,
		OperationName:  record.OperationName,
		Outcome:        record.Outcome,
		DecisionKind:   record.DecisionKind,
		RuleSetDBID:    record.RuleSetID,
		RuleSetID:      record.RuleSetCode,
		RuleID:         record.RuleCode,
		SnapshotDBID:   record.SnapshotID,
		SnapshotID:     record.SnapshotCode,
		FallbackReason: record.FallbackReason,
		DurationMS:     record.DurationMS,
		EventTime:      record.EventTime,
		ExpireTime:     record.ExpireTime,
		EventJSON:      record.EventJSON,
		DecisionJSON:   record.DecisionJSON,
		ExplainJSON:    record.ExplainJSON,
		ErrorMessage:   record.ErrorMessage,
	}
}

func encodeTrafficEventIndexRecords(indexes []bo.TrafficEventIndex) []modeldo.TrafficEventIndex {
	records := make([]modeldo.TrafficEventIndex, 0, len(indexes))
	for _, index := range indexes {
		records = append(records, modeldo.TrafficEventIndex{
			TrafficEventID:    index.TrafficEventID,
			ProtocolName:      index.ProtocolName,
			FieldPath:         index.FieldPath,
			FieldValuePreview: index.FieldValuePreview,
			FieldValueHash:    index.FieldValueHash,
			FieldValueText:    index.FieldValueText,
			EventTime:         index.EventTime,
			ExpireTime:        index.ExpireTime,
		})
	}
	return records
}

func decodeTrafficEventIndexRecords(records []modeldo.TrafficEventIndex) []bo.TrafficEventIndex {
	items := make([]bo.TrafficEventIndex, 0, len(records))
	for _, record := range records {
		items = append(items, bo.TrafficEventIndex{
			ID:                record.Id,
			TrafficEventID:    record.TrafficEventID,
			ProtocolName:      record.ProtocolName,
			FieldPath:         record.FieldPath,
			FieldValuePreview: record.FieldValuePreview,
			FieldValueHash:    record.FieldValueHash,
			FieldValueText:    record.FieldValueText,
			EventTime:         record.EventTime,
			ExpireTime:        record.ExpireTime,
		})
	}
	return items
}

func emptyTrafficStats() bo.TrafficStats {
	return bo.TrafficStats{
		ByOutcome:   map[string]uint64{},
		ByProtocol:  map[string]uint64{},
		ByNamespace: map[string]uint64{},
	}
}
