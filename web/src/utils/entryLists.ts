import type {
  Condition,
  NamespaceConfig,
  NamespaceFallbackAction,
  PublishedRuleSetSnapshot,
  RuleSet,
  Selector,
} from '@/types'

export type EntryTone = 'neutral' | 'ok' | 'warn' | 'danger' | 'accent'
export type RulesetPublishState = 'draft_only' | 'published_current' | 'changed'
export type RulesetEnabledFilter = 'all' | 'enabled' | 'disabled'
export type RulesetPublishFilter = 'all' | RulesetPublishState
export type RulesetRulePresenceFilter = 'all' | 'with_rules' | 'empty'
export type RulesetSortKey = 'name' | 'namespace' | 'rule_count' | 'publish_state' | 'version'

export interface RulesetEntryFilters {
  query: string
  namespace: string
  enabled: RulesetEnabledFilter
  publishState: RulesetPublishFilter
  rulePresence: RulesetRulePresenceFilter
  sort: RulesetSortKey
}

export interface RulesetEntryRow {
  ruleSet: RuleSet
  publishState: RulesetPublishState
  publishLabel: string
  publishTone: EntryTone
  enabledLabel: string
  enabledTone: EntryTone
  selectorTotal: number
  selectorValues: string[]
  ruleCount: number
  versionLabel: string
  searchableText: string
}

export interface RulesetEntryMetrics {
  total: number
  shown: number
  published: number
  changed: number
  draftOnly: number
  enabled: number
  empty: number
}

export type NamespaceFallbackFilter = 'all' | 'forward' | 'response'
export type NamespaceUsageFilter = 'all' | 'used' | 'unused'
export type NamespaceSortKey = 'id' | 'usage' | 'ruleset_miss' | 'rule_miss'

export interface NamespaceEntryFilters {
  query: string
  rulesetMissType: NamespaceFallbackFilter
  ruleMissType: NamespaceFallbackFilter
  usage: NamespaceUsageFilter
  sort: NamespaceSortKey
}

export interface FallbackEntrySummary {
  type: 'forward' | 'response'
  label: string
  detail: string
  tone: EntryTone
}

export interface NamespaceUsageSummary {
  rulesetCount: number
  ruleCount: number
  rulesets: Array<{ id: string; name: string }>
  rulesetIds: string[]
  rulesetNames: string[]
}

export interface NamespaceEntryRow {
  namespace: NamespaceConfig
  displayName: string
  rulesetMiss: FallbackEntrySummary
  ruleMiss: FallbackEntrySummary
  usage: NamespaceUsageSummary
  searchableText: string
}

export interface NamespaceEntryMetrics {
  total: number
  shown: number
  forward: number
  response: number
  used: number
  unused: number
}

export function defaultRulesetEntryFilters(): RulesetEntryFilters {
  return {
    query: '',
    namespace: '',
    enabled: 'all',
    publishState: 'all',
    rulePresence: 'all',
    sort: 'name',
  }
}

export function defaultNamespaceEntryFilters(): NamespaceEntryFilters {
  return {
    query: '',
    rulesetMissType: 'all',
    ruleMissType: 'all',
    usage: 'all',
    sort: 'id',
  }
}

export function buildRulesetEntryRows(
  drafts: RuleSet[],
  published: PublishedRuleSetSnapshot[],
  filters: RulesetEntryFilters
) {
  const publishedMap = new Map(published.map((item) => [item.ruleset.id, item]))
  return drafts
    .map((ruleSet) => buildRulesetEntryRow(ruleSet, publishedMap.get(ruleSet.id) || null))
    .filter((row) => rulesetEntryMatchesFilters(row, filters))
    .sort((left, right) => compareRulesetEntryRows(left, right, filters.sort))
}

