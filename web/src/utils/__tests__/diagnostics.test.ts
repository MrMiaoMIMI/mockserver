import { describe, expect, it } from 'vitest'
import type { RuleSet, SimulateRuleSetResponse, ValidateRuleSetResponse } from '@/types'
import { buildPublishReadinessView, buildValidationRuleDiagnostics } from '@/utils/publishReadiness'
import {
  buildDefaultSimulationEvent,
  buildSimulationDiagnostics,
  buildSimulationEventWarnings,
  buildSimulationRuleDiagnostics,
} from '@/utils/simulationDiagnostics'
import { buildRollbackPreviewView, buildSnapshotHistory } from '@/utils/snapshotRollback'
import {
  defaultWorkbenchModeForIntent,
  isWorkbenchIntent,
  isWorkbenchMode,
  workbenchIntentForMode,
  workbenchTitle,
} from '@/utils/workbenchTasks'

const ruleSet: RuleSet = {
  id: 'rs-1',
  name: 'checkout rules',
  enabled: true,
  protocol: 'http',
  namespace: 'default',
  selector: {
    all: [
      { field: 'request.host', op: 'eq', value: 'shop.test' },
      { field: 'request.path', op: 'prefix', value: '/api' },
    ],
  },
  version: 3,
  rules: [
    {
      id: 'primary-rule',
      name: 'Primary checkout',
      enabled: true,
      priority: 10,
      when: { field: 'request.method', op: 'eq', value: 'POST' },
      action: { type: 'static_response', status: 200, body: { ok: true } },
    },
    {
      id: 'fallback-rule',
      name: 'Fallback checkout',
      enabled: true,
      priority: 20,
      when: { field: 'request.path', op: 'prefix', value: '/api/fallback' },
      action: { type: 'static_response', status: 404, body: { ok: false } },
    },
  ],
}

