<template>
  <div class="condition-node" :class="{ nested: depth > 0 }">
    <div class="node-toolbar">
      <el-select :model-value="kind" size="small" class="kind-select" @change="changeKind">
        <el-option label="Predicate" value="predicate" />
        <el-option label="ALL" value="all" />
        <el-option label="ANY" value="any" />
        <el-option label="NOT" value="not" />
        <el-option label="CEL Expr" value="expr" />
      </el-select>
      <el-button v-if="kind === 'all' || kind === 'any'" size="small" @click="appendChild">
        Add child
      </el-button>
      <el-popover
        v-if="kind === 'predicate'"
        placement="top-end"
        trigger="hover"
        :width="420"
        popper-class="condition-payload-popover"
      >
        <template #reference>
          <button class="inline-tool" type="button">
            <el-icon><InfoFilled /></el-icon>
            <span>Payload</span>
          </button>
        </template>
        <div class="payload-popover">
          <span>condition payload</span>
          <code>{{ predicatePreview }}</code>
        </div>
      </el-popover>
    </div>

    <div v-if="kind === 'predicate'" class="predicate-builder">
      <div v-if="!protocolFieldOptions.length" class="predicate-preset-row">
        <el-select
          :model-value="activePreset.id"
          size="small"
          class="preset-select"
          @change="changePreset"
        >
          <el-option
            v-for="preset in conditionPresets"
            :key="preset.id"
            :label="preset.label"
            :value="preset.id"
          />
          <el-option label="Custom field" value="custom" />
        </el-select>

        <el-input
          v-if="activePreset.fieldPrefix && !protocolFieldOptions.length"
          :model-value="activePresetKey"
          size="small"
          :placeholder="activePreset.keyPlaceholder"
          @update:model-value="updatePresetKeyValue"
        >
          <template #prepend>{{ activePreset.keyLabel }}</template>
        </el-input>
      </div>

      <div class="predicate-grid">
        <div class="field-editor">
          <el-select
            v-if="protocolFieldOptions.length"
            :model-value="fieldSelectValue"
            size="small"
            class="protocol-field-select"
            filterable
            allow-create
            default-first-option
            placeholder="field"
            @update:model-value="updateProtocolField"
          >
            <el-option
              v-for="field in protocolFieldOptions"
              :key="field.path"
              :label="field.path"
              :value="field.path"
            />
          </el-select>
          <el-input
            v-else-if="activePreset.id === 'custom'"
            :model-value="modelValue.field"
            size="small"
            placeholder="request.path"
            @update:model-value="updateProtocolField"
          />
          <div v-else class="predicate-field-label">
            <span>Field</span>
            <strong>{{ activePreset.label }}</strong>
          </div>

          <div v-if="dynamicParts" class="dynamic-field-builder">
            <el-input
              :model-value="dynamicParts.suffix"
              size="small"
              :placeholder="dynamicParts.placeholder"
              @update:model-value="updateDynamicSuffix"
            >
              <template #prepend>{{ dynamicParts.label }}</template>
            </el-input>
            <el-select
              v-if="dynamicParts.indexed"
              :model-value="dynamicParts.indexMode"
              size="small"
              class="index-mode-select"
              @update:model-value="updateDynamicIndexMode"
            >
              <el-option label="first" value="first" />
              <el-option label="any" value="any" />
            </el-select>
          </div>
        </div>

        <el-select
          :model-value="modelValue.op || activePreset.defaultOp"
          size="small"
          placeholder="op"
          @update:model-value="updateOperator"
        >
          <el-option
            v-for="op in predicateOperatorOptions"
            :key="op"
            :label="operatorLabel(op)"
            :value="op"
          />
        </el-select>

        <SmartValueInput
          :model-value="modelValue.value"
          :spec="valueInputSpec"
          @update:model-value="updatePredicateValue"
        />
      </div>

      <div v-if="predicateErrors.length" class="predicate-errors">
        <span v-for="error in predicateErrors" :key="error">{{ error }}</span>
      </div>
      <div v-if="predicateWarnings.length" class="predicate-warnings">
        <span v-for="warning in predicateWarnings" :key="warning">{{ warning }}</span>
      </div>
    </div>

    <el-input
      v-else-if="kind === 'expr'"
      :model-value="modelValue.expr"
      type="textarea"
      :rows="4"
      placeholder="request.path == '/api/v1/debug'"
      @update:model-value="emitUpdate({ expr: $event })"
    />

    <div v-else-if="kind === 'not'" class="children-block">
      <ConditionTreeEditor
        :model-value="modelValue.not || defaultCondition('predicate')"
        :depth="depth + 1"
        :protocol-spec="protocolSpec"
        :sample-fields="sampleFields"
        @update:model-value="emitUpdate({ not: $event })"
      />
    </div>

    <div v-else class="children-block">
      <div v-for="(child, index) in groupChildren" :key="index" class="child-row">
        <ConditionTreeEditor
          :model-value="child"
          :depth="depth + 1"
          :protocol-spec="protocolSpec"
          :sample-fields="sampleFields"
          @update:model-value="updateChild(index, $event)"
        />
        <el-button
          class="remove-child"
          size="small"
          :disabled="groupChildren.length <= 1"
          @click="removeChild(index)"
        >
          Delete
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'
import SmartValueInput from '@/components/rulesets/SmartValueInput.vue'
import type { Condition, ProtocolSpec } from '@/types'
import type { SampleRequestField } from '@/utils/sampleRequestFields'
import { sampleTypeWarning } from '@/utils/sampleRequestFields'
import {
  buildDynamicFieldPath,
  dynamicFieldParts,
  fieldForPath,
  isRuleConditionFieldPath,
  operatorsForField,
  ruleConditionFields,
  type DynamicIndexMode,
} from '@/utils/protocolFields'
import {
  buildValueInputSpec,
  defaultValueForSpec,
  isInvalidJSONLiteralValue,
  normalizeValueForSpec,
  valueInputNeedsValue,
  type ValueInputSpec,
} from '@/utils/valueInputSpec'
import {
  CONDITION_PRESETS,
  applyConditionPreset,
  conditionPresetKey,
  operatorLabel,
  resolveConditionPreset,
  updateConditionPresetKey,
  validatePredicateCondition,
  type ConditionPresetId,
} from '@/utils/ruleAuthoring'

