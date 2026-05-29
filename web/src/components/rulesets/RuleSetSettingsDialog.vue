<template>
  <el-dialog
    :model-value="modelValue"
    :title="ruleSet?.id ? 'Edit ruleset' : 'Create ruleset'"
    width="min(900px, calc(100vw - 32px))"
    append-to-body
    destroy-on-close
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" class="settings-form">
      <section class="dialog-section">
        <div class="section-title">Identity</div>
        <div class="form-grid">
          <el-form-item label="Name">
            <el-input v-model="form.name" placeholder="order status mock" />
          </el-form-item>
          <el-form-item label="Protocol">
            <el-select v-model="form.protocol" filterable placeholder="protocol">
              <el-option
                v-for="protocol in protocolOptions"
                :key="protocol.value"
                :label="protocol.label"
                :value="protocol.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Namespace">
            <el-select
              v-model="form.namespace"
              filterable
              placeholder="namespace"
              :loading="store.loading"
            >
              <el-option
                v-for="namespace in namespaceOptions"
                :key="namespace.name"
                :label="namespaceLabel(namespace)"
                :value="namespace.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Enabled">
            <el-switch v-model="form.enabled" active-text="Enabled" inactive-text="Disabled" />
          </el-form-item>
        </div>
      </section>

      <el-alert v-if="validationIssues.length" type="warning" :closable="false" class="validation-alert">
        <ul>
          <li v-for="issue in validationIssues" :key="issue">{{ issue }}</li>
        </ul>
      </el-alert>

      <section class="dialog-section selector-builder">
        <div class="section-title">Selector</div>
        <div v-if="!selectorRows.length" class="empty-selector">At least one selector condition is required.</div>
        <div v-for="row in selectorRows" :key="row.id" class="selector-row">
          <div class="selector-field-editor">
            <el-select
              :model-value="selectorFieldSelectValue(row)"
              class="field-select"
              filterable
              :allow-create="hasDynamicSelectors"
              default-first-option
              placeholder="field"
              @update:model-value="handleSelectorFieldChange(row, $event)"
            >
              <el-option
                v-for="option in selectorFieldOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
                :disabled="isFieldUsed(option.value, row.id)"
              />
            </el-select>
            <div v-if="selectorDynamicParts(row)" class="selector-dynamic-field">
              <el-input
                :model-value="selectorDynamicParts(row)?.suffix"
                size="small"
                :placeholder="selectorDynamicParts(row)?.placeholder"
                @update:model-value="updateSelectorDynamicSuffix(row, $event)"
              >
                <template #prepend>{{ selectorDynamicParts(row)?.label }}</template>
              </el-input>
              <el-select
                v-if="selectorDynamicParts(row)?.indexed"
                :model-value="selectorDynamicParts(row)?.indexMode"
                size="small"
                class="selector-index-mode"
                @update:model-value="updateSelectorDynamicIndexMode(row, $event)"
              >
                <el-option label="first" value="first" />
                <el-option label="any" value="any" />
              </el-select>
            </div>
          </div>
          <el-select
            v-model="row.op"
            class="op-select"
            placeholder="op"
            @change="handleSelectorOperatorChange(row)"
          >
            <el-option
              v-for="op in selectorOperatorOptions(row)"
              :key="op"
              :label="operatorLabel(op)"
              :value="op"
            />
          </el-select>
          <SmartValueInput
            class="value-editor"
            :model-value="row.value"
            :spec="selectorValueSpec(row)"
            @update:model-value="row.value = $event"
          />
          <div class="selector-row-actions">
            <el-popover
              placement="top-end"
              trigger="hover"
              :width="420"
              popper-class="selector-payload-popover"
            >
              <template #reference>
                <button class="inline-tool" type="button">
                  <el-icon><InfoFilled /></el-icon>
                  <span>Payload</span>
                </button>
              </template>
              <div class="payload-popover">
                <span>selector payload</span>
                <code>{{ formatSelectorCondition(row) }}</code>
              </div>
            </el-popover>
            <el-button link type="danger" @click="removeSelectorRow(row.id)">Delete</el-button>
          </div>
        </div>
        <el-button
          class="add-selector"
          :disabled="!availableSelectorField"
          @click="addSelectorRow"
        >
          Add selector condition
        </el-button>
      </section>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">Cancel</el-button>
      <el-button type="primary" :loading="store.saving" @click="submit">Save</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { InfoFilled } from '@element-plus/icons-vue'
