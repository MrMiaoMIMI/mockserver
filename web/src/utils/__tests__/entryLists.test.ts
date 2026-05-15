import { describe, expect, it } from 'vitest'
import type { NamespaceConfig, PublishedRuleSetSnapshot, RuleSet } from '@/types'
import {
  buildNamespaceEntryMetrics,
  buildNamespaceEntryRows,
  buildRulesetEntryMetrics,
  buildRulesetEntryRows,
  defaultNamespaceEntryFilters,
  defaultRulesetEntryFilters,
  fallbackSummary,
  rulesetPublishState,
} from '@/utils/entryLists'

const draftCurrent: RuleSet = {
  id: 'rs-current',
  name: 'Current',
  enabled: true,
  protocol: 'http',
  namespace: 'default',
  selector: {
    all: [
      { field: 'request.host', op: 'eq', value: 'a.test' },
      { field: 'request.path', op: 'prefix', value: '/api' },
    ],
  },
  rules: [
    {
      id: 'rule-a',
      name: 'Rule A',
      enabled: true,
      priority: 10,
      when: { field: 'request.path', op: 'eq', value: '/api/a' },
      action: { type: 'respond', renderer: 'static', response: { payload: { status: 200 } } },
    },
  ],
  version: 2,
}

const draftChanged: RuleSet = {
  ...draftCurrent,
  id: 'rs-changed',
  name: 'Changed',
  namespace: 'payments',
  selector: {
    all: [
      { field: 'request.host', op: 'eq', value: 'pay.test' },
      { field: 'request.path', op: 'prefix', value: '/pay' },
    ],
  },
  version: 5,
}

const draftSpex: RuleSet = {
  ...draftCurrent,
  id: 'rs-spex',
  name: 'SPEX Changed',
  protocol: 'spex',
  namespace: 'payments',
  selector: {
    all: [{ field: 'request.service', op: 'eq', value: 'payment.spex' }],
  },
  rules: [
    {
      id: 'rule-spex',
      name: 'Rule SPEX',
      enabled: true,
      priority: 10,
      when: { field: 'request.method', op: 'eq', value: 'GetPayment' },
      action: { type: 'respond', renderer: 'static', response: { payload: { code: 0, resp: {} } } },
    },
  ],
  version: 1,
}

const draftOnly: RuleSet = {
  ...draftCurrent,
  id: 'rs-empty',
  name: 'Empty',
  namespace: 'unused',
  rules: [],
  version: 1,
}

const published: PublishedRuleSetSnapshot[] = [
  {
    snapshot_id: 'snap-current',
    published_at: '2026-05-01T00:00:00Z',
    ruleset: { ...draftCurrent, version: 1 },
  },
  {
    snapshot_id: 'snap-changed',
    published_at: '2026-05-01T00:00:00Z',
    ruleset: {
      ...draftChanged,
      rules: [
        {
          ...draftChanged.rules[0],
          action: { type: 'respond', renderer: 'static', response: { payload: { status: 204 } } },
        },
      ],
      version: 4,
    },
  },
]

const namespaces: NamespaceConfig[] = [
  {
    id: 'default',
    name: 'Default',
    policies: {
      http: {
        ruleset_miss_action: { type: 'forward', forward: { timeout_ms: 5000 } },
        rule_miss_action: { type: 'forward', forward: { timeout_ms: 5000 } },
      },
    },
  },
  {
    id: 'payments',
    name: 'Payments',
    description: 'Payment mocks',
    policies: {
      http: {
        ruleset_miss_action: { type: 'respond', renderer: 'static', response: { payload: { status: 404 } } },
        rule_miss_action: { type: 'forward', forward: { timeout_ms: 1000 } },
      },
      spex: {
        ruleset_miss_action: { type: 'respond', renderer: 'static', response: { payload: { code: 404, resp: {} } } },
        rule_miss_action: { type: 'forward', forward: { timeout_ms: 2000 } },
      },
    },
  },
]