type ConditionKind = 'predicate' | 'all' | 'any' | 'not' | 'expr'

const props = withDefaults(
  defineProps<{
    modelValue: Condition
    depth?: number
    protocolSpec?: ProtocolSpec
    sampleFields?: SampleRequestField[]
  }>(),
  {
    depth: 0,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: Condition]
}>()

const conditionPresets = CONDITION_PRESETS.filter((preset) => isRuleConditionFieldPath(preset.defaultField))
const kind = computed(() => conditionKind(props.modelValue))
const activePreset = computed(() => resolveConditionPreset(props.modelValue))
const activePresetKey = computed(() => conditionPresetKey(props.modelValue))
const protocolFieldOptions = computed(() => mergeProtocolFieldOptions(
  ruleConditionFields(props.protocolSpec?.fields || []),
  props.sampleFields || []
))
const combinedProtocolSpec = computed<ProtocolSpec | undefined>(() => props.protocolSpec
  ? { ...props.protocolSpec, fields: protocolFieldOptions.value }
  : props.sampleFields?.length
    ? { name: 'sample', fields: protocolFieldOptions.value }
    : undefined)
const selectedFieldPath = computed(() => {
  const field = props.modelValue.field || ''
  return isRuleConditionFieldPath(field) ? field : defaultProtocolField.value || activePreset.value.defaultField
})
const selectedProtocolField = computed(() => protocolFieldForPath(selectedFieldPath.value))
const defaultProtocolField = computed(() => {
  return protocolFieldOptions.value.find((field) => field.path.startsWith('request.'))?.path || protocolFieldOptions.value[0]?.path || ''
})
const fieldSelectValue = computed(() => {
  return selectedProtocolField.value?.dynamic_path ? selectedProtocolField.value.path : selectedFieldPath.value
})
const predicateOperatorOptions = computed(() => {
  const fieldOperators = operatorsForField(selectedProtocolField.value)
  return fieldOperators.length ? fieldOperators : activePreset.value.operators
})
const valueInputSpec = computed(() =>
  buildValueInputSpec({
    fieldPath: selectedFieldPath.value,
    field: selectedProtocolField.value,
    operator: props.modelValue.op || activePreset.value.defaultOp,
  })
)
const dynamicParts = computed(() => dynamicFieldParts(selectedFieldPath.value, selectedProtocolField.value))
const predicatePreview = computed(() => formatJSON(cleanPredicate(props.modelValue, { omitInvalidJSONLiteral: true })))
const predicateErrors = computed(() =>
  kind.value === 'predicate' ? validatePredicateCondition(props.modelValue) : []
)
const predicateWarnings = computed(() => {
  if (kind.value !== 'predicate') return []
  const warning = sampleTypeWarning(
    props.modelValue.field || '',
    props.modelValue.op || '',
    props.modelValue.value,
    props.sampleFields || []
  )
  return warning ? [warning] : []
})
const groupChildren = computed(() => {
  if (kind.value === 'all') {
    return props.modelValue.all || []
  }
  if (kind.value === 'any') {
    return props.modelValue.any || []
  }
  return []
})