import SmartValueInput from '@/components/rulesets/SmartValueInput.vue'
import { useMockserverStore } from '@/store'
import type { Condition, NamespaceConfig, ProtocolSelectorSpec, RuleSet, Selector } from '@/types'
import {
  buildDynamicFieldPath,
  dynamicFieldParts,
  fieldForPath,
  operatorsForField,
  type DynamicIndexMode,
} from '@/utils/protocolFields'
import {
  buildValueInputSpec,
  defaultValueForSpec,
  isInvalidJSONLiteralValue,
  isSmartValueEmpty,
  normalizeValueForSpec,
  valueInputNeedsValue,
  type ValueInputSpec,
} from '@/utils/valueInputSpec'
import { operatorLabel } from '@/utils/ruleAuthoring'

const props = defineProps<{
  modelValue: boolean
  ruleSet: RuleSet | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  saved: [ruleSet: RuleSet]
}>()

interface SelectorRow {
  id: string
  field: string
  op: string
  value: unknown
}

const store = useMockserverStore()

const form = reactive({
  name: '',
  enabled: true,
  protocol: 'http',
  namespace: 'default',
})
const selectorRows = ref<SelectorRow[]>([])
const validationIssues = computed(() => validateSettings())
const currentProtocolSpec = computed(() => store.protocolMap.get(form.protocol))
const protocolOptions = computed(() => {
  const options = store.protocols.map((protocol) => ({
    label: protocol.name.toUpperCase(),
    value: protocol.name,
  }))
  if (form.protocol && !options.some((option) => option.value === form.protocol)) {
    options.unshift({ label: form.protocol.toUpperCase(), value: form.protocol })
  }
  return options.length ? options : [{ label: 'HTTP', value: 'http' }]
})
const selectorFieldOptions = computed(() => {
  return (currentProtocolSpec.value?.selectors || []).map((selector) => ({
    label: selector.path,
    value: selector.path,
  }))
})
const hasDynamicSelectors = computed(() => {
  return Boolean(currentProtocolSpec.value?.selectors?.some((selector) => selector.dynamic_path))
})
const availableSelectorField = computed(() => {
  const nextStaticField = selectorFieldOptions.value.find((option) => !isFieldUsed(option.value))?.value
  if (nextStaticField) return nextStaticField
  return currentProtocolSpec.value?.selectors?.find((selector) => selector.dynamic_path)?.path || ''
})
const namespaceOptions = computed(() => {
  const items = [...store.namespaces]
  const current = form.namespace.trim()
  if (current && !items.some((item) => item.name === current)) {
    items.unshift({
      name: current,
      policies: {
        [form.protocol || 'http']: {
          ruleset_miss_action: {
            type: 'forward',
            forward: { timeout_ms: 5000 },
          },
          rule_miss_action: {
            type: 'forward',
            forward: { timeout_ms: 5000 },
          },
        },
      },
    })
  }
  return items
})

watch(
  () => [props.modelValue, props.ruleSet] as const,
  () => {
    resetForm()
    if (props.modelValue) {
      void Promise.all([store.fetchNamespaces(), store.fetchProtocols()])
    }
  },
  { immediate: true, deep: true }
)

watch(
  () => form.protocol,
  () => {
    if (!currentProtocolSpec.value) return
    selectorRows.value = selectorRows.value.filter((row) => selectorSpec(row.field))
  }
)

function resetForm() {
  const ruleSet = props.ruleSet
  form.name = ruleSet?.name || ''
  form.enabled = ruleSet?.enabled ?? true
  form.protocol = ruleSet?.protocol || 'http'
  form.namespace = ruleSet?.namespace || 'default'
  selectorRows.value = selectorToRows(ruleSet?.selector)
}

function addSelectorRow() {
  const nextField = availableSelectorField.value
  if (!nextField) return
  const op = defaultSelectorOperator(nextField)
  const spec = buildSelectorValueSpec(nextField, op)
  selectorRows.value.push({
    id: createRowId(),
    field: nextField,
    op,
    value: defaultValueForSpec(spec),
  })
}

function removeSelectorRow(id: string) {
  selectorRows.value = selectorRows.value.filter((row) => row.id !== id)
}