describe('ruleset entry view models', () => {
  it('derives publish state while ignoring version-only differences', () => {
    expect(rulesetPublishState(draftCurrent, published[0])).toBe('published_current')
    expect(rulesetPublishState(draftChanged, published[1])).toBe('changed')
    expect(rulesetPublishState(draftOnly, null)).toBe('draft_only')
  })

  it('filters, searches, sorts, and counts ruleset rows', () => {
    const filters = defaultRulesetEntryFilters()
    filters.query = 'payments'
    filters.publishState = 'changed'
    filters.sort = 'rule_count'

    const allRows = buildRulesetEntryRows([draftCurrent, draftChanged, draftOnly], published, {
      ...defaultRulesetEntryFilters(),
    })
    const shownRows = buildRulesetEntryRows([draftCurrent, draftChanged, draftOnly], published, filters)

    expect(shownRows.map((row) => row.ruleSet.id)).toEqual(['rs-changed'])
    expect(buildRulesetEntryMetrics(allRows, shownRows)).toMatchObject({
      total: 3,
      shown: 1,
      published: 2,
      changed: 1,
      draftOnly: 1,
      empty: 1,
    })
  })
})

describe('namespace entry view models', () => {
  it('summarizes forward and response fallback behavior', () => {
    expect(fallbackSummary(namespaces[0].policies.http.ruleset_miss_action)).toMatchObject({
      type: 'forward',
      detail: 'original request / 5000ms',
    })
    expect(fallbackSummary(namespaces[1].policies.http.ruleset_miss_action)).toMatchObject({
      type: 'respond',
      detail: 'HTTP 404',
    })
  })

  it('filters, searches, sorts, and counts namespace rows with ruleset usage', () => {
    const filters = defaultNamespaceEntryFilters()
    filters.query = 'rs-changed'
    filters.rulesetMissType = 'respond'
    filters.usage = 'used'

    const allRows = buildNamespaceEntryRows(namespaces, [draftCurrent, draftChanged, draftSpex, draftOnly], {
      ...defaultNamespaceEntryFilters(),
    })
    const shownRows = buildNamespaceEntryRows(namespaces, [draftCurrent, draftChanged, draftSpex, draftOnly], filters)

    expect(shownRows.map((row) => row.namespace.id)).toEqual(['payments'])
    expect(shownRows[0].protocols).toEqual(['http', 'spex'])
    expect(shownRows[0].usage).toMatchObject({
      rulesetCount: 2,
      ruleCount: 2,
    })
    expect(buildNamespaceEntryMetrics(allRows, shownRows)).toMatchObject({
      total: 2,
      shown: 1,
      namespaces: 1,
      forward: 2,
      response: 1,
      used: 2,
      unused: 0,
    })
  })

  it('expands namespace policies per protocol and scopes protocol filtering and usage', () => {
    const filters = defaultNamespaceEntryFilters()
    filters.protocol = 'spex'

    const rows = buildNamespaceEntryRows(namespaces, [draftCurrent, draftChanged, draftSpex, draftOnly], filters)

    expect(rows.map((row) => row.namespace.id)).toEqual(['payments'])
    const spexPolicy = rows[0].policies.find((policy) => policy.protocol === 'spex')
    expect(spexPolicy?.rulesetMiss).toMatchObject({
      type: 'respond',
      detail: 'SPEX 404',
    })
    expect(spexPolicy?.usage).toMatchObject({
      rulesetCount: 1,
      ruleCount: 1,
      rulesetIds: ['rs-spex'],
      rulesetNames: ['SPEX Changed'],
    })
    expect(rows[0].searchableText).toContain('spex')
  })

  it('keeps every linked ruleset available for explicit namespace navigation', () => {
    const rows = buildNamespaceEntryRows(
      namespaces,
      [draftCurrent, { ...draftChanged, namespace: 'default' }, draftOnly],
      defaultNamespaceEntryFilters()
    )

    const defaultNamespace = rows.find((row) => row.namespace.id === 'default')

    expect(defaultNamespace?.usage).toMatchObject({
      rulesetCount: 2,
      ruleCount: 2,
      rulesetIds: ['rs-current', 'rs-changed'],
      rulesetNames: ['Current', 'Changed'],
    })
    expect(defaultNamespace?.searchableText).toContain('rs-changed')
  })
})