export function buildRulesetEntryRow(
  ruleSet: RuleSet,
  published: PublishedRuleSetSnapshot | null
): RulesetEntryRow {
  const publishState = rulesetPublishState(ruleSet, published)
  const selectorValues = selectorValueSummaries(ruleSet.selector)
  const ruleCount = ruleSet.rules.length
  const searchableText = [
    ruleSet.id,
    ruleSet.name,
    ruleSet.protocol,
    ruleSet.namespace,
    publishState,
    publishStateLabel(publishState),
    ruleSet.enabled ? 'enabled' : 'disabled',
    ...selectorValues,
    ...ruleSet.rules.map((rule) => `${rule.id} ${rule.name || ''} ${rule.action.type}`),
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

  return {
    ruleSet,
    publishState,
    publishLabel: publishStateLabel(publishState),
    publishTone: publishStateTone(publishState),
    enabledLabel: ruleSet.enabled ? 'enabled' : 'disabled',
    enabledTone: ruleSet.enabled ? 'ok' : 'neutral',
    selectorTotal: selectorValues.length,
    selectorValues,
    ruleCount,
    versionLabel: ruleSet.version ? `v${ruleSet.version}` : 'unversioned',
    searchableText,
  }
}

export function selectorValueSummaries(selector?: Selector): string[] {
  return (selector?.all || []).map(selectorConditionSummary).filter(Boolean)
}

export function buildRulesetEntryMetrics(allRows: RulesetEntryRow[], shownRows: RulesetEntryRow[]) {
  const shownIds = new Set(shownRows.map((row) => row.ruleSet.id))
  return allRows.reduce<RulesetEntryMetrics>(
    (metrics, row) => {
      metrics.total += 1
      if (shownIds.has(row.ruleSet.id)) metrics.shown += 1
      if (row.publishState !== 'draft_only') metrics.published += 1
      if (row.publishState === 'changed') metrics.changed += 1
      if (row.publishState === 'draft_only') metrics.draftOnly += 1
      if (row.ruleSet.enabled) metrics.enabled += 1
      if (row.ruleCount === 0) metrics.empty += 1
      return metrics
    },
    { total: 0, shown: 0, published: 0, changed: 0, draftOnly: 0, enabled: 0, empty: 0 }
  )
}

export function buildNamespaceEntryRows(
  namespaces: NamespaceConfig[],
  ruleSets: RuleSet[],
  filters: NamespaceEntryFilters
) {
  return namespaces
    .map((namespace) => buildNamespaceEntryRow(namespace, ruleSets))
    .filter((row) => namespaceEntryMatchesFilters(row, filters))
    .sort((left, right) => compareNamespaceEntryRows(left, right, filters.sort))
}

export function buildNamespaceEntryRow(
  namespace: NamespaceConfig,
  ruleSets: RuleSet[]
): NamespaceEntryRow {
  const usage = namespaceUsage(namespace.id, ruleSets)
  const rulesetMiss = fallbackSummary(namespace.ruleset_miss_action)
  const ruleMiss = fallbackSummary(namespace.rule_miss_action)
  const searchableText = [
    namespace.id,
    namespace.name,
    namespace.description,
    rulesetMiss.type,
    rulesetMiss.label,
    rulesetMiss.detail,
    ruleMiss.type,
    ruleMiss.label,
    ruleMiss.detail,
    ...usage.rulesetIds,
    ...usage.rulesetNames,
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

  return {
    namespace,
    displayName: namespace.name || namespace.id,
    rulesetMiss,
    ruleMiss,
    usage,
    searchableText,
  }
}

export function buildNamespaceEntryMetrics(
  allRows: NamespaceEntryRow[],
  shownRows: NamespaceEntryRow[]
): NamespaceEntryMetrics {
  const shownIds = new Set(shownRows.map((row) => row.namespace.id))
  return allRows.reduce<NamespaceEntryMetrics>(
    (metrics, row) => {
      metrics.total += 1
      if (shownIds.has(row.namespace.id)) metrics.shown += 1
      if (row.rulesetMiss.type === 'forward' || row.ruleMiss.type === 'forward') metrics.forward += 1
      if (row.rulesetMiss.type === 'response' || row.ruleMiss.type === 'response') metrics.response += 1
      if (row.usage.rulesetCount > 0) metrics.used += 1
      else metrics.unused += 1
      return metrics
    },
    { total: 0, shown: 0, forward: 0, response: 0, used: 0, unused: 0 }
  )
}

export function fallbackSummary(action: NamespaceFallbackAction): FallbackEntrySummary {
  if (action.type === 'forward') {
    return {
      type: 'forward',
      label: 'Forward',
      detail: action.forward?.timeout_ms
        ? `original request / ${action.forward.timeout_ms}ms`
        : 'original request',
      tone: 'warn',
    }
  }
  return {
    type: 'response',
    label: 'Response',
    detail: `HTTP ${action.response?.status || 404}`,
    tone: 'ok',
  }
}

export function rulesetPublishState(
  ruleSet: RuleSet,
  published: PublishedRuleSetSnapshot | null
): RulesetPublishState {
  if (!published) return 'draft_only'
  return stableJSON(withoutVersion(ruleSet)) === stableJSON(withoutVersion(published.ruleset))
    ? 'published_current'
    : 'changed'
}

function rulesetEntryMatchesFilters(row: RulesetEntryRow, filters: RulesetEntryFilters) {
  const query = filters.query.trim().toLowerCase()
  const namespace = filters.namespace.trim().toLowerCase()
  if (query && !row.searchableText.includes(query)) return false
  if (namespace && !row.ruleSet.namespace.toLowerCase().includes(namespace)) return false
  if (filters.enabled === 'enabled' && !row.ruleSet.enabled) return false
  if (filters.enabled === 'disabled' && row.ruleSet.enabled) return false
  if (filters.publishState !== 'all' && row.publishState !== filters.publishState) return false
  if (filters.rulePresence === 'with_rules' && row.ruleCount === 0) return false
  if (filters.rulePresence === 'empty' && row.ruleCount > 0) return false
  return true
}

function namespaceEntryMatchesFilters(row: NamespaceEntryRow, filters: NamespaceEntryFilters) {
  const query = filters.query.trim().toLowerCase()
  if (query && !row.searchableText.includes(query)) return false
  if (filters.rulesetMissType !== 'all' && row.rulesetMiss.type !== filters.rulesetMissType) return false
  if (filters.ruleMissType !== 'all' && row.ruleMiss.type !== filters.ruleMissType) return false
  if (filters.usage === 'used' && row.usage.rulesetCount === 0) return false
  if (filters.usage === 'unused' && row.usage.rulesetCount > 0) return false
  return true
}

function compareRulesetEntryRows(left: RulesetEntryRow, right: RulesetEntryRow, sort: RulesetSortKey) {
  if (sort === 'namespace') return compareText(left.ruleSet.namespace, right.ruleSet.namespace)
  if (sort === 'rule_count') return right.ruleCount - left.ruleCount || compareText(left.ruleSet.name, right.ruleSet.name)
  if (sort === 'publish_state') return compareText(left.publishLabel, right.publishLabel)
  if (sort === 'version') return (right.ruleSet.version || 0) - (left.ruleSet.version || 0)
  return compareText(left.ruleSet.name || left.ruleSet.id, right.ruleSet.name || right.ruleSet.id)
}

function compareNamespaceEntryRows(left: NamespaceEntryRow, right: NamespaceEntryRow, sort: NamespaceSortKey) {
  if (sort === 'usage') return right.usage.rulesetCount - left.usage.rulesetCount || compareText(left.namespace.id, right.namespace.id)
  if (sort === 'ruleset_miss') return compareText(left.rulesetMiss.type, right.rulesetMiss.type)
  if (sort === 'rule_miss') return compareText(left.ruleMiss.type, right.ruleMiss.type)
  return compareText(left.namespace.id, right.namespace.id)
}

function namespaceUsage(namespaceId: string, ruleSets: RuleSet[]): NamespaceUsageSummary {
  const related = ruleSets.filter((ruleSet) => ruleSet.namespace === namespaceId)
  return {
    rulesetCount: related.length,
    ruleCount: related.reduce((count, ruleSet) => count + ruleSet.rules.length, 0),
    rulesets: related.map((ruleSet) => ({ id: ruleSet.id, name: ruleSet.name || ruleSet.id })),
    rulesetIds: related.map((ruleSet) => ruleSet.id),
    rulesetNames: related.map((ruleSet) => ruleSet.name).filter(Boolean),
  }
}

function publishStateLabel(state: RulesetPublishState) {
  if (state === 'published_current') return 'published current'
  if (state === 'changed') return 'draft changed'
  return 'draft only'
}

function publishStateTone(state: RulesetPublishState): EntryTone {
  if (state === 'published_current') return 'ok'
  if (state === 'changed') return 'warn'
  return 'neutral'
}

function withoutVersion(ruleSet: RuleSet) {
  const { version: _version, ...rest } = ruleSet
  return rest
}

function stableJSON(value: unknown): string {
  return JSON.stringify(sortValue(value))
}

function selectorConditionSummary(condition: Condition): string {
  if (condition.field && condition.op) {
    if (['exists', 'not_exists', 'is_null', 'is_not_null'].includes(condition.op)) {
      return `${condition.field} ${condition.op}`
    }
    return `${condition.field} ${condition.op} ${formatSelectorValue(condition.value)}`
  }
  if (condition.all?.length) return `all(${condition.all.map(selectorConditionSummary).join(', ')})`
  if (condition.any?.length) return `any(${condition.any.map(selectorConditionSummary).join(', ')})`
  if (condition.not) return `not(${selectorConditionSummary(condition.not)})`
  return ''
}

function formatSelectorValue(value: unknown): string {
  if (value === undefined) return ''
  if (typeof value === 'string') return value
  return JSON.stringify(value)
}

function sortValue(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sortValue)
  if (!value || typeof value !== 'object') return value
  return Object.keys(value as Record<string, unknown>)
    .sort()
    .reduce<Record<string, unknown>>((result, key) => {
      result[key] = sortValue((value as Record<string, unknown>)[key])
      return result
    }, {})
}

function compareText(left: string, right: string) {
  return left.localeCompare(right)
}