async function submit() {
  if (validationIssues.value.length) {
    ElMessage.error(validationIssues.value[0])
    return
  }
  const name = form.name.trim()
  const payload: RuleSet = {
    id: props.ruleSet?.id || '',
    name,
    enabled: form.enabled,
    protocol: form.protocol,
    namespace: form.namespace.trim() || 'default',
    selector: rowsToSelector(selectorRows.value),
    rules: props.ruleSet?.rules || [],
    version: props.ruleSet?.version,
  }
  const saved = await store.saveDraft(payload)
  emit('saved', saved)
  emit('update:modelValue', false)
}

function selectorToRows(selector?: Selector): SelectorRow[] {
  return (selector?.all || [])
    .filter((condition) => condition.field && condition.op)
    .map((condition) => ({
      id: createRowId(),
      field: condition.field || '',
      op: condition.op || '',
      value: condition.value,
    }))
}

function rowsToSelector(rows: SelectorRow[]): Selector {
  const all = rows
    .map(rowToCondition)
    .filter((condition): condition is Condition => Boolean(condition))
  return all.length ? { all } : {}
}

function validateSettings() {
  const issues: string[] = []
  if (!form.name.trim()) {
    issues.push('Name is required')
  }
  const duplicateName = store.drafts.find((draft) => {
    if (props.ruleSet?.id === draft.id) return false
    if (normalizeScopeValue(draft.namespace) !== normalizeScopeValue(form.namespace)) return false
    if (normalizeScopeValue(draft.protocol) !== normalizeScopeValue(form.protocol)) return false
    return normalizedName(draft.name) === normalizedName(form.name)
  })
  if (duplicateName) {
    issues.push(`Ruleset name is already used by ${duplicateName.id} in this namespace and protocol`)
  }
  if (!form.protocol.trim()) {
    issues.push('Protocol is required')
  }
  if (!form.namespace.trim()) {
    issues.push('Namespace is required')
  }
  if (store.protocols.length && !currentProtocolSpec.value) {
    issues.push(`Protocol ${form.protocol} is not registered`)
  }
  if (!selectorRows.value.length) {
    issues.push('At least one selector condition is required')
  }
  for (const row of selectorRows.value) {
    if (!row.field) {
      issues.push('Selector field is required')
      continue
    }
    if (!currentProtocolSpec.value) continue
    const spec = selectorSpec(row.field)
    if (!spec) {
      issues.push(`${row.field} is not an allowed selector field for protocol ${form.protocol}`)
      continue
    }
    if (!row.op) {
      issues.push(`${row.field} operator is required`)
      continue
    }
    if (!selectorOperatorOptions(row).includes(row.op)) {
      issues.push(`${row.field} does not support selector operator ${row.op}`)
    }
    if (isInvalidJSONLiteralValue(row.value)) {
      issues.push(`${selectorFieldLabel(row.field)} ${row.value.message}`)
      continue
    }
    if (isSmartValueEmpty(row.value, selectorValueSpec(row))) {
      issues.push(`${selectorFieldLabel(row.field)} value is required`)
    }
  }
  return issues
}

function normalizedName(value: string) {
  return value.trim().replace(/\s+/g, ' ').toLowerCase()
}

function normalizeScopeValue(value: string) {
  return value.trim().toLowerCase()
}

function rowToCondition(row: SelectorRow): Condition | null {
  const field = row.field.trim()
  const op = row.op.trim()
  if (!field || !op) return null
  const spec = selectorValueSpec(row)
  if (!valueInputNeedsValue(spec)) {
    return { field, op }
  }
  return {
    field,
    op,
    value: normalizeValueForSpec(row.value, spec),
  }
}

function selectorSpec(field: string): ProtocolSelectorSpec | undefined {
  return currentProtocolSpec.value?.selectors?.find((selector) => {
    if (selector.path === field) return true
    if (!selector.dynamic_path) return false
    return field.startsWith(`${selector.path}.`) || field.startsWith(`${selector.path}[`)
  })
}

function selectorOperatorOptions(row: SelectorRow) {
  const selectorOperators = selectorSpec(row.field)?.operators || []
  if (selectorOperators.length) return selectorOperators
  return operatorsForField(fieldForPath(currentProtocolSpec.value, row.field))
}