function changeKind(nextKind: ConditionKind) {
  emitUpdate(defaultCondition(nextKind))
}

function updatePredicate(partial: Partial<Condition>) {
  const field = partial.field || selectedFieldPath.value
  const op = partial.op || props.modelValue.op || activePreset.value.defaultOp
  const nextSpec = buildValueInputSpec({
    fieldPath: field,
    field: protocolFieldForPath(field),
    operator: op,
  })
  const next: Condition = {
    field,
    op,
    ...partial,
  }
  if ('value' in partial) {
    next.value = partial.value
  } else {
    next.value = valueForSpec(nextSpec, props.modelValue.value)
  }
  emitUpdate(cleanPredicate(next))
}

function updateProtocolField(fieldPath: string) {
  const field = fieldPath.trim()
  if (!isRuleConditionFieldPath(field)) return
  const fieldSpec = protocolFieldForPath(field)
  const operators = operatorsForField(fieldSpec)
  const effectiveOperators = operators.length ? operators : activePreset.value.operators
  const currentOp = props.modelValue.op || activePreset.value.defaultOp
  const op = effectiveOperators.includes(currentOp) ? currentOp : effectiveOperators[0] || 'eq'
  const nextSpec = buildValueInputSpec({
    fieldPath: field,
    field: fieldSpec,
    operator: op,
  })
  emitUpdate(cleanPredicate({
    field,
    op,
    value: valueForSpec(nextSpec, props.modelValue.value, valueInputSpec.value),
  }))
}

function updateOperator(op: string) {
  const field = selectedFieldPath.value
  const nextSpec = buildValueInputSpec({
    fieldPath: field,
    field: selectedProtocolField.value,
    operator: op,
  })
  emitUpdate(cleanPredicate({
    field,
    op,
    value: valueForSpec(nextSpec, props.modelValue.value, valueInputSpec.value),
  }))
}

function updatePredicateValue(value: unknown) {
  updatePredicate({ value })
}

function updateDynamicSuffix(suffix: string) {
  if (!dynamicParts.value) return
  updateProtocolField(buildDynamicFieldPath(
    dynamicParts.value.rootPath,
    suffix,
    dynamicParts.value.indexMode
  ))
}

function updateDynamicIndexMode(indexMode: DynamicIndexMode) {
  if (!dynamicParts.value) return
  updateProtocolField(buildDynamicFieldPath(
    dynamicParts.value.rootPath,
    dynamicParts.value.suffix,
    indexMode
  ))
}

function changePreset(presetId: ConditionPresetId) {
  emitUpdate(applyConditionPreset(props.modelValue, presetId))
}

function updatePresetKeyValue(key: string) {
  emitUpdate(updateConditionPresetKey(props.modelValue, key))
}

function appendChild() {
  const children = [...groupChildren.value, defaultCondition('predicate')]
  updateGroupChildren(children)
}

function updateChild(index: number, child: Condition) {
  const children = [...groupChildren.value]
  children[index] = child
  updateGroupChildren(children)
}

function removeChild(index: number) {
  const children = groupChildren.value.filter((_, childIndex) => childIndex !== index)
  updateGroupChildren(children.length > 0 ? children : [defaultCondition('predicate')])
}

function updateGroupChildren(children: Condition[]) {
  if (kind.value === 'all') {
    emitUpdate({ all: children })
    return
  }
  if (kind.value === 'any') {
    emitUpdate({ any: children })
  }
}

function emitUpdate(value: Condition) {
  emit('update:modelValue', value)
}

function conditionKind(condition: Condition): ConditionKind {
  if (condition.all) return 'all'
  if (condition.any) return 'any'
  if (condition.not) return 'not'
  if (condition.expr !== undefined) return 'expr'
  return 'predicate'
}

function protocolFieldForPath(fieldPath: string) {
  return fieldForPath(combinedProtocolSpec.value, fieldPath)
}

