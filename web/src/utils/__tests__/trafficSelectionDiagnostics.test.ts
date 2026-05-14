import { describe, expect, it } from 'vitest'
import { buildTrafficSelectionDiagnostics } from '@/utils/trafficSelectionDiagnostics'
import type { TrafficEvent } from '@/types'

describe('trafficSelectionDiagnostics', () => {
  it('extracts ruleset and rule selection diagnostics', () => {
    const event = {
      explain: {
        ruleset_selection: {
          winner_ruleset_id: 'rs-api',
          winner_snapshot_id: 'snap-1',
          candidates: [
            {
              ruleset_id: 'rs-api',
              snapshot_id: 'snap-1',
              selector_matched: true,
              selected: true,
              selector_specificity: 120,
              message: 'selected as most specific ruleset',
            },
            {
              ruleset_id: 'rs-default',
              selector_matched: true,
              selected: false,
              selector_specificity: 10,
              message: 'ruleset selector matched, but a more specific ruleset was selected',
            },
          ],
        },
        rule_selection: {
          candidate_rule_ids: ['rule-a', 'rule-b'],
          winner_rule_id: 'rule-b',
        },
      },
    } as TrafficEvent

    const diagnostics = buildTrafficSelectionDiagnostics(event)

    expect(diagnostics.hasDiagnostics).toBe(true)
    expect(diagnostics.winnerRuleSetId).toBe('rs-api')
    expect(diagnostics.winnerSnapshotId).toBe('snap-1')
    expect(diagnostics.winnerRuleId).toBe('rule-b')
    expect(diagnostics.candidateRuleIds).toEqual(['rule-a', 'rule-b'])
    expect(diagnostics.rulesetCandidates).toHaveLength(2)
    expect(diagnostics.rulesetCandidates[0]).toMatchObject({
      rulesetId: 'rs-api',
      selected: true,
      selectorSpecificity: 120,
    })
  })

  it('returns an empty view model for traffic without diagnostics', () => {
    const diagnostics = buildTrafficSelectionDiagnostics({ explain: { trace: {} } } as TrafficEvent)

    expect(diagnostics.hasDiagnostics).toBe(false)
    expect(diagnostics.rulesetCandidates).toEqual([])
    expect(diagnostics.candidateRuleIds).toEqual([])
  })
})
