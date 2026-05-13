package dao

import (
	"context"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	"github.com/MrMiaoMIMI/mockserver/internal/trafficutil"
)

type trafficRepository struct {
	tableDAO trafficTableDAO
}

func newTrafficRepository(tableDAO trafficTableDAO) TrafficRepository {
	return &trafficRepository{tableDAO: tableDAO}
}

func (r *trafficRepository) CreateTrafficEvent(ctx context.Context, event bo.TrafficEvent) (bo.TrafficEvent, error) {
	record, err := r.tableDAO.CreateTrafficEvent(ctx, encodeTrafficEventRecord(event), encodeTrafficEventIndexRecords(event.Indexes))
	if err != nil {
		return bo.TrafficEvent{}, err
	}
	event.ID = record.Id
	for i := range event.Indexes {
		event.Indexes[i].TrafficEventID = record.Id
		event.Indexes[i].EventID = event.EventID
		event.Indexes[i].ProtocolName = event.ProtocolName
		event.Indexes[i].EventTime = event.EventTime
		event.Indexes[i].ExpireTime = event.ExpireTime
	}
	return event, nil
}

func (r *trafficRepository) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery) (bo.TrafficEventList, error) {
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
	if query.IncludeIndexes {
		if err := r.attachTrafficIndexes(ctx, items); err != nil {
			return bo.TrafficEventList{}, err
		}
	}
	statsRecords, err := r.tableDAO.ListTrafficEventsForStats(ctx, query, eventIDs, trafficStatsLimit)
	if err != nil {
		return bo.TrafficEventList{}, err
	}
	stats := buildTrafficStats(statsRecords)
	stats.Total = total
	return bo.TrafficEventList{
		Items: items,
		Total: total,
		Stats: stats,
	}, nil
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
		EventID:        event.EventID,
		TraceID:        event.TraceID,
		TrafficSource:  event.TrafficSource,
		ProtocolName:   event.ProtocolName,
		NamespaceID:    event.NamespaceID,
		OperationName:  event.OperationName,
		Outcome:        event.Outcome,
		DecisionKind:   event.DecisionKind,
		RuleSetID:      event.RuleSetID,
		RuleID:         event.RuleID,
		SnapshotID:     event.SnapshotID,
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
		EventID:        record.EventID,
		TraceID:        record.TraceID,
		TrafficSource:  record.TrafficSource,
		ProtocolName:   record.ProtocolName,
		NamespaceID:    record.NamespaceID,
		OperationName:  record.OperationName,
		Outcome:        record.Outcome,
		DecisionKind:   record.DecisionKind,
		RuleSetID:      record.RuleSetID,
		RuleID:         record.RuleID,
		SnapshotID:     record.SnapshotID,
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
			EventID:           index.EventID,
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
			EventID:           record.EventID,
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

func buildTrafficStats(records []modeldo.TrafficEvent) bo.TrafficStats {
	stats := emptyTrafficStats()
	for _, record := range records {
		stats.ByOutcome[record.Outcome]++
		stats.ByProtocol[record.ProtocolName]++
		stats.ByNamespace[record.NamespaceID]++
	}
	return stats
}

func emptyTrafficStats() bo.TrafficStats {
	return bo.TrafficStats{
		ByOutcome:   map[string]uint64{},
		ByProtocol:  map[string]uint64{},
		ByNamespace: map[string]uint64{},
	}
}
