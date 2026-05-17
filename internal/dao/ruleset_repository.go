package dao

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

type ruleSetRepository struct {
	tableDAO          ruleSetTableDAO
	publishedRevision uint64
}

func newRuleSetRepository(tableDAO ruleSetTableDAO) RuleSetRepository {
	return &ruleSetRepository{tableDAO: tableDAO}
}

func (r *ruleSetRepository) UpsertDraft(ctx context.Context, ruleSet bo.RuleSet) (bo.RuleSet, error) {
	current, exists, err := r.tableDAO.GetDraft(ctx, ruleSet.ID)
	if err != nil {
		return bo.RuleSet{}, err
	}

	expectedVersion := 0
	if exists {
		expectedVersion = current.Version
		ruleSet.Version = current.Version + 1
	} else {
		ruleSet.Version = 1
	}

	raw, err := json.Marshal(ruleSet)
	if err != nil {
		return bo.RuleSet{}, fmt.Errorf("marshal draft ruleset %s: %w", ruleSet.ID, err)
	}
	if err := r.tableDAO.UpsertDraft(ctx, modeldo.RuleSetDraft{
		RuleSetCode:   ruleSet.ID,
		RuleSetName:   ruleSet.Name,
		ProtocolName:  ruleSet.Protocol,
		NamespaceCode: ruleSet.Namespace,
		Version:       ruleSet.Version,
		RuleSetJSON:   string(raw),
	}, expectedVersion); err != nil {
		return bo.RuleSet{}, err
	}
	if record, ok, err := r.tableDAO.GetDraft(ctx, ruleSet.ID); err != nil {
		return bo.RuleSet{}, err
	} else if ok {
		ruleSet.DBID = record.Id
	}
	return ruleSet, nil
}

func (r *ruleSetRepository) GetDraft(ctx context.Context, id string) (bo.RuleSet, bool, error) {
	record, ok, err := r.tableDAO.GetDraft(ctx, id)
	if err != nil || !ok {
		return bo.RuleSet{}, false, err
	}
	ruleSet, err := decodeRuleSetRecord(record)
	if err != nil {
		return bo.RuleSet{}, false, err
	}
	return ruleSet, true, nil
}

func (r *ruleSetRepository) ListDrafts(ctx context.Context) ([]bo.RuleSet, error) {
	records, err := r.tableDAO.ListDrafts(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]bo.RuleSet, 0, len(records))
	for _, record := range records {
		ruleSet, err := decodeRuleSetRecord(record)
		if err != nil {
			return nil, err
		}
		items = append(items, ruleSet)
	}
	return items, nil
}

func (r *ruleSetRepository) Publish(ctx context.Context, id string, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	ctx = contextWithAuditOperator(ctx, audit)
	ruleSet, ok, err := r.GetDraft(ctx, id)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	if !ok {
		return bo.PublishedRuleSetSnapshot{}, fmt.Errorf("ruleset %s not found", id)
	}
	return r.publishSnapshot(ctx, ruleSet, audit)
}

