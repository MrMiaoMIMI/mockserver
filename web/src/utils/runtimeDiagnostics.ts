import type { RuntimeMetrics, RuntimeRequestOutcome, RuntimeRequestRecord } from '@/types'

export type RuntimeOutcomeFilter = 'all' | RuntimeRequestOutcome
export type RuntimeSortKey = 'newest' | 'slowest' | 'status' | 'outcome'
export type RuntimeTone = 'neutral' | 'ok' | 'warn' | 'danger' | 'accent'

export interface RuntimeDiagnosticFilters {
  query: string
  outcome: RuntimeOutcomeFilter
  namespace: string
  sort: RuntimeSortKey
}

export interface RuntimeDiagnosticRow {
  record: RuntimeRequestRecord
  outcome: RuntimeRequestOutcome
  outcomeLabel: string
  tone: RuntimeTone
  statusLabel: string
  durationLabel: string
  pathLabel: string
  observedAtLabel: string
  searchableText: string
  canOpenRuleset: boolean
  canReplay: boolean
}

export interface RuntimeMetricCard {
  label: string
  value: string
  caption: string
  tone: RuntimeTone
}

export interface RuntimeMetricEntry {
  key: string
  value: number
  label: string
}

export function defaultRuntimeDiagnosticFilters(): RuntimeDiagnosticFilters {
  return {
    query: '',
    outcome: 'all',
    namespace: '',
    sort: 'newest',
  }
}

export function buildRuntimeDiagnosticRows(
  metrics: RuntimeMetrics | null | undefined,
  filters: RuntimeDiagnosticFilters
) {
  return (metrics?.recent_requests || [])
    .map(buildRuntimeDiagnosticRow)
    .filter((row) => runtimeRowMatchesFilters(row, filters))
    .sort((left, right) => compareRuntimeRows(left, right, filters.sort))
}

export function buildRuntimeDiagnosticRow(record: RuntimeRequestRecord): RuntimeDiagnosticRow {
  const outcome = runtimeOutcome(record)
  const pathLabel = record.raw_query ? `${record.path || '/'}?${record.raw_query}` : record.path || '/'
  const searchableText = [
    record.id,
    record.source,
    record.protocol,
    record.operation,
    record.namespace,
    record.method,
    record.host,
    record.path,
    record.raw_query,
    record.trace_id,
    record.outcome,
    outcome,
    record.status,
    record.ruleset_id,
    record.rule_id,
    record.fallback_reason,
    record.message,
  ]
    .filter((item) => item !== undefined && item !== null && item !== '')
    .join(' ')
    .toLowerCase()

  return {
    record,
    outcome,
    outcomeLabel: outcomeLabel(outcome),
    tone: outcomeTone(outcome),
    statusLabel: record.status ? String(record.status) : '-',
    durationLabel: formatDuration(record.duration_ms),
    pathLabel,
    observedAtLabel: formatObservedAt(record.observed_at),
    searchableText,
    canOpenRuleset: Boolean(record.ruleset_id),
    canReplay: Boolean(record.event),
  }
}

export function buildRuntimeMetricCards(metrics: RuntimeMetrics | null | undefined): RuntimeMetricCard[] {
  const total = metrics?.total_requests || 0
  const matched = metrics?.matched_requests || 0
  const errors = metrics?.error_requests || 0
  const fallback = sumMetricValues(metrics?.fallback_reasons)
  const matchRate = runtimeMatchRate(metrics)
  return [
    {
      label: 'Total traffic',
      value: formatNumber(total),
      caption: 'runtime requests',
      tone: 'neutral',
    },
    {
      label: 'Match rate',
      value: `${matchRate}%`,
      caption: `${formatNumber(matched)} matched`,
      tone: matchRate >= 80 ? 'ok' : matchRate >= 40 ? 'warn' : 'neutral',
    },
    {
      label: 'Fallback',
      value: formatNumber(fallback),
      caption: 'ruleset or rule miss',
      tone: fallback ? 'warn' : 'ok',
    },
    {
      label: 'Errors',
      value: formatNumber(errors),
      caption: 'runtime failures',
      tone: errors ? 'danger' : 'ok',
    },
    {
      label: 'Avg duration',
      value: formatDuration(metrics?.average_duration_ms || 0),
      caption: 'mean latency',
      tone: 'accent',
    },
  ]
}