function defaultCondition(nextKind: ConditionKind): Condition {
  switch (nextKind) {
    case 'all':
      return { all: [defaultCondition('predicate')] }
    case 'any':
      return { any: [defaultCondition('predicate')] }
    case 'not':
      return { not: defaultCondition('predicate') }
    case 'expr':
      return { expr: '' }
    default:
      return applyConditionPreset({}, 'path')
  }
}

function valueForSpec(spec: ValueInputSpec, currentValue: unknown, previousSpec?: ValueInputSpec): unknown {
  if (!valueInputNeedsValue(spec)) return undefined
  if (isInvalidJSONLiteralValue(currentValue)) return currentValue
  if (previousSpec && previousSpec.editor !== spec.editor) return defaultValueForSpec(spec)
  return currentValue === undefined
    ? defaultValueForSpec(spec)
    : normalizeValueForSpec(currentValue, spec)
}

function cleanPredicate(
  condition: Condition,
  options: { omitInvalidJSONLiteral?: boolean } = {}
): Condition {
  const rawField = condition.field || ''
  const field = isRuleConditionFieldPath(rawField) ? rawField : selectedFieldPath.value
  const next: Condition = {
    field,
    op: condition.op,
  }
  const spec = buildValueInputSpec({
    fieldPath: field || '',
    field: protocolFieldForPath(field || ''),
    operator: condition.op || '',
  })
  const value = normalizeValueForSpec(condition.value, spec)
  if (value !== undefined && (!options.omitInvalidJSONLiteral || !isInvalidJSONLiteralValue(value))) {
    next.value = value
  }
  return next
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function mergeProtocolFieldOptions(
  protocolFields: NonNullable<ProtocolSpec['fields']>,
  sampleFields: SampleRequestField[]
) {
  const result = new Map<string, NonNullable<ProtocolSpec['fields']>[number]>()
  sampleFields.forEach((field) => result.set(field.path, field))
  protocolFields.forEach((field) => {
    if (!result.has(field.path)) result.set(field.path, field)
  })
  return [...result.values()]
}
</script>

<style lang="scss" scoped>
.condition-node {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-bg-base);

  &.nested {
    background: var(--ms-gray-50);
  }
}

.node-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-2);
}

.kind-select {
  width: 140px;
}

.predicate-builder {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.predicate-preset-row {
  display: grid;
  grid-template-columns: minmax(140px, 220px) minmax(180px, 1fr);
  gap: var(--ms-space-2);
  align-items: center;
}

.preset-select {
  width: 100%;
}

.predicate-grid {
  display: grid;
  grid-template-columns: minmax(190px, 1fr) 150px minmax(220px, 1.4fr);
  align-items: start;
  gap: var(--ms-space-2);
}

.field-editor,
.dynamic-field-builder {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

.protocol-field-select {
  width: 100%;
}

.dynamic-field-builder {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
}

.index-mode-select {
  width: 86px;
}

.predicate-field-label {
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
  min-height: 32px;
  padding: 5px 9px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);

  span {
    color: var(--ms-text-tertiary);
    font-size: 11px;
    line-height: 1;
    text-transform: uppercase;
  }

  strong {
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.predicate-errors {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);

  span {
    padding: 3px 8px;
    border-radius: var(--ms-radius-pill);
    color: var(--ms-red-600);
    background: var(--ms-red-50);
    font-size: var(--ms-text-sm);
  }
}

.predicate-warnings {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);

  span {
    padding: 3px 8px;
    border-radius: var(--ms-radius-pill);
    color: var(--ms-amber-600);
    background: var(--ms-amber-50);
    font-size: var(--ms-text-sm);
  }
}

.children-block {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.child-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 64px;
  align-items: start;
  gap: var(--ms-space-2);
}

.inline-tool {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-tertiary);
  background: transparent;
  font-size: 11px;
  line-height: 1.45;
  cursor: pointer;

  &:hover,
  &:focus-visible {
    color: var(--ms-teal-700);
    border-color: var(--ms-teal-200);
    background: var(--ms-teal-50);
  }
}

.payload-popover {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    max-height: 220px;
    overflow: auto;
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-primary);
    background: var(--ms-control-bg);
    font-size: var(--ms-text-sm);
    line-height: 1.5;
    white-space: pre-wrap;
  }
}

.remove-child {
  margin-top: var(--ms-space-1);
}

@media (max-width: 900px) {
  .predicate-preset-row,
  .predicate-grid,
  .dynamic-field-builder,
  .child-row {
    grid-template-columns: 1fr;
  }

  .index-mode-select {
    width: 100%;
  }
}
</style>