function defaultSelectorOperator(field: string) {
  const selectorOperators = selectorSpec(field)?.operators || []
  if (selectorOperators.length) return selectorOperators[0]
  return operatorsForField(fieldForPath(currentProtocolSpec.value, field))[0] || 'eq'
}

function handleSelectorFieldChange(row: SelectorRow, field: string) {
  const previousSpec = selectorValueSpec(row)
  row.field = field.trim()
  const operators = selectorOperatorOptions(row)
  if (!operators.includes(row.op)) {
    row.op = operators[0] || ''
  }
  row.value = valueForSpec(selectorValueSpec(row), row.value, previousSpec)
}

function handleSelectorOperatorChange(row: SelectorRow) {
  row.value = valueForSpec(selectorValueSpec(row), row.value)
}

function selectorFieldSelectValue(row: SelectorRow) {
  const field = selectorFieldSpec(row)
  return field?.dynamic_path ? field.path : row.field
}

function selectorFieldSpec(row: SelectorRow) {
  return fieldForPath(currentProtocolSpec.value, row.field)
}

function selectorValueSpec(row: SelectorRow) {
  return buildSelectorValueSpec(row.field, row.op)
}

function buildSelectorValueSpec(field: string, op: string) {
  return buildValueInputSpec({
    fieldPath: field,
    field: fieldForPath(currentProtocolSpec.value, field),
    operator: op,
    dynamicJSONLiteral: false,
  })
}

function selectorDynamicParts(row: SelectorRow) {
  return dynamicFieldParts(row.field, selectorFieldSpec(row))
}

function updateSelectorDynamicSuffix(row: SelectorRow, suffix: string) {
  const parts = selectorDynamicParts(row)
  if (!parts) return
  handleSelectorFieldChange(row, buildDynamicFieldPath(parts.rootPath, suffix, parts.indexMode))
}

function updateSelectorDynamicIndexMode(row: SelectorRow, indexMode: DynamicIndexMode) {
  const parts = selectorDynamicParts(row)
  if (!parts) return
  handleSelectorFieldChange(row, buildDynamicFieldPath(parts.rootPath, parts.suffix, indexMode))
}

function formatSelectorCondition(row: SelectorRow) {
  return formatJSON(rowToCondition(row) || {})
}

function isFieldUsed(field: string, currentRowId = '') {
  return selectorRows.value.some((row) => row.field === field && row.id !== currentRowId)
}

function selectorFieldLabel(field: string) {
  return selectorFieldOptions.value.find((item) => item.value === field)?.label || field
}

function namespaceLabel(namespace: NamespaceConfig) {
  return namespace.name
}

function valueForSpec(spec: ValueInputSpec, currentValue: unknown, previousSpec?: ValueInputSpec): unknown {
  if (!valueInputNeedsValue(spec)) return undefined
  if (isInvalidJSONLiteralValue(currentValue)) return currentValue
  if (previousSpec && previousSpec.editor !== spec.editor) return defaultValueForSpec(spec)
  return currentValue === undefined
    ? defaultValueForSpec(spec)
    : normalizeValueForSpec(currentValue, spec)
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function createRowId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}
</script>

<style lang="scss" scoped>
.settings-form {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
}

.dialog-section {
  padding: var(--ms-space-4);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);
}

.section-title {
  margin-bottom: var(--ms-space-3);
  color: var(--ms-teal-700);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

.form-grid {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) 120px minmax(180px, 1fr) 120px;
  gap: var(--ms-space-3);
}

.empty-selector {
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
}

.selector-builder {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.selector-row {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) 132px minmax(180px, 1fr) 100px;
  gap: var(--ms-space-2);
  align-items: start;
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);
}

.selector-field-editor,
.selector-dynamic-field {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

.selector-dynamic-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
}

.selector-index-mode {
  width: 86px;
}

.field-select,
.op-select {
  width: 100%;
}

.selector-row-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--ms-space-1);
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

.validation-alert {
  :deep(.el-alert__content) {
    width: 100%;
  }

  ul {
    margin: 0;
    padding-left: var(--ms-space-4);
  }
}

.value-editor {
  min-width: 0;
}

.add-selector {
  align-self: flex-start;
}

@media (max-width: 760px) {
  .form-grid,
  .selector-row,
  .selector-dynamic-field {
    grid-template-columns: 1fr;
  }

  .selector-index-mode,
  .selector-row-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
