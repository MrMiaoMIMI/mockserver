import type {
  FieldDiffSummary,
  PublishedRuleSetSnapshot,
  RollbackPreviewResponse,
  RollbackRuleSetResponse,
  RuleDiffSummary,
} from '@/types'

export interface SnapshotHistoryItem {
  id: string
  version: string
  publishedAt: string
  operator: string
  reason: string
  sourceSnapshot: string
}

export interface FieldDiffView {
  path: string
  message: string
  current: string
  target: string
}

export interface RuleDiffView {
  ruleId: string
  changeType: string
  changeLabel: string
  message: string
  conditionChanged: boolean
  actionChanged: boolean
  fieldDiffs: FieldDiffView[]
}

export interface RollbackPreviewView {
  changed: boolean
  changedLabel: string
  currentSnapshot: string
  targetSnapshot: string
  messages: string[]
  ruleSetFieldDiffs: FieldDiffView[]
  ruleDiffs: RuleDiffView[]
  hasDiffs: boolean
}

export function buildSnapshotHistory(
  snapshots: PublishedRuleSetSnapshot[]
): SnapshotHistoryItem[] {
  return [...snapshots]
    .sort((left, right) => timestampOf(right.published_at) - timestampOf(left.published_at))
    .map((snapshot) => ({
      id: snapshot.snapshot_id,
      version: snapshot.ruleset.version ? `v${snapshot.ruleset.version}` : 'unversioned',
      publishedAt: formatDateTime(snapshot.published_at),
      operator: snapshot.audit?.operator || 'unknown operator',
      reason: snapshot.audit?.reason || 'no reason',
      sourceSnapshot: snapshot.audit?.source_snapshot_id || '',
    }))
}

export function buildRollbackPreviewView(
  preview: RollbackPreviewResponse | null
): RollbackPreviewView | null {
  if (!preview) return null

  const diff = preview.result.diff
  const ruleSetFieldDiffs = (diff.ruleset_field_diffs || []).map(buildFieldDiffView)
  const ruleDiffs = (diff.rule_diffs || []).map(buildRuleDiffView)
  const hasDiffs = ruleSetFieldDiffs.length > 0 || ruleDiffs.length > 0

  return {
    changed: diff.changed,
    changedLabel: diff.changed ? 'published state will change' : 'target matches current',
    currentSnapshot: diff.current_snapshot_id || 'no current published snapshot',
    targetSnapshot: diff.target_snapshot_id || preview.result.snapshot.snapshot_id,
    messages: diff.messages?.length
      ? diff.messages
      : [diff.changed ? 'Rollback will update the published ruleset.' : 'No structural diff found.'],
    ruleSetFieldDiffs,
    ruleDiffs,
    hasDiffs,
  }
}

export function rollbackResultLabel(result: RollbackRuleSetResponse | null) {
  if (!result) return ''
  const version = result.ruleset.version ? `v${result.ruleset.version}` : 'unversioned'
  return `${version} published as ${result.snapshot.snapshot_id}`
}

export function formatDateTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function buildRuleDiffView(diff: RuleDiffSummary): RuleDiffView {
  return {
    ruleId: diff.rule_id,
    changeType: diff.change_type,
    changeLabel: formatRuleChangeType(diff.change_type),
    message: diff.message || 'rule would change',
    conditionChanged: Boolean(diff.condition_changed),
    actionChanged: Boolean(diff.action_changed),
    fieldDiffs: (diff.field_diffs || []).map(buildFieldDiffView),
  }
}

function buildFieldDiffView(diff: FieldDiffSummary): FieldDiffView {
  return {
    path: diff.path,
    message: diff.message || `field ${diff.path} would change`,
    current: formatDiffValue(diff.current),
    target: formatDiffValue(diff.target),
  }
}

function formatRuleChangeType(type: string) {
  if (type === 'added') return 'added by rollback'
  if (type === 'removed') return 'removed by rollback'
  if (type === 'modified') return 'modified'
  return type || 'changed'
}

function formatDiffValue(value: unknown) {
  if (value === undefined) return 'missing'
  if (value === null) return 'null'
  if (typeof value === 'string') return value || 'empty string'
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  return JSON.stringify(value, null, 2)
}

function timestampOf(value: string) {
  const timestamp = new Date(value).getTime()
  return Number.isNaN(timestamp) ? 0 : timestamp
}
