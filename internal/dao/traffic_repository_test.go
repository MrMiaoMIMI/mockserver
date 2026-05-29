package dao

import (
	"context"
	"strings"
	"testing"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
)

type fakeTrafficTableDAO struct {
	listQuery  bo.TrafficQuery
	statsQuery bo.TrafficQuery
	stats      bo.TrafficStats
}

func (d *fakeTrafficTableDAO) CreateTrafficEvent(ctx context.Context, event modeldo.TrafficEvent, indexes []modeldo.TrafficEventIndex) (modeldo.TrafficEvent, error) {
	_, _ = ctx, indexes
	event.Id = 1
	return event, nil
}

func (d *fakeTrafficTableDAO) GetTrafficEvent(ctx context.Context, id uint64) (modeldo.TrafficEvent, bool, error) {
	_, _ = ctx, id
	return modeldo.TrafficEvent{}, false, nil
}

func (d *fakeTrafficTableDAO) ListTrafficEvents(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64) ([]modeldo.TrafficEvent, uint64, error) {
	_, _ = ctx, eventIDs
	d.listQuery = query
	return []modeldo.TrafficEvent{{EventCode: "te_1", Outcome: bo.TrafficOutcomeMatched}}, 3, nil
}

func (d *fakeTrafficTableDAO) ListTrafficEventIndexesByEventIDs(ctx context.Context, eventIDs []uint64) (map[uint64][]modeldo.TrafficEventIndex, error) {
	_, _ = ctx, eventIDs
	return nil, nil
}

func (d *fakeTrafficTableDAO) ListTrafficEventIDsByIndexFilter(ctx context.Context, query bo.TrafficQuery, filter bo.TrafficIndexFilter, valueHash uint64) ([]uint64, error) {
	_ = ctx
	_ = query
	_ = filter
	_ = valueHash
	return nil, nil
}

func (d *fakeTrafficTableDAO) TrafficStats(ctx context.Context, query bo.TrafficQuery, eventIDs []uint64) (bo.TrafficStats, error) {
	_, _ = ctx, eventIDs
	d.statsQuery = query
	return d.stats, nil
}

func TestTrafficRepositoryUsesExactStatsAndPreservesRulesetCodeFilter(t *testing.T) {
	table := &fakeTrafficTableDAO{
		stats: bo.TrafficStats{
			ByOutcome:   map[string]uint64{bo.TrafficOutcomeMatched: 2, bo.TrafficOutcomeFallback: 1},
			ByProtocol:  map[string]uint64{"http": 3},
			ByNamespace: map[string]uint64{"default": 3},
		},
	}
	ruleSetTable := newFakeRuleSetTableDAO()
	draft := modeldo.RuleSetDraft{RuleSetCode: "rs-api"}
	draft.Id = 7
	ruleSetTable.drafts["rs-api"] = draft
	repository := newTrafficRepository(table, nil, ruleSetTable)

	list, err := repository.ListTrafficEvents(context.Background(), bo.TrafficQuery{RuleSetID: "rs-api"})
	if err != nil {
		t.Fatalf("ListTrafficEvents() error = %v", err)
	}
	if list.Total != 3 || list.Stats.Total != 3 || list.Stats.ByOutcome[bo.TrafficOutcomeMatched] != 2 {
		t.Fatalf("expected exact stats with total from count, got %+v", list)
	}
	if table.statsQuery.RuleSetDBID != 7 || table.statsQuery.RuleSetID != "rs-api" {
		t.Fatalf("expected stats query to preserve db id and stable system id, got %+v", table.statsQuery)
	}
}

func TestGosharedTrafficWhereSQLUsesCodeOrDBIDForHistoricalQueries(t *testing.T) {
	dao := &gosharedTrafficTableDAO{}
	where, args := dao.buildEventWhereSQL(bo.TrafficQuery{
		NamespaceID:   "shop",
		NamespaceDBID: 11,
		RuleSetID:     "rs-api",
		RuleSetDBID:   7,
	}, nil)

	if !strings.Contains(where, "(namespace_id = ? OR namespace_name = ?)") || !strings.Contains(where, "(ruleset_id = ? OR ruleset_code = ?)") {
		t.Fatalf("expected db id or code filters, where=%s", where)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d: %+v", len(args), args)
	}
}
