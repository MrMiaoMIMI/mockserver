import type {
  Condition,
  NamespaceConfig,
  NamespaceFallbackAction,
  NamespacePolicy,
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

export type NamespaceFallbackFilter = 'all' | 'forward' | 'respond'
export type NamespaceUsageFilter = 'all' | 'used' | 'unused'
export type NamespaceProtocolFilter = 'all' | string
export type NamespaceSortKey = 'name' | 'protocols' | 'usage'

export interface NamespaceEntryFilters {
  query: string
  protocol: NamespaceProtocolFilter
  rulesetMissType: NamespaceFallbackFilter
  ruleMissType: NamespaceFallbackFilter
  usage: NamespaceUsageFilter
  sort: NamespaceSortKey
}

export interface FallbackEntrySummary {
  type: 'forward' | 'respond'
  label: string
  detail: string
  tone: EntryTone
}

export interface NamespaceUsageSummary {
  rulesetCount: number
  ruleCount: number
  rulesets: Array<{ id: string; name: string; protocol: string }>
  rulesetIds: string[]
  rulesetNames: string[]
}

export interface NamespacePolicySummary {
  protocol: string
  policy: NamespacePolicy
  rulesetMiss: FallbackEntrySummary
  ruleMiss: FallbackEntrySummary
  usage: NamespaceUsageSummary
}

export interface NamespaceEntryRow {
  namespace: NamespaceConfig
  displayName: string
  policies: NamespacePolicySummary[]
  protocols: string[]
  usage: NamespaceUsageSummary
  primaryPolicy: NamespacePolicySummary
  searchableText: string
}

export interface NamespaceEntryMetrics {
  total: number
  shown: number
  namespaces: number
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
    protocol: 'all',
    rulesetMissType: 'all',
    ruleMissType: 'all',
    usage: 'all',
    sort: 'name',
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
    ...ruleSet.rules.map((rule) => `${rule.id} ${rule.name || ''} ${ruleActionKind(rule.action)}`),
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

export function buildNamespaceEntryRow(namespace: NamespaceConfig, ruleSets: RuleSet[]): NamespaceEntryRow {
  const policies = namespacePolicyEntries(namespace).map(([protocol, policy]) => {
    const normalizedProtocol = normalizeProtocol(protocol)
    return {
      protocol: normalizedProtocol,
      policy,
      rulesetMiss: fallbackSummary(policy.ruleset_miss_action),
      ruleMiss: fallbackSummary(policy.rule_miss_action),
      usage: namespaceUsage(namespace.name, normalizedProtocol, ruleSets),
    }
  })
  const usage = namespaceUsage(namespace.name, '', ruleSets)
  const protocols = policies.map((policy) => policy.protocol)
  const primaryPolicy = policies[0] || {
    protocol: 'http',
    policy: defaultNamespacePolicy(),
    rulesetMiss: fallbackSummary(defaultNamespacePolicy().ruleset_miss_action),
    ruleMiss: fallbackSummary(defaultNamespacePolicy().rule_miss_action),
    usage: namespaceUsage(namespace.name, 'http', ruleSets),
  }
  const searchableText = [
    namespace.name,
    namespace.description,
    ...protocols,
    ...policies.flatMap((policy) => [
      policy.rulesetMiss.type,
      policy.rulesetMiss.label,
      policy.rulesetMiss.detail,
      policy.ruleMiss.type,
      policy.ruleMiss.label,
      policy.ruleMiss.detail,
      ...policy.usage.rulesetIds,
      ...policy.usage.rulesetNames,
    ]),
    ...usage.rulesetIds,
    ...usage.rulesetNames,
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

  return {
    namespace,
    displayName: namespace.name,
    policies,
    protocols,
    usage,
    primaryPolicy,
    searchableText,
  }
}

function namespacePolicyEntries(namespace: NamespaceConfig): Array<[string, NamespacePolicy]> {
  const policies = new Map<string, NamespacePolicy>()
  Object.entries(namespace.policies || {}).forEach(([protocol, policy]) => {
    const normalizedProtocol = normalizeProtocol(protocol)
    if (normalizedProtocol) policies.set(normalizedProtocol, policy)
  })
  const entries = Array.from(policies.entries()).sort(([left], [right]) => compareText(left, right))
  return entries.length ? entries : [['http', defaultNamespacePolicy()]]
}

function defaultNamespacePolicy(): NamespacePolicy {
  return {
    ruleset_miss_action: { type: 'forward', forward: { timeout_ms: 5000 } },
    rule_miss_action: { type: 'forward', forward: { timeout_ms: 5000 } },
  }
}

export function buildNamespaceEntryMetrics(
  allRows: NamespaceEntryRow[],
  shownRows: NamespaceEntryRow[]
): NamespaceEntryMetrics {
  const shownNamespaceNames = new Set(shownRows.map((row) => row.namespace.name))
  const metrics = allRows.reduce<NamespaceEntryMetrics>(
    (metrics, row) => {
      metrics.total += 1
      if (shownNamespaceNames.has(row.namespace.name)) metrics.shown += 1
      if (row.policies.some((policy) => policy.rulesetMiss.type === 'forward' || policy.ruleMiss.type === 'forward')) {
        metrics.forward += 1
      }
      if (row.policies.some((policy) => policy.rulesetMiss.type === 'respond' || policy.ruleMiss.type === 'respond')) {
        metrics.response += 1
      }
      if (row.usage.rulesetCount > 0) metrics.used += 1
      else metrics.unused += 1
      return metrics
    },
    { total: 0, shown: 0, namespaces: 0, forward: 0, response: 0, used: 0, unused: 0 }
  )
  metrics.namespaces = shownNamespaceNames.size
  return metrics
}

export function fallbackSummary(action?: NamespaceFallbackAction): FallbackEntrySummary {
  if (!action || action.type === 'forward') {
    const timeoutMs = action?.forward?.timeout_ms
    return {
      type: 'forward',
      label: 'Forward',
      detail: timeoutMs ? `original request / ${timeoutMs}ms` : 'original request',
      tone: 'warn',
    }
  }
  const payload = action.response?.payload || {}
  return {
    type: 'respond',
    label: 'Response',
    detail:
      payload.status !== undefined
        ? `HTTP ${payload.status}`
        : payload.code !== undefined
          ? `SPEX ${payload.code}`
          : 'protocol payload',
    tone: 'ok',
  }
}

function ruleActionKind(action: { type?: string; renderer?: string }) {
  return action.renderer || action.type || ''
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
  if (filters.protocol !== 'all' && !row.protocols.includes(normalizeProtocol(filters.protocol))) return false
  if (
    filters.rulesetMissType !== 'all' &&
    !row.policies.some((policy) => policy.rulesetMiss.type === filters.rulesetMissType)
  ) {
    return false
  }
  if (
    filters.ruleMissType !== 'all' &&
    !row.policies.some((policy) => policy.ruleMiss.type === filters.ruleMissType)
  ) {
    return false
  }
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
  const defaultOrder = compareText(left.namespace.name, right.namespace.name)
  if (sort === 'name') return compareText(left.displayName, right.displayName) || defaultOrder
  if (sort === 'protocols') return right.protocols.length - left.protocols.length || defaultOrder
  if (sort === 'usage') return right.usage.rulesetCount - left.usage.rulesetCount || defaultOrder
  return defaultOrder
}

function namespaceUsage(namespaceId: string, protocol: string, ruleSets: RuleSet[]): NamespaceUsageSummary {
  const normalizedProtocol = normalizeProtocol(protocol)
  const related = ruleSets.filter(
    (ruleSet) =>
      ruleSet.namespace === namespaceId && (!normalizedProtocol || normalizeProtocol(ruleSet.protocol) === normalizedProtocol)
  )
  return {
    rulesetCount: related.length,
    ruleCount: related.reduce((count, ruleSet) => count + ruleSet.rules.length, 0),
    rulesets: related.map((ruleSet) => ({
      id: ruleSet.id,
      name: ruleSet.name || ruleSet.id,
      protocol: normalizeProtocol(ruleSet.protocol),
    })),
    rulesetIds: related.map((ruleSet) => ruleSet.id),
    rulesetNames: related.map((ruleSet) => ruleSet.name).filter(Boolean),
  }
}

function normalizeProtocol(protocol: string) {
  return protocol.trim().toLowerCase()
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
