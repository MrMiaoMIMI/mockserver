import { describe, expect, it } from 'vitest'
import type { RuleSet } from '@/types'
import {
  applyConditionPreset,
  buildDraftOverride,
  buildRuleAuthoringView,
  conditionPresetKey,
  operatorLabel,
  parsePredicateValueInput,
  resolveConditionPreset,
  updateConditionPresetKey,
  validatePredicateCondition,
} from '@/utils/ruleAuthoring'
import { defaultRuleForm, ruleToForm } from '@/utils/ruleFormAdapter'

const ruleSet: RuleSet = {
  id: 'rs-1',
  name: 'checkout rules',
  enabled: true,
  protocol: 'http',
  namespace: 'shop',
  selector: {
    all: [
      { field: 'request.host', op: 'eq', value: 'shop.test' },
      { field: 'request.path', op: 'prefix', value: '/api' },
    ],
  },
  rules: [
    {
      id: 'rule-001',
      name: 'Old checkout response',
      enabled: true,
      priority: 100,
      when: { field: 'request.path', op: 'eq', value: '/api/old' },
      action: {
        type: 'respond',
        renderer: 'static',
        response: { payload: { status: 200, body: { old: true } } },
      },
    },
  ],
}

describe('rule authoring helper', () => {
  it('creates backend-compatible conditions from common presets', () => {
    const header = updateConditionPresetKey(applyConditionPreset({}, 'header'), 'X-Env')
    const query = updateConditionPresetKey(applyConditionPreset({}, 'query'), 'state')

    expect(header).toMatchObject({
      field: 'request.headers.x-env[0]',
      op: 'eq',
      value: 'test',
    })
    expect(query).toMatchObject({
      field: 'request.query.state[0]',
      op: 'eq',
      value: 'qv1',
    })
    expect(resolveConditionPreset(header).label).toBe('Header')
    expect(conditionPresetKey(header)).toBe('x-env')
    expect(operatorLabel('prefix')).toBe('starts with')
  })

  it('parses predicate values while keeping plain text friendly', () => {
    expect(parsePredicateValueInput('true')).toBe(true)
    expect(parsePredicateValueInput('42')).toBe(42)
    expect(parsePredicateValueInput('GET')).toBe('GET')
    expect(parsePredicateValueInput('["a","b"]')).toEqual(['a', 'b'])
  })

  it('validates preset key and predicate field state', () => {
    const missingKey = updateConditionPresetKey(applyConditionPreset({}, 'query'), '')
    const missingField = { field: '', op: 'eq', value: 'x' }

    expect(validatePredicateCondition(missingKey)).toContain('Query key is required')
    expect(validatePredicateCondition(missingField)).toContain('Field is required')
  })

  it('builds readiness, event suggestions, and an unsaved draft override', () => {
    const form = defaultRuleForm(ruleSet)
    form.name = 'New checkout response'
    form.id = 'rule-002'
    form.conditionTree = {
      all: [
        { field: 'request.method', op: 'eq', value: 'POST' },
        { field: 'request.path', op: 'eq', value: '/api/checkout' },
      ],
    }

    const view = buildRuleAuthoringView(form, ruleSet, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
    })

    expect(view.errors).toEqual([])
    expect(view.canSimulate).toBe(true)
    expect(view.event?.request.method).toBe('POST')
    expect(view.event?.request.path).toBe('/api/checkout')
    expect(view.draftOverride?.rules.map((rule) => rule.id)).toEqual(['rule-001', 'rule-002'])
    expect(view.sectionReadiness.every((item) => item.errorCount === 0)).toBe(true)
  })

  it('replaces the edited rule in the temporary draft override', () => {
    const form = ruleToForm(ruleSet.rules[0])
    form.bodyJson = '{"status":200,"body":{"updated":true}}'
    const view = buildRuleAuthoringView(form, ruleSet, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
      lockedRuleId: 'rule-001',
    })
    const override = buildDraftOverride(ruleSet, view.rule!, 'rule-001')

    expect(override.rules).toHaveLength(1)
    expect(override.rules[0].action.response?.payload?.body).toEqual({ updated: true })
  })
})
