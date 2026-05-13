package dao

import (
	"context"
	"fmt"

	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"

	modeldo "github.com/MrMiaoMIMI/mockserver/internal/model/do"
	modelfmo "github.com/MrMiaoMIMI/mockserver/internal/model/fmo"
)

type gosharedRuleSetTableDAO struct {
	draftStore     dbspi.SoftDeleteTableStore[*modeldo.RuleSetDraft]
	currentStore   dbspi.SoftDeleteTableStore[*modeldo.PublishedRuleSetCurrent]
	snapshotStore  dbspi.SoftDeleteTableStore[*modeldo.PublishedRuleSetSnapshot]
	draftFields    modelfmo.RuleSetDraftFields
	currentFields  modelfmo.PublishedRuleSetCurrentFields
	snapshotFields modelfmo.PublishedRuleSetSnapshotFields
}

func newGosharedRuleSetTableDAO(manager dbspi.Manager) ruleSetTableDAO {
	return &gosharedRuleSetTableDAO{
		draftStore:     dbhelper.NewSoftDeleteTableStore(&modeldo.RuleSetDraft{}, dbhelper.WithManager(manager)),
		currentStore:   dbhelper.NewSoftDeleteTableStore(&modeldo.PublishedRuleSetCurrent{}, dbhelper.WithManager(manager)),
		snapshotStore:  dbhelper.NewSoftDeleteTableStore(&modeldo.PublishedRuleSetSnapshot{}, dbhelper.WithManager(manager)),
		draftFields:    modelfmo.NewRuleSetDraftFields(),
		currentFields:  modelfmo.NewPublishedRuleSetCurrentFields(),
		snapshotFields: modelfmo.NewPublishedRuleSetSnapshotFields(),
	}
}

func (d *gosharedRuleSetTableDAO) UpsertDraft(ctx context.Context, draft modeldo.RuleSetDraft, expectedVersion int) error {
	if expectedVersion <= 0 {
		if err := d.draftStore.Create(ctx, &draft); err != nil {
			return fmt.Errorf("insert draft %s: %w", draft.RuleSetCode, err)
		}
		return nil
	}

	updater := dbhelper.NewUpdater().
		Set(d.draftFields.Version, draft.Version).
		Set(d.draftFields.RuleSetJSON, draft.RuleSetJSON)
	query := dbhelper.Q(
		d.draftFields.RuleSetCode.Eq(&draft.RuleSetCode),
		d.draftFields.Version.Eq(&expectedVersion),
	)
	if err := d.draftStore.UpdateByQuery(ctx, query, updater); err != nil {
		return fmt.Errorf("update draft %s: %w", draft.RuleSetCode, err)
	}
	updated, ok, err := d.GetDraft(ctx, draft.RuleSetCode)
	if err != nil {
		return err
	}
	if !ok || updated.Version != draft.Version {
		return ErrConflict
	}
	return nil
}

func (d *gosharedRuleSetTableDAO) GetDraft(ctx context.Context, id string) (modeldo.RuleSetDraft, bool, error) {
	query := dbhelper.Q(d.draftFields.RuleSetCode.Eq(&id))
	exists, draft, err := d.draftStore.ExistsNotDeleted(ctx, query)
	if err != nil {
		return modeldo.RuleSetDraft{}, false, fmt.Errorf("get draft %s: %w", id, err)
	}
	if !exists || draft == nil {
		return modeldo.RuleSetDraft{}, false, nil
	}
	return *draft, true, nil
}

