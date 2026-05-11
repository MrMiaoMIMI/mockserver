import type { Condition } from '@/types'
import { formatInlineValue } from '@/utils/ruleSummaries'

export interface ConditionTreeRow {
  number: string
  depth: number
  label: string
  detail: string
}

export function buildConditionTreeRows(
  condition: Condition,
  number = '1',
  depth = 0
): ConditionTreeRow[] {
  if (condition.all) {
    return buildGroupRows('ALL', condition.all, number, depth)
  }
  if (condition.any) {
    return buildGroupRows('ANY', condition.any, number, depth)
  }
  if (condition.not) {
    return [
      { number, depth, label: 'NOT', detail: 'invert child result' },
      ...buildConditionTreeRows(condition.not, `${number}.1`, depth + 1),
    ]
  }
  if (condition.expr !== undefined) {
    return [{ number, depth, label: 'CEL', detail: condition.expr || 'empty expression' }]
  }
  const value = formatInlineValue(condition.value)
  return [
    {
      number,
      depth,
      label: condition.field || 'field',
      detail: value ? `${condition.op || 'eq'} ${value}` : condition.op || 'eq',
    },
  ]
}

function buildGroupRows(kind: 'ALL' | 'ANY', children: Condition[], number: string, depth: number) {
  return [
    {
      number,
      depth,
      label: kind,
      detail: `${children.length} child${children.length === 1 ? '' : 'ren'}`,
    },
    ...children.flatMap((child, index) =>
      buildConditionTreeRows(child, `${number}.${index + 1}`, depth + 1)
    ),
  ]
}
