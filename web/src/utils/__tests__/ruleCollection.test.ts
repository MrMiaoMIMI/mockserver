import { describe, expect, it } from 'vitest'
import type { RuleSet } from '@/types'
import {
  buildRuleRows,
  buildRuleStatusCounts,
  defaultRuleFilters,
  resolveStableSelectedRuleId,
} from '@/utils/ruleCollection'
import { actionSummary, conditionSummary, nextDuplicateRuleId } from '@/utils/ruleSummaries'

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
  rules: [
    {
      id: 'disabled-rule',
      name: 'Disabled fallback',
      enabled: false,
      priority: 20,
      when: { field: 'request.path', op: 'prefix', value: '/api/fallback' },
      action: { type: 'respond', renderer: 'template', response_template: '{"status":202,"body":"{{ request.path }}"}' },
    },
    {
      id: 'primary-rule',
      name: 'Primary checkout',
      enabled: true,
      priority: 10,
      when: {
        all: [
          { field: 'request.method', op: 'eq', value: 'POST' },
          { field: 'request.path', op: 'eq', value: '/api/checkout' },
        ],
      },
      action: { type: 'respond', renderer: 'static', response: { payload: { status: 200, body: { ok: true } } } },
    },
  ],
}

describe('rule collection view helpers', () => {
  it('orders rule rows by priority and keeps readable summaries', () => {
    const rows = buildRuleRows(ruleSet, defaultRuleFilters(), 'primary-rule')

    expect(rows.map((row) => row.rule.id)).toEqual(['primary-rule', 'disabled-rule'])
    expect(rows[0].selected).toBe(true)
    expect(rows[0].condition).toContain('ALL 2')
    expect(rows[0].action).toContain('static response')
  })

  it('filters by search, enabled state, action type, and diagnostics', () => {
    const filters = defaultRuleFilters()
    filters.query = 'fallback'
    filters.enabled = 'disabled'
    filters.actionType = 'template'
    filters.status = 'invalid'

    const rows = buildRuleRows(ruleSet, filters, '', {
      'disabled-rule': {
        validation: 'invalid',
        messages: ['rules[1].action: bad template'],
      },
    })

    expect(rows).toHaveLength(1)
    expect(rows[0].rule.id).toBe('disabled-rule')
    expect(rows[0].statuses.some((status) => status.key === 'invalid')).toBe(true)
  })

  it('derives counts and stable selection independently from filters', () => {
    const rows = buildRuleRows(ruleSet, defaultRuleFilters(), 'missing-rule', {
      'primary-rule': { simulation: 'matched' },
      'disabled-rule': { validation: 'invalid' },
    })

    expect(buildRuleStatusCounts(rows)).toEqual({
      total: 2,
      enabled: 1,
      disabled: 1,
      invalid: 1,
      matched: 1,
      missed: 0,
    })
    expect(resolveStableSelectedRuleId(ruleSet, 'missing-rule')).toBe('primary-rule')
    expect(resolveStableSelectedRuleId(ruleSet, 'disabled-rule')).toBe('disabled-rule')
  })

  it('uses distinct warning tones for missed and not reached simulation states', () => {
    const rows = buildRuleRows(ruleSet, defaultRuleFilters(), '', {
      'primary-rule': { simulation: 'missed' },
      'disabled-rule': { simulation: 'not_reached' },
    })

    const missed = rows[0].statuses.find((status) => status.key === 'missed')
    const notReached = rows[1].statuses.find((status) => status.key === 'not_reached')

    expect(missed?.tone).toBe('warn')
    expect(notReached?.tone).toBe('caution')
    expect(notReached?.tone).not.toBe(missed?.tone)
  })
})

describe('rule summary helpers', () => {
  it('summarizes condition, action, and duplicate ids consistently', () => {
    expect(conditionSummary(ruleSet.rules[0].when)).toContain('request.path prefix')
    expect(actionSummary(ruleSet.rules[1].action)).toContain('HTTP 200')
    expect(nextDuplicateRuleId('primary-rule', ruleSet.rules)).toBe('primary-rule-copy')
  })
})