export function runtimeNamespaceOptions(metrics: RuntimeMetrics | null | undefined) {
  return Array.from(
    new Set((metrics?.recent_requests || []).map((record) => record.namespace).filter(Boolean))
  ).sort() as string[]
}

export function runtimeMatchRate(metrics: RuntimeMetrics | null | undefined) {
  const total = metrics?.total_requests || 0
  if (!total) return 0
  return Math.round(((metrics?.matched_requests || 0) / total) * 100)
}

export function toSortedMetricEntries(data?: Record<string, number>) {
  return Object.entries(data || {})
    .map(([key, value]) => ({ key, value, label: metricLabel(key) }))
    .sort((left, right) => right.value - left.value || left.label.localeCompare(right.label))
}

export function sumMetricValues(data?: Record<string, number>) {
  return Object.values(data || {}).reduce((sum, value) => sum + value, 0)
}

export function runtimeOutcome(record: RuntimeRequestRecord): RuntimeRequestOutcome {
  if (record.outcome === 'matched' || record.outcome === 'fallback' || record.outcome === 'unmatched' || record.outcome === 'error') {
    return record.outcome
  }
  if (record.error) return 'error'
  if (record.matched) return 'matched'
  if (record.fallback) return 'fallback'
  return 'unmatched'
}

function runtimeRowMatchesFilters(row: RuntimeDiagnosticRow, filters: RuntimeDiagnosticFilters) {
  const query = filters.query.trim().toLowerCase()
  const namespace = filters.namespace.trim().toLowerCase()
  if (query && !row.searchableText.includes(query)) return false
  if (namespace && row.record.namespace?.toLowerCase() !== namespace) return false
  if (filters.outcome !== 'all' && row.outcome !== filters.outcome) return false
  return true
}

function compareRuntimeRows(left: RuntimeDiagnosticRow, right: RuntimeDiagnosticRow, sort: RuntimeSortKey) {
  if (sort === 'slowest') {
    return (right.record.duration_ms || 0) - (left.record.duration_ms || 0)
  }
  if (sort === 'status') {
    return (right.record.status || 0) - (left.record.status || 0) || compareNewest(left, right)
  }
  if (sort === 'outcome') {
    return left.outcomeLabel.localeCompare(right.outcomeLabel) || compareNewest(left, right)
  }
  return compareNewest(left, right)
}

function compareNewest(left: RuntimeDiagnosticRow, right: RuntimeDiagnosticRow) {
  return Date.parse(right.record.observed_at || '') - Date.parse(left.record.observed_at || '')
}

function outcomeLabel(outcome: RuntimeRequestOutcome) {
  if (outcome === 'matched') return 'matched'
  if (outcome === 'fallback') return 'fallback'
  if (outcome === 'error') return 'error'
  return 'unmatched'
}

function outcomeTone(outcome: RuntimeRequestOutcome): RuntimeTone {
  if (outcome === 'matched') return 'ok'
  if (outcome === 'fallback') return 'warn'
  if (outcome === 'error') return 'danger'
  return 'neutral'
}

function metricLabel(value: string) {
  return value.replace(/_/g, ' ')
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('en-US').format(value)
}

function formatDuration(value: number) {
  if (!Number.isFinite(value) || value <= 0) return '0 ms'
  if (value < 1000) return `${value.toFixed(value < 10 ? 1 : 0)} ms`
  return `${(value / 1000).toFixed(2)} s`
}

function formatObservedAt(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