func (d *gosharedRuleSetTableDAO) ListDrafts(ctx context.Context) ([]modeldo.RuleSetDraft, error) {
	items, err := d.draftStore.FindNotDeleted(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}
	result := make([]modeldo.RuleSetDraft, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (d *gosharedRuleSetTableDAO) PublishSnapshot(ctx context.Context, snapshot modeldo.PublishedRuleSetSnapshot) (modeldo.PublishedRuleSetSnapshot, error) {
	if err := d.snapshotStore.Create(ctx, &snapshot); err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("insert published snapshot %s: %w", snapshot.SnapshotCode, err)
	}

	query := dbhelper.Q(d.currentFields.RuleSetID.Eq(&snapshot.RuleSetID))
	exists, _, err := d.currentStore.Exists(ctx, query)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("get current published %d: %w", snapshot.RuleSetID, err)
	}
	if !exists {
		current := &modeldo.PublishedRuleSetCurrent{
			RuleSetID:         snapshot.RuleSetID,
			CurrentSnapshotID: snapshot.Id,
		}
		if err := d.currentStore.Create(ctx, current); err != nil {
			return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("insert current published %d: %w", snapshot.RuleSetID, err)
		}
		return snapshot, nil
	}

	updater := dbhelper.NewUpdater().
		Set(d.currentFields.CurrentSnapshotID, snapshot.Id)
	if err := d.currentStore.UpdateByQuery(ctx, query, updater); err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, fmt.Errorf("update current published %d: %w", snapshot.RuleSetID, err)
	}
	return snapshot, nil
}

func (d *gosharedRuleSetTableDAO) GetCurrentPublished(ctx context.Context, id string) (modeldo.PublishedRuleSetSnapshot, bool, error) {
	draft, ok, err := d.GetDraft(ctx, id)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, err
	}
	if !ok {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	query := dbhelper.Q(d.currentFields.RuleSetID.Eq(&draft.Id))
	exists, current, err := d.currentStore.ExistsNotDeleted(ctx, query)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, fmt.Errorf("get current published %s: %w", id, err)
	}
	if !exists || current == nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	exists, snapshot, err := d.snapshotStore.ExistsByIdNotDeleted(ctx, current.CurrentSnapshotID)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, fmt.Errorf("get current snapshot %d: %w", current.CurrentSnapshotID, err)
	}
	if !exists || snapshot == nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	return *snapshot, true, nil
}

func (d *gosharedRuleSetTableDAO) GetPublishedSnapshot(ctx context.Context, id string, snapshotID string) (modeldo.PublishedRuleSetSnapshot, bool, error) {
	draft, ok, err := d.GetDraft(ctx, id)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, err
	}
	if !ok {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	query := dbhelper.Q(
		d.snapshotFields.RuleSetID.Eq(&draft.Id),
		d.snapshotFields.SnapshotCode.Eq(&snapshotID),
	)
	exists, snapshot, err := d.snapshotStore.ExistsNotDeleted(ctx, query)
	if err != nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, fmt.Errorf("get published snapshot %s: %w", snapshotID, err)
	}
	if !exists || snapshot == nil {
		return modeldo.PublishedRuleSetSnapshot{}, false, nil
	}
	return *snapshot, true, nil
}

func (d *gosharedRuleSetTableDAO) ListPublished(ctx context.Context) ([]modeldo.PublishedRuleSetSnapshot, error) {
	currents, err := d.currentStore.FindNotDeleted(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list current published: %w", err)
	}
	items := make([]modeldo.PublishedRuleSetSnapshot, 0, len(currents))
	for _, current := range currents {
		if current == nil {
			continue
		}
		exists, snapshot, err := d.snapshotStore.ExistsByIdNotDeleted(ctx, current.CurrentSnapshotID)
		if err != nil {
			return nil, err
		}
		if exists && snapshot != nil {
			items = append(items, *snapshot)
		}
	}
	return items, nil
}

func (d *gosharedRuleSetTableDAO) ListPublishedSnapshots(ctx context.Context, id string) ([]modeldo.PublishedRuleSetSnapshot, error) {
	draft, ok, err := d.GetDraft(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	query := dbhelper.Q(d.snapshotFields.RuleSetID.Eq(&draft.Id))
	items, err := d.snapshotStore.FindNotDeleted(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("list published snapshots %s: %w", id, err)
	}
	result := make([]modeldo.PublishedRuleSetSnapshot, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}
