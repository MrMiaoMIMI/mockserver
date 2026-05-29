import { describe, expect, it } from 'vitest'
import type { ProtocolSpec, Rule, RuleSet } from '@/types'
import {
  applyRawRuleJson,
  defaultRuleForm,
  formToRule,
  generateRuleId,
  newSequenceStepForm,
  ruleToForm,
} from '@/utils/ruleFormAdapter'
import { invalidJSONLiteralValue } from '@/utils/valueInputSpec'

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
      action: {
        type: 'respond',
        renderer: 'static',
        response: { payload: { status: 200, body: { ok: true } } },
      },
    },
  ],
}

const httpSpec: ProtocolSpec = {
  name: 'http',
  fields: [],
  response: {
    defaults: {
      status: 200,
      headers: { 'content-type': ['application/json'] },
      body: {},
    },
    fields: [
      { path: 'status', type: 'number', required: true, default: 200, min: 100, max: 599 },
      { path: 'headers', type: 'object', default: { 'content-type': ['application/json'] } },
      { path: 'body', type: 'json', default: {} },
    ],
  },
}

const cacheSpec: ProtocolSpec = {
  name: 'cache',
  fields: [],
  response: {
    defaults: { hit: true, value: null },
    fields: [
      { path: 'hit', type: 'bool', required: true, default: true },
      { path: 'value', type: 'json' },
    ],
  },
}

const spexSpec: ProtocolSpec = {
  name: 'spex',
  fields: [],
  response: {
    defaults: { code: 0, resp: {} },
    fields: [
      { path: 'code', type: 'number', required: true, default: 0 },
      { path: 'resp', type: 'json', required: true, default: {} },
    ],
  },
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
    type: 'respond',
    renderer: 'static',
    response: {
      payload: {
        status: 201,
        headers: { 'content-type': ['application/json'] },
        body: { ok: true },
      },
    },
  },
}