describe('simulation diagnostics', () => {
  it('builds context-aware default events from selector and selected rule hints', () => {
    const event = buildDefaultSimulationEvent(ruleSet, ruleSet.rules[0])

    expect(event.namespace).toBe('default')
    expect(event.request.host).toBe('shop.test')
    expect(event.request.path).toBe('/api')
    expect(event.request.method).toBe('POST')
  })

  it('keeps hard selector values and warns when selected rule is unreachable', () => {
    const exactSelectorRuleSet: RuleSet = {
      ...ruleSet,
      selector: {
        all: [
          { field: 'request.host', op: 'eq', value: 'shop.test' },
          { field: 'request.path', op: 'eq', value: '/api' },
        ],
      },
      rules: [
        {
          ...ruleSet.rules[0],
          when: { field: 'request.path', op: 'eq', value: '/api/v1/debug' },
        },
      ],
    }

    const event = buildDefaultSimulationEvent(exactSelectorRuleSet, exactSelectorRuleSet.rules[0])
    const warnings = buildSimulationEventWarnings(exactSelectorRuleSet, exactSelectorRuleSet.rules[0])

    expect(event.request.path).toBe('/api')
    expect(warnings[0].message).toContain('may be unreachable')
  })

  it('treats HTTP path prefix constraints as path segments for event hints', () => {
    const prefixSelectorRuleSet: RuleSet = {
      ...ruleSet,
      selector: {
        all: [{ field: 'request.path', op: 'prefix', value: '/api' }],
      },
      rules: [
        {
          ...ruleSet.rules[0],
          when: { field: 'request.path', op: 'eq', value: '/api2/v2/debug' },
        },
      ],
    }

    const event = buildDefaultSimulationEvent(prefixSelectorRuleSet, prefixSelectorRuleSet.rules[0])
    const warnings = buildSimulationEventWarnings(prefixSelectorRuleSet, prefixSelectorRuleSet.rules[0])

    expect(event.request.path).toBe('/api')
    expect(warnings[0].message).toContain('request.path')
  })

  it('builds cache default events from selector and rule hints', () => {
    const cacheRuleSet: RuleSet = {
      id: 'cache-rs',
      name: 'cache rules',
      enabled: true,
      protocol: 'cache',
      namespace: 'default',
      selector: {
        all: [
          { field: 'request.operation', op: 'eq', value: 'get' },
          { field: 'request.key', op: 'prefix', value: 'user:' },
        ],
      },
      rules: [
        {
          id: 'cache-rule',
          name: 'Cache user lookup',
          enabled: true,
          priority: 10,
          when: { field: 'request.key', op: 'eq', value: 'user:42' },
          action: { type: 'static_response', status: 200 },
        },
      ],
    }

    const event = buildDefaultSimulationEvent(cacheRuleSet, cacheRuleSet.rules[0])

    expect(event.protocol).toBe('cache')
    expect(event.request.operation).toBe('get')
    expect(event.request.key).toBe('user:42')
    expect(event.request.ttl_ms).toBe(3000)
  })

  it('does not let cache rule hints override hard operation selectors', () => {
    const cacheRuleSet: RuleSet = {
      id: 'cache-rs',
      name: 'cache rules',
      enabled: true,
      protocol: 'cache',
      namespace: 'default',
      selector: {
        all: [
          { field: 'request.operation', op: 'eq', value: 'get' },
          { field: 'request.key', op: 'prefix', value: 'user:' },
        ],
      },
      rules: [
        {
          id: 'cache-rule',
          name: 'Cache set mismatch',
          enabled: true,
          priority: 10,
          when: { field: 'request.operation', op: 'eq', value: 'set' },
          action: { type: 'static_response', status: 200 },
        },
      ],
    }

    const event = buildDefaultSimulationEvent(cacheRuleSet, cacheRuleSet.rules[0])
    const warnings = buildSimulationEventWarnings(cacheRuleSet, cacheRuleSet.rules[0])

    expect(event.request.operation).toBe('get')
    expect(warnings[0].message).toContain('request.operation')
  })

  it('builds spex default events from selector and req hints', () => {
    const spexRuleSet: RuleSet = {
      id: 'spex-rs',
      name: 'spex rules',
      enabled: true,
      protocol: 'spex',
      namespace: 'default',
      selector: {
        all: [
          { field: 'request.cmd', op: 'prefix', value: 'shop.' },
        ],
      },
      rules: [
        {
          id: 'spex-rule',
          name: 'SPEX get order',
          enabled: true,
          priority: 10,
          when: {
            all: [
              { field: 'request.cmd', op: 'eq', value: 'shop.GetOrder' },
              { field: 'request.req.order_id', op: 'eq', value: '1001' },
              { field: 'request.param', op: 'eq', value: 'region=sg' },
            ],
          },
          action: { type: 'static_response', status: 200 },
        },
      ],
    }

    const event = buildDefaultSimulationEvent(spexRuleSet, spexRuleSet.rules[0])

    expect(event.protocol).toBe('spex')
    expect(event.request.cmd).toBe('shop.GetOrder')
    expect(event.request.req).toMatchObject({ order_id: '1001' })
    expect(event.request.param).toBe('region=sg')
  })

  it('does not let host hints override hard host selectors', () => {
    const hostRuleSet: RuleSet = {
      ...ruleSet,
      selector: {
        all: [{ field: 'request.host', op: 'in', value: ['shop.test', 'api.test'] }],
      },
      rules: [
        {
          ...ruleSet.rules[0],
          when: { field: 'request.host', op: 'eq', value: 'other.test' },
        },
      ],
    }

    const event = buildDefaultSimulationEvent(hostRuleSet, hostRuleSet.rules[0])
    const warnings = buildSimulationEventWarnings(hostRuleSet, hostRuleSet.rules[0])

    expect(event.request.host).toBe('shop.test')
    expect(warnings[0].message).toContain('request.host')
  })

  it('maps simulation results to structured diagnostics and rule statuses', () => {
    const response: SimulateRuleSetResponse = {
      result: {
        matched: true,
        trace: {
          ruleset_id: 'rs-1',
          rule_id: 'primary-rule',
        },
        candidates: ['primary-rule', 'fallback-rule'],
        response: {
          status: 200,
        },
      },
    }

    expect(buildSimulationDiagnostics(response.result)?.matchedRule).toBe('primary-rule')
    expect(buildSimulationRuleDiagnostics(response.result, ruleSet.rules)).toMatchObject({
      'primary-rule': { simulation: 'matched' },
      'fallback-rule': { simulation: 'candidate' },
    })
  })

  it('marks rules as not reached when simulation misses at the ruleset selector', () => {
    const response: SimulateRuleSetResponse = {
      result: {
        matched: false,
        trace: {},
        explain: {
          rule_set_explanations: [
            {
              ruleset_id: 'rs-1',
              matched: false,
              selector_checks: [
                {
                  name: 'selector.all[0]',
                  matched: false,
                  message: 'field request.path did not match operator eq',
                },
              ],
            },
          ],
        },
      },
    }

    expect(buildSimulationRuleDiagnostics(response.result, ruleSet.rules)).toMatchObject({
      'primary-rule': { simulation: 'not_reached' },
      'fallback-rule': { simulation: 'not_reached' },
    })
  })

  it('surfaces a ruleset miss even when individual selector checks passed', () => {
    const response: SimulateRuleSetResponse = {
      result: {
        matched: false,
        trace: {},
        explain: {
          rule_set_explanations: [
            {
              ruleset_id: 'rs-1',
              matched: false,
              message: 'ruleset selector matched, but no rule matched',
              selector_checks: [
                {
                  name: 'request.path',
                  matched: true,
                  expected: '/api',
                  actual: ['/api/v2/miss'],
                  message: 'field request.path matched operator prefix',
                },
              ],
              rule_explanations: [
                {
                  rule_id: 'primary-rule',
                  priority: 10,
                  matched: false,
                  condition: {
                    kind: 'predicate',
                    field: 'request.path',
                    operator: 'eq',
                    expected: '/api/v2/debug',
                    actual: ['/api/v2/miss'],
                    matched: false,
                    message: 'field request.path did not match indexed equality',
                  },
                },
              ],
            },
          ],
        },
      },
    }

    const diagnostics = buildSimulationDiagnostics(response.result)

    expect(diagnostics?.matched).toBe(false)
    expect(diagnostics?.rulesetTraces[0]).toMatchObject({
      label: 'rs-1',
      matched: false,
      message: 'ruleset selector matched, but no rule matched',
    })
    expect(diagnostics?.selectorTraces[0]).toMatchObject({
      matched: true,
      label: 'rs-1 / request.path',
    })
    expect(diagnostics?.conditionTraces[0]).toMatchObject({
      matched: false,
      label: 'primary-rule / request.path',
      detail: 'expected /api/v2/debug / actual ["/api/v2/miss"]',
    })
  })
})