func (r *ruleSetRepository) GetPublished(ctx context.Context, id string) (bo.PublishedRuleSetSnapshot, bool, error) {
	record, ok, err := r.tableDAO.GetCurrentPublished(ctx, id)
	if err != nil || !ok {
		return bo.PublishedRuleSetSnapshot{}, false, err
	}
	snapshot, err := decodePublishedSnapshotRecord(record)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (r *ruleSetRepository) GetPublishedSnapshot(ctx context.Context, id, snapshotID string) (bo.PublishedRuleSetSnapshot, bool, error) {
	record, ok, err := r.tableDAO.GetPublishedSnapshot(ctx, id, snapshotID)
	if err != nil || !ok {
		return bo.PublishedRuleSetSnapshot{}, false, err
	}
	snapshot, err := decodePublishedSnapshotRecord(record)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (r *ruleSetRepository) Rollback(ctx context.Context, id, snapshotID string, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	ctx = contextWithAuditOperator(ctx, audit)
	snapshot, ok, err := r.GetPublishedSnapshot(ctx, id, snapshotID)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	if !ok {
		return bo.PublishedRuleSetSnapshot{}, fmt.Errorf("snapshot %s for ruleset %s not found", snapshotID, id)
	}
	return r.publishSnapshot(ctx, snapshot.RuleSet, audit)
}

func (r *ruleSetRepository) ListPublished(ctx context.Context) ([]bo.PublishedRuleSetSnapshot, error) {
	records, err := r.tableDAO.ListPublished(ctx)
	if err != nil {
		return nil, err
	}
	return decodePublishedSnapshotRecords(records)
}

func (r *ruleSetRepository) ListPublishedSnapshots(ctx context.Context, id string) ([]bo.PublishedRuleSetSnapshot, error) {
	records, err := r.tableDAO.ListPublishedSnapshots(ctx, id)
	if err != nil {
		return nil, err
	}
	return decodePublishedSnapshotRecords(records)
}

func (r *ruleSetRepository) publishSnapshot(ctx context.Context, ruleSet bo.RuleSet, audit *bo.AuditInfo) (bo.PublishedRuleSetSnapshot, error) {
	publishedAt := time.Now().UTC()
	snapshot := bo.PublishedRuleSetSnapshot{
		SnapshotID:  newSnapshotID(publishedAt),
		PublishedAt: publishedAt,
		RuleSet:     ruleSet,
		Audit:       cloneAuditInfo(audit),
	}
	record, err := encodePublishedSnapshotRecord(snapshot)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	record, err = r.tableDAO.PublishSnapshot(ctx, record)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	snapshot.DBID = record.Id
	atomic.AddUint64(&r.publishedRevision, 1)
	return snapshot, nil
}

func (r *ruleSetRepository) PublishedRevision() uint64 {
	return atomic.LoadUint64(&r.publishedRevision)
}

func decodeRuleSetRecord(record modeldo.RuleSetDraft) (bo.RuleSet, error) {
	ruleSet, err := decodeRuleSetJSON(record.RuleSetJSON)
	if err != nil {
		return bo.RuleSet{}, err
	}
	ruleSet.DBID = record.Id
	if ruleSet.ID == "" {
		ruleSet.ID = record.RuleSetCode
	}
	if ruleSet.Name == "" {
		ruleSet.Name = record.RuleSetName
	}
	if ruleSet.Protocol == "" {
		ruleSet.Protocol = record.ProtocolName
	}
	if ruleSet.Namespace == "" {
		ruleSet.Namespace = record.NamespaceCode
	}
	return ruleSet, nil
}

func decodeRuleSetJSON(raw string) (bo.RuleSet, error) {
	var ruleSet bo.RuleSet
	if err := json.Unmarshal([]byte(raw), &ruleSet); err != nil {
		return bo.RuleSet{}, fmt.Errorf("decode ruleset json: %w", err)
	}
	return ruleSet, nil
}

func encodePublishedSnapshotRecord(snapshot bo.PublishedRuleSetSnapshot) (modeldo.PublishedRuleSetSnapshot, error) {
	if snapshot.RuleSet.DBID == 0 {
		return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("ruleset %s missing internal id", snapshot.RuleSet.ID)
	}
	ruleSetRaw, err := json.Marshal(snapshot.RuleSet)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("marshal snapshot ruleset %s: %w", snapshot.SnapshotID, err)
	}
	var auditRaw string
	if snapshot.Audit != nil {
		raw, err := json.Marshal(snapshot.Audit)
		if err != nil {
			return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("marshal snapshot audit %s: %w", snapshot.SnapshotID, err)
		}
		auditRaw = string(raw)
	}
	return modeldo.PublishedRuleSetSnapshot{
		SnapshotCode:   snapshot.SnapshotID,
		RuleSetID:      snapshot.RuleSet.DBID,
		RuleSetVersion: snapshot.RuleSet.Version,
		RuleSetJSON:    string(ruleSetRaw),
		AuditJSON:      auditRaw,
		PublishTime:    uint64(snapshot.PublishedAt.UnixMilli()),
	}, nil
}

func decodePublishedSnapshotRecord(record modeldo.PublishedRuleSetSnapshot) (bo.PublishedRuleSetSnapshot, error) {
	ruleSet, err := decodeRuleSetJSON(record.RuleSetJSON)
	if err != nil {
		return bo.PublishedRuleSetSnapshot{}, err
	}
	ruleSet.DBID = record.RuleSetID
	var audit *bo.AuditInfo
	if record.AuditJSON != "" {
		var item bo.AuditInfo
		if err := json.Unmarshal([]byte(record.AuditJSON), &item); err != nil {
			return bo.PublishedRuleSetSnapshot{}, fmt.Errorf("decode snapshot audit %s: %w", record.SnapshotCode, err)
		}
		audit = &item
	}
	return bo.PublishedRuleSetSnapshot{
		DBID:        record.Id,
		SnapshotID:  record.SnapshotCode,
		PublishedAt: time.UnixMilli(int64(record.PublishTime)).UTC(),
		RuleSet:     ruleSet,
		Audit:       audit,
	}, nil
}

func decodePublishedSnapshotRecords(records []modeldo.PublishedRuleSetSnapshot) ([]bo.PublishedRuleSetSnapshot, error) {
	items := make([]bo.PublishedRuleSetSnapshot, 0, len(records))
	for _, record := range records {
		snapshot, err := decodePublishedSnapshotRecord(record)
		if err != nil {
			return nil, err
		}
		items = append(items, snapshot)
	}
	return items, nil
}

func newSnapshotID(now time.Time) string {
	var data [8]byte
	if _, err := rand.Read(data[:]); err == nil {
		return "snap_" + hex.EncodeToString(data[:])
	}
	return fmt.Sprintf("snap_%d", now.UnixMilli())
}

func cloneAuditInfo(audit *bo.AuditInfo) *bo.AuditInfo {
	if audit == nil {
		return nil
	}
	cloned := *audit
	return &cloned
}

func contextWithAuditOperator(ctx context.Context, audit *bo.AuditInfo) context.Context {
	if operator, ok := dbspi.OperatorFromContext(ctx); ok && strings.TrimSpace(operator) != "" {
		return ctx
	}
	if audit == nil {
		return ctx
	}
	operator := strings.TrimSpace(audit.Operator)
	if operator == "" {
		return ctx
	}
	return dbspi.WithOperator(ctx, operator)
}
