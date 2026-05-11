import { describe, expect, it } from 'vitest'
import type { RuntimeMetrics } from '@/types'
import {
  buildRuntimeDiagnosticRows,
  buildRuntimeMetricCards,
  defaultRuntimeDiagnosticFilters,
  runtimeMatchRate,
  runtimeNamespaceOptions,
  runtimeOutcome,
  toSortedMetricEntries,
} from '@/utils/runtimeDiagnostics'

const metrics: RuntimeMetrics = {
  total_requests: 4,
  matched_requests: 1,
  unmatched_requests: 2,
  error_requests: 1,
  average_duration_ms: 18.5,
  fallback_reasons: {
    ruleset_miss: 1,
    rule_miss: 1,
  },
  status_codes: {
    '2xx': 1,
    '4xx': 2,
    '5xx': 1,
  },
  recent_requests: [
    {
      id: 4,
      observed_at: '2026-05-08T08:00:04Z',
      namespace: 'default',
      method: 'GET',
      path: '/api/error',
      trace_id: 'trace-error',
      outcome: 'error',
      matched: false,
      error: true,
      status: 500,
      message: 'boom',
      duration_ms: 30,
    },
    {
      id: 3,
      observed_at: '2026-05-08T08:00:03Z',
      namespace: 'orders',
      method: 'POST',
      path: '/api/orders',
      raw_query: 'debug=true',
      trace_id: 'trace-fallback',
      outcome: 'fallback',
      matched: false,
      fallback: true,
      fallback_reason: 'rule_miss',
      status: 404,
      duration_ms: 21,
      event: {
        protocol: 'http',
        namespace: 'orders',
        request: {
          method: 'POST',
          host: 'demo.test',
          path: '/api/orders',
        },
      },
    },
    {
      id: 2,
      observed_at: '2026-05-08T08:00:02Z',
      namespace: 'orders',
      method: 'GET',
      path: '/api/orders/1',
      trace_id: 'trace-hit',
      outcome: 'matched',
      matched: true,
      status: 200,
      ruleset_id: 'rs-orders',
      rule_id: 'rule-get-order',
      duration_ms: 8,
      event: {
        protocol: 'http',
        namespace: 'orders',
        request: {
          method: 'GET',
          host: 'demo.test',
          path: '/api/orders/1',
        },
      },
    },
    {
      id: 1,
      observed_at: '2026-05-08T08:00:01Z',
      namespace: 'payments',
      method: 'GET',
      path: '/api/payments',
      outcome: 'unmatched',
      matched: false,
      status: 404,
      duration_ms: 3,
    },
  ],
}

describe('runtime diagnostics view models', () => {
  it('classifies runtime outcomes from explicit outcome or flags', () => {
    expect(runtimeOutcome(metrics.recent_requests![0])).toBe('error')
    expect(runtimeOutcome({ ...metrics.recent_requests![0], outcome: undefined, error: false, matched: true })).toBe('matched')
    expect(runtimeOutcome({ ...metrics.recent_requests![0], outcome: undefined, error: false, fallback: true })).toBe('fallback')
    expect(runtimeOutcome({ ...metrics.recent_requests![0], outcome: undefined, error: false, fallback: false })).toBe('unmatched')
  })

  it('filters and searches recent requests', () => {
    const filters = defaultRuntimeDiagnosticFilters()
    filters.query = 'rule-get-order'
    filters.outcome = 'matched'
    filters.namespace = 'orders'

    const rows = buildRuntimeDiagnosticRows(metrics, filters)

    expect(rows).toHaveLength(1)
    expect(rows[0]).toMatchObject({
      outcome: 'matched',
      canOpenRuleset: true,
      canReplay: true,
      pathLabel: '/api/orders/1',
    })
  })

  it('sorts recent requests by slowest and status', () => {
    const slowest = defaultRuntimeDiagnosticFilters()
    slowest.sort = 'slowest'
    expect(buildRuntimeDiagnosticRows(metrics, slowest).map((row) => row.record.id)).toEqual([4, 3, 2, 1])

    const status = defaultRuntimeDiagnosticFilters()
    status.sort = 'status'
    expect(buildRuntimeDiagnosticRows(metrics, status).map((row) => row.record.status)).toEqual([500, 404, 404, 200])
  })

  it('derives dashboard cards, namespace options, and metric rankings', () => {
    expect(runtimeMatchRate(metrics)).toBe(25)
    expect(runtimeNamespaceOptions(metrics)).toEqual(['default', 'orders', 'payments'])
    expect(toSortedMetricEntries(metrics.fallback_reasons).map((item) => item.label)).toEqual([
      'rule miss',
      'ruleset miss',
    ])
    expect(buildRuntimeMetricCards(metrics).slice(0, 3)).toMatchObject([
      { label: 'Total traffic', value: '4' },
      { label: 'Match rate', value: '25%' },
      { label: 'Fallback', value: '2' },
    ])
  })
})
