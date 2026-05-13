import { describe, expect, it } from 'vitest'
import type { Rule, RuleSet } from '@/types'
import {
  applyRawRuleJson,
  defaultRuleForm,
  formToRule,
  generateRuleIdFromName,
  newSequenceStepForm,
  ruleToForm,
} from '@/utils/ruleFormAdapter'

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
      id: 'existing-rule',
      name: 'Existing rule',
      enabled: true,
      priority: 10,
      when: { field: 'request.path', op: 'eq', value: '/api/existing' },
      action: { type: 'static_response', status: 200, body: { ok: true } },
    },
  ],
}

const staticRule: Rule = {
  id: 'static-rule',
  name: 'Static rule',
  enabled: true,
  priority: 20,
  when: {
    all: [
      { field: 'request.method', op: 'eq', value: 'GET' },
      { field: 'request.path', op: 'eq', value: '/api/static' },
    ],
  },
  action: {
    type: 'static_response',
    status: 201,
    headers: { 'content-type': ['application/json'] },
    body: { ok: true },
  },
}

describe('rule form adapter', () => {
  it('round-trips a static rule without changing API payload shape', () => {
    const form = ruleToForm(staticRule)
    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
      lockedRuleId: staticRule.id,
    })

    expect(result.errors).toEqual([])
    expect(result.rule).toEqual(staticRule)
  })

  it('builds sequence response payloads from ordered step forms', () => {
    const form = defaultRuleForm(ruleSet)
    form.name = 'Sequence rule'
    form.id = 'sequence-rule'
    form.priority = 30
    form.actionType = 'sequence_response'
    form.sequenceStrategy = 'loop'
    form.sequenceSteps = [
      { ...newSequenceStepForm(1), status: 202, bodyJson: '{"step":1}' },
      { ...newSequenceStepForm(2), status: 203, bodyJson: '{"step":2}' },
    ]

    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
    })

    expect(result.errors).toEqual([])
    expect(result.rule?.action).toMatchObject({
      type: 'sequence_response',
      sequence_strategy: 'loop',
      sequence: [{ status: 202 }, { status: 203 }],
    })
  })

  it('uses cache-friendly default conditions for cache rulesets', () => {
    const form = defaultRuleForm({ ...ruleSet, protocol: 'cache' })

    expect(form.conditionTree).toMatchObject({
      all: [
        { field: 'request.operation', op: 'eq', value: 'get' },
        { field: 'request.key', op: 'prefix', value: 'user:' },
      ],
    })
  })

  it('uses spex-friendly default conditions for spex rulesets', () => {
    const form = defaultRuleForm({ ...ruleSet, protocol: 'spex' })

    expect(form.conditionTree).toMatchObject({
      all: [
        { field: 'request.cmd', op: 'prefix', value: 'service.' },
        { field: 'request.req.id', op: 'eq', value: 'demo' },
      ],
    })
  })

  it('validates raw JSON mode and preserves locked rule IDs', () => {
    const form = ruleToForm(staticRule)
    form.rawRuleJson = JSON.stringify({ ...staticRule, id: 'changed-id' })

    const result = applyRawRuleJson(form)
    const lockedResult = formToRule(form, {
      source: 'raw',
      lockedRuleId: staticRule.id,
    })

    expect(result.errors).toEqual([])
    expect(lockedResult.errors[0]?.field).toBe('rawRuleJson')
  })

  it('requires rule names for normal form and raw JSON payloads', () => {
    const form = defaultRuleForm(ruleSet)
    form.name = ''

    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
    })

    expect(result.errors.some((error) => error.field === 'name')).toBe(true)

    form.rawRuleJson = JSON.stringify({ ...staticRule, name: '' })
    const rawResult = formToRule(form, {
      source: 'raw',
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
    })

    expect(rawResult.errors.some((error) => error.message.includes('name'))).toBe(true)
  })

  it('generates stable rule IDs from rule names and avoids duplicates', () => {
    expect(generateRuleIdFromName('Debug API Response', ruleSet)).toBe('debug-api-response')
    expect(generateRuleIdFromName('Existing Rule', ruleSet)).toBe('existing-rule-02')
    expect(generateRuleIdFromName('', ruleSet)).toBe('rule-002')
  })

  it('does not treat the locked edit rule ID as a duplicate when generating IDs', () => {
    expect(generateRuleIdFromName('Existing Rule', ruleSet, 'existing-rule')).toBe('existing-rule')
  })
})