describe('rule form adapter', () => {
  it('round-trips a static rule without changing API payload shape', () => {
    const form = ruleToForm(staticRule, httpSpec)
    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
      lockedRuleId: staticRule.id,
      protocolSpec: httpSpec,
    })

    expect(result.errors).toEqual([])
    expect(result.rule).toEqual({
      ...staticRule,
      action: {
        ...staticRule.action,
        response: {
          protocol: 'http',
          payload: staticRule.action.response?.payload,
        },
      },
    })
  })

  it('builds sequence response payloads from ordered step forms', () => {
    const form = defaultRuleForm(ruleSet, httpSpec)
    form.name = 'Sequence rule'
    form.id = 'sequence-rule'
    form.priority = 30
    form.actionType = 'sequence'
    form.sequenceStrategy = 'loop'
    form.sequenceSteps = [
      {
        ...newSequenceStepForm(1, httpSpec),
        responsePayload: { status: 202, headers: {}, body: { step: 1 } },
        responseFieldDrafts: { body: '{"step":1}' },
      },
      {
        ...newSequenceStepForm(2, httpSpec),
        responsePayload: { status: 203, headers: {}, body: { step: 2 } },
        responseFieldDrafts: { body: '{"step":2}' },
      },
    ]

    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
      protocolSpec: httpSpec,
    })

    expect(result.errors).toEqual([])
    expect(result.rule?.action).toMatchObject({
      type: 'respond',
      renderer: 'sequence',
      sequence_strategy: 'loop',
      sequence: [
        { response: { payload: { status: 202 } } },
        { response: { payload: { status: 203 } } },
      ],
    })
  })

  it('rejects duplicate rule names and priorities before saving', () => {
    const form = defaultRuleForm(ruleSet, httpSpec)
    form.id = 'new-rule'
    form.name = ' existing   rule '
    form.priority = 10

    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
      existingRules: ruleSet.rules,
      protocolSpec: httpSpec,
    })

    expect(result.rule).toBeUndefined()
    expect(result.errors.map((error) => error.field)).toEqual(expect.arrayContaining(['name', 'priority']))
  })

  it('uses cache-friendly default conditions for cache rulesets', () => {
    const form = defaultRuleForm({ ...ruleSet, protocol: 'cache' }, cacheSpec)

    expect(form.conditionTree).toMatchObject({
      all: [
        { field: 'request.operation', op: 'eq', value: 'get' },
        { field: 'request.key', op: 'prefix', value: 'user:' },
      ],
    })
  })

  it('uses spex-friendly default conditions for spex rulesets', () => {
    const form = defaultRuleForm({ ...ruleSet, protocol: 'spex' }, spexSpec)

    expect(form.conditionTree).toMatchObject({
      all: [
        { field: 'request.cmd', op: 'prefix', value: 'service.' },
        { field: 'request.req.id', op: 'eq', value: 'demo' },
      ],
    })
  })

  it('validates raw JSON mode and preserves locked rule IDs', () => {
    const form = ruleToForm(staticRule, httpSpec)
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
    const form = defaultRuleForm(ruleSet, httpSpec)
    form.name = ''

    const result = formToRule(form, {
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
      protocolSpec: httpSpec,
    })

    expect(result.errors.some((error) => error.field === 'name')).toBe(true)

    form.rawRuleJson = JSON.stringify({ ...staticRule, name: '' })
    const rawResult = formToRule(form, {
      source: 'raw',
      existingRuleIds: ruleSet.rules.map((rule) => rule.id),
    })

    expect(rawResult.errors.some((error) => error.message.includes('name'))).toBe(true)
  })

  it('generates stable-looking opaque rule IDs and avoids duplicates', () => {
    expect(generateRuleId(ruleSet)).toMatch(/^rule_[a-f0-9]{8}(?:-\d{2})?$/)
  })

  it('does not treat the locked edit rule ID as a duplicate when generating IDs', () => {
    expect(generateRuleId(ruleSet, 'existing-rule')).toMatch(/^rule_[a-f0-9]{8}$/)
  })

  it('rejects invalid JSON literal condition values before building a rule payload', () => {
    const form = defaultRuleForm(ruleSet, httpSpec)
    form.name = 'Invalid literal'
    form.id = 'invalid-literal'
    form.conditionTree = {
      field: 'request.body.count',
      op: 'eq',
      value: invalidJSONLiteralValue('abc', 'Invalid JSON value. String values must use double quotes.'),
    }

    const result = formToRule(form, { protocolSpec: httpSpec })

    expect(result.rule).toBeUndefined()
    expect(result.errors).toContainEqual({
      section: 'condition',
      field: 'conditionTree',
      message: 'Invalid JSON value. String values must use double quotes.',
    })
  })

  it('persists valid rule-level sample request authoring metadata', () => {
    const form = defaultRuleForm(ruleSet, httpSpec)
    form.name = 'Sample request rule'
    form.id = 'sample-request-rule'
    form.sampleRequestRaw = `curl -X POST 'https://demo.com/api/order?region=SG' -H 'x-env: test' --data-raw '{"amount":123}'`

    const result = formToRule(form, { protocolSpec: httpSpec })

    expect(result.errors).toEqual([])
    expect(result.rule?.authoring?.sample_request).toMatchObject({
      format: 'curl',
      root: 'request',
      raw: form.sampleRequestRaw,
      parsed: {
        method: 'POST',
        host: 'demo.com',
        path: '/api/order',
        query: { region: ['SG'] },
        headers: { 'x-env': ['test'] },
        body: { amount: 123 },
      },
    })

    const roundTripped = ruleToForm(result.rule as Rule, httpSpec)
    expect(roundTripped.sampleRequestRaw).toBe(form.sampleRequestRaw)
  })

  it('rejects invalid sample request content before saving a rule', () => {
    const form = defaultRuleForm(ruleSet, httpSpec)
    form.name = 'Bad sample request'
    form.id = 'bad-sample-request'
    form.sampleRequestRaw = `curl https://demo.com -d 'amount=123'`

    const result = formToRule(form, { protocolSpec: httpSpec })

    expect(result.rule).toBeUndefined()
    expect(result.errors.some((error) => (
      error.section === 'condition'
        && error.field === 'sampleRequestRaw'
        && error.message.includes('valid JSON')
    ))).toBe(true)
  })
})
