import { describe, expect, it } from 'vitest'
import type { RuntimeMetrics, RuleSet } from '@/types'
import { buildRuntimeDiagnosticRows, defaultRuntimeDiagnosticFilters } from '@/utils/runtimeDiagnostics'
import {
  buildDebugToFixRoute,
  buildRuleSeedFromDebugPayload,
  buildRuntimeDiagnosis,
  createDebugToFixPayload,
  inferDebugToFixTarget,
} from '@/utils/debugToFix'

const drafts: RuleSet[] = [
  {
    id: 'rs-orders',
    name: 'orders',
    enabled: true,
    protocol: 'http',
    namespace: 'orders',
    selector: { all: [] },
    rules: [
      {
        id: 'runtime-2',
        name: 'Runtime fallback rule',
        enabled: true,
        priority: 100,
        when: { field: 'request.path', op: 'eq', value: '/old' },
        action: { type: 'static_response', status: 200 },
      },
    ],
  },
  {
    id: 'rs-default',
    name: 'default',
    enabled: false,
    protocol: 'http',
    namespace: 'orders',
    selector: { all: [] },
    rules: [],
  },
]

const metrics: RuntimeMetrics = {
  total_requests: 2,
  matched_requests: 1,
  unmatched_requests: 1,
  error_requests: 0,
  average_duration_ms: 12,
  recent_requests: [
    {
      id: 2,
      observed_at: '2026-05-08T08:00:02Z',
      namespace: 'orders',
      method: 'POST',
      host: 'shop.test',
      path: '/api/orders',
      raw_query: 'debug=true',
      trace_id: 'trace-2',
      outcome: 'fallback',
      matched: false,
      fallback: true,
      fallback_reason: 'rule_miss',
      status: 404,
      duration_ms: 12,
      event: {
        protocol: 'http',
        namespace: 'orders',
        request: {
          method: 'POST',
          host: 'shop.test',
          path: '/api/orders',
          query: {
            debug: ['true'],
          },
        },
        meta: {
          trace_id: 'trace-2',
        },
      },
    },
    {
      id: 1,
      observed_at: '2026-05-08T08:00:01Z',
      namespace: 'orders',
      method: 'GET',
      host: 'shop.test',
      path: '/api/orders/1',
      trace_id: 'trace-1',
      outcome: 'matched',
      matched: true,
      ruleset_id: 'rs-hit',
      rule_id: 'rule-hit',
      status: 200,
      duration_ms: 6,
      event: {
        protocol: 'http',
        namespace: 'orders',
        request: {
          method: 'GET',
          host: 'shop.test',
          path: '/api/orders/1',
        },
      },
    },
  ],
}

const rows = buildRuntimeDiagnosticRows(metrics, defaultRuntimeDiagnosticFilters())

describe('debug-to-fix helpers', () => {
  it('prefers explicit runtime ruleset targets and falls back to namespace/protocol drafts', () => {
    expect(inferDebugToFixTarget(rows[1], drafts)).toMatchObject({
      rulesetId: 'rs-hit',
      reason: 'runtime_match',
    })
    expect(inferDebugToFixTarget(rows[0], drafts)).toMatchObject({
      rulesetId: 'rs-orders',
      reason: 'namespace_protocol',
    })
  })

  it('creates portable debug payloads and routes to the target workspace', () => {
    const target = inferDebugToFixTarget(rows[0], drafts)
    expect(target).not.toBeNull()
    const payload = createDebugToFixPayload(rows[0], 'simulate', target!, new Date('2026-05-10T00:00:00Z'))

    expect(payload).toMatchObject({
      id: 'runtime-2-1778371200000',
      action: 'simulate',
      targetRulesetId: 'rs-orders',
      request: {
        recordId: 2,
        method: 'POST',
        path: '/api/orders?debug=true',
        traceId: 'trace-2',
      },
    })
    expect(buildDebugToFixRoute(payload!)).toEqual({
      path: '/rulesets/rs-orders/rules',
      query: {
        workbench: 'simulate',
        debug: 'runtime-2-1778371200000',
      },
    })
  })

  it('derives diagnosis copy from runtime outcomes', () => {
    expect(buildRuntimeDiagnosis(rows[1])).toMatchObject({
      title: 'Matched a rule',
      tone: 'ok',
    })
    expect(buildRuntimeDiagnosis(rows[0])).toMatchObject({
      title: 'Rule condition miss',
      tone: 'warn',
    })
  })

  it('generates a safe rule seed from request facts without trace overfitting', () => {
    const target = inferDebugToFixTarget(rows[0], drafts)
    const payload = createDebugToFixPayload(rows[0], 'create_rule', target!, new Date('2026-05-10T00:00:00Z'))
    const rule = buildRuleSeedFromDebugPayload(payload!, drafts[0].rules)

    expect(rule).toMatchObject({
      id: 'runtime-2-2',
      enabled: true,
      priority: 110,
      action: {
        type: 'static_response',
        status: 200,
      },
    })
    expect(JSON.stringify(rule.when)).toContain('request.method')
    expect(JSON.stringify(rule.when)).toContain('request.query.debug[0]')
    expect(JSON.stringify(rule.when)).not.toContain('meta.trace_id')
  })
})