describe('validation and rollback diagnostics', () => {
  it('maps validation issues back to rule diagnostics', () => {
    const validation: ValidateRuleSetResponse = {
      result: {
        valid: false,
        issues: [{ path: 'rules[1].action.status', message: 'status must be >= 100' }],
      },
    }

    expect(buildValidationRuleDiagnostics(validation, ruleSet)).toMatchObject({
      'fallback-rule': { validation: 'invalid' },
    })
  })

  it('keeps publish allowed while surfacing validation warnings', () => {
    const validation: ValidateRuleSetResponse = {
      result: {
        valid: true,
        warnings: [{ path: 'rules[0].when', message: 'rule may be unreachable' }],
      },
    }

    const view = buildPublishReadinessView(ruleSet, null, validation, null)

    expect(view.canPublish).toBe(true)
    expect(view.warnings).toHaveLength(1)
    expect(view.validationDetail).toContain('runtime warning')
  })

  it('builds rollback and snapshot view models for contextual workflows', () => {
    const snapshots = [
      {
        snapshot_id: 'snap-old',
        published_at: '2026-05-01T00:00:00Z',
        ruleset: { ...ruleSet, version: 1 },
        audit: { operator: 'dev', reason: 'baseline' },
      },
      {
        snapshot_id: 'snap-new',
        published_at: '2026-05-02T00:00:00Z',
        ruleset: ruleSet,
        audit: { operator: 'dev', reason: 'latest' },
      },
    ]
    const preview = buildRollbackPreviewView({
      result: {
        snapshot: snapshots[0],
        valid: true,
        validation: { valid: true },
        diff: {
          current_snapshot_id: 'snap-new',
          target_snapshot_id: 'snap-old',
          changed: true,
          rule_diffs: [{ rule_id: 'primary-rule', change_type: 'modified' }],
        },
      },
    })

    expect(buildSnapshotHistory(snapshots)[0].id).toBe('snap-new')
    expect(preview?.ruleDiffs[0].ruleId).toBe('primary-rule')
  })
})

describe('workbench task model', () => {
  it('accepts known route query modes and derives selected-rule titles', () => {
    expect(isWorkbenchMode('inspect')).toBe(true)
    expect(isWorkbenchMode('unknown')).toBe(false)
    expect(workbenchTitle('editor', ruleSet.rules[0], 'edit')).toBe('Edit Primary checkout')
  })

  it('groups detailed workbench modes into user intents', () => {
    expect(isWorkbenchIntent('rules')).toBe(true)
    expect(isWorkbenchIntent('unknown')).toBe(false)
    expect(workbenchIntentForMode('inspect')).toBe('rules')
    expect(workbenchIntentForMode('editor')).toBe('rules')
    expect(workbenchIntentForMode('simulate')).toBe('test')
    expect(workbenchIntentForMode('result')).toBe('test')
    expect(workbenchIntentForMode('readiness')).toBe('release')
    expect(workbenchIntentForMode('snapshots')).toBe('release')
    expect(defaultWorkbenchModeForIntent('rules')).toBe('inspect')
    expect(defaultWorkbenchModeForIntent('test')).toBe('simulate')
    expect(defaultWorkbenchModeForIntent('release')).toBe('readiness')
  })
})
