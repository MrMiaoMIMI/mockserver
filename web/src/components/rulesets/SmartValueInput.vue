<template>
  <div class="smart-value-input" :class="`is-${spec.editor}`">
    <div v-if="spec.editor === 'none'" class="no-value-box">
      <span>No value</span>
    </div>

    <el-input
      v-else-if="spec.editor === 'text'"
      :model-value="textValue"
      :size="size"
      clearable
      :placeholder="spec.placeholder"
      @update:model-value="emitValue"
    />

    <el-input-number
      v-else-if="spec.editor === 'number'"
      :model-value="numberValue"
      :size="size"
      class="number-input"
      controls-position="right"
      :placeholder="spec.placeholder"
      @update:model-value="emitValue"
    />

    <el-radio-group
      v-else-if="spec.editor === 'boolean'"
      :model-value="booleanValue"
      :size="size"
      @update:model-value="emitValue"
    >
      <el-radio-button :value="true">true</el-radio-button>
      <el-radio-button :value="false">false</el-radio-button>
    </el-radio-group>

    <div v-else-if="spec.editor === 'list'" class="list-editor">
      <div class="list-values">
        <el-tag
          v-for="(item, index) in listValues"
          :key="`${index}-${formatInline(item)}`"
          size="small"
          closable
          @close="removeListValue(index)"
        >
          {{ formatInline(item) }}
        </el-tag>
        <small v-if="!listValues.length">empty list</small>
      </div>
      <el-input
        v-model="listDraft"
        :size="size"
        clearable
        :placeholder="spec.placeholder"
        @keyup.enter="addListValues"
        @blur="addListValues"
      >
        <template #append>
          <el-button @click="addListValues">Add</el-button>
        </template>
      </el-input>
    </div>

    <div v-else-if="spec.editor === 'regex'" class="stack-editor">
      <el-input
        :model-value="textValue"
        :size="size"
        clearable
        :placeholder="spec.placeholder"
        @update:model-value="emitRegexValue"
      />
      <p v-if="regexError" class="input-error">{{ regexError }}</p>
    </div>

    <div v-else class="stack-editor">
      <el-input
        :model-value="jsonText"
        :size="size"
        type="textarea"
        :rows="jsonTextareaRows"
        :autosize="jsonTextareaAutosize"
        :placeholder="spec.placeholder"
        @update:model-value="emitJSONValue"
      />
      <p v-if="jsonError" class="input-error">{{ jsonError }}</p>
    </div>

    <div v-if="hasGuidance" class="input-guidance">
      <el-popover
        v-if="spec.helperText"
        placement="top-start"
        trigger="hover"
        :width="300"
        popper-class="smart-value-popover"
      >
        <template #reference>
          <button class="guidance-chip" type="button">
            <el-icon><InfoFilled /></el-icon>
            <span>Help</span>
          </button>
        </template>
        <div class="guidance-popover">{{ spec.helperText }}</div>
      </el-popover>

      <el-popover
        v-if="hasExamples"
        placement="bottom-start"
        trigger="click"
        :width="320"
        popper-class="smart-value-popover"
      >
        <template #reference>
          <button class="guidance-chip" type="button">Examples</button>
        </template>
        <div class="example-row">
          <button
            v-for="example in spec.examples"
            :key="formatExample(example)"
            type="button"
            @click="applyExample(example)"
          >
            {{ formatExample(example) }}
          </button>
        </div>
      </el-popover>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'
import type { ValueInputSpec } from '@/utils/valueInputSpec'
import {
  coerceListItem,
  defaultValueForSpec,
  invalidJSONLiteralValue,
  isInvalidJSONLiteralValue,
  normalizeValueForSpec,
} from '@/utils/valueInputSpec'

const props = withDefaults(
  defineProps<{
    modelValue?: unknown
    spec: ValueInputSpec
    size?: 'small' | 'default' | 'large'
  }>(),
  {
    size: 'small',
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
}>()

const listDraft = ref('')
const jsonText = ref(formatJSON(props.modelValue ?? defaultValueForSpec(props.spec)))
const jsonError = ref('')
const regexError = ref('')

const textValue = computed(() => {
  if (props.modelValue === undefined || props.modelValue === null) return ''
  if (typeof props.modelValue === 'string') return props.modelValue
  return String(props.modelValue)
})
const numberValue = computed(() => {
  if (typeof props.modelValue === 'number') return props.modelValue
  if (typeof props.modelValue === 'string' && props.modelValue.trim()) {
    const parsed = Number(props.modelValue)
    return Number.isFinite(parsed) ? parsed : undefined
  }
  return undefined
})
const booleanValue = computed(() => {
  if (typeof props.modelValue === 'boolean') return props.modelValue
  if (props.modelValue === 'false') return false
  return true
})
const listValues = computed(() => {
  const normalized = normalizeValueForSpec(props.modelValue, props.spec)
  return Array.isArray(normalized) ? normalized : []
})
const hasExamples = computed(() => props.spec.editor !== 'none' && props.spec.examples.length > 0)
const hasGuidance = computed(() => Boolean(props.spec.helperText) || hasExamples.value)
const jsonTextareaRows = computed(() => props.spec.jsonLiteral ? compactJSONLiteralRows.value : 4)
const jsonTextareaAutosize = computed(() => (
  props.spec.jsonLiteral ? { minRows: compactJSONLiteralRows.value, maxRows: 6 } : false
))
const compactJSONLiteralRows = computed(() => {
  if (!props.spec.jsonLiteral) return 4
  if (props.spec.operator === 'in' || props.spec.operator === 'not_in') return 2
  if (['request.body', 'request.req', 'request.value'].includes(props.spec.fieldPath)) return 2
  return 1
})

watch(
  () => [props.modelValue, props.spec.editor] as const,
  () => {
    if (props.spec.editor === 'json') {
      if (isInvalidJSONLiteralValue(props.modelValue)) {
        jsonText.value = props.modelValue.raw
        jsonError.value = props.modelValue.message
      } else {
        jsonText.value = formatJSON(props.modelValue ?? defaultValueForSpec(props.spec))
        jsonError.value = ''
      }
    }
    if (props.spec.editor === 'none' && props.modelValue !== undefined) {
      emit('update:modelValue', undefined)
    }
  },
  { deep: true }
)

function emitValue(value: unknown) {
  emit('update:modelValue', value)
}

function emitRegexValue(value: string) {
  if (value.trim()) {
    try {
      new RegExp(value)
      regexError.value = ''
    } catch (error) {
      regexError.value = error instanceof Error ? error.message : String(error)
    }
  } else {
    regexError.value = ''
  }
  emit('update:modelValue', value)
}

function emitJSONValue(value: string) {
  jsonText.value = value
  try {
    const parsed = JSON.parse(value)
    jsonError.value = ''
    emit('update:modelValue', parsed)
  } catch (error) {
    jsonError.value = jsonLiteralErrorMessage(value, error)
    emit('update:modelValue', invalidJSONLiteralValue(value, jsonError.value))
  }
}

function addListValues() {
  const values = listDraft.value
    .split(/[,;\n]/)
    .map((item) => coerceListItem(item, props.spec))
    .filter((item) => !(typeof item === 'string' && !item.trim()))
  if (!values.length) return
  emit('update:modelValue', uniqueValues([...listValues.value, ...values]))
  listDraft.value = ''
}

function removeListValue(index: number) {
  emit('update:modelValue', listValues.value.filter((_, itemIndex) => itemIndex !== index))
}

function applyExample(example: unknown) {
  if (props.spec.editor === 'list') {
    emit('update:modelValue', Array.isArray(example) ? example : [example])
    return
  }
  if (props.spec.editor === 'json') {
    jsonText.value = formatJSON(example)
    jsonError.value = ''
  }
  if (props.spec.editor === 'regex') {
    regexError.value = ''
  }
  emit('update:modelValue', example)
}

function uniqueValues(values: unknown[]) {
  const seen = new Set<string>()
  return values.filter((value) => {
    const key = formatInline(value)
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

function formatInline(value: unknown) {
  if (typeof value === 'string') return value
  return JSON.stringify(value)
}

function formatExample(value: unknown) {
  if (props.spec.editor === 'json') return JSON.stringify(value)
  return formatInline(value)
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function jsonLiteralErrorMessage(value: string, error: unknown) {
  const trimmed = value.trim()
  if (!trimmed) return 'Enter a JSON value.'
  if (/^[A-Za-z_][A-Za-z0-9_ -]*$/.test(trimmed)) {
    return `Invalid JSON value. String values must use double quotes, for example "${trimmed}".`
  }
  const detail = error instanceof Error ? error.message : String(error)
  return `Invalid JSON value: ${detail}`
}
</script>

<style lang="scss" scoped>
.smart-value-input {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

.number-input {
  width: 100%;
}

.no-value-box {
  min-height: 32px;
  display: flex;
  align-items: center;
  padding: 6px 9px;
  border: 1px dashed var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);

  span {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }
}

.list-editor,
.stack-editor {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

.list-values {
  min-height: 32px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 5px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-panel-bg);

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.input-guidance {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);
}

.guidance-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 7px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-tertiary);
  background: transparent;
  font-size: 11px;
  line-height: 1.5;
  cursor: pointer;

  &:hover,
  &:focus-visible {
    color: var(--ms-teal-700);
    border-color: var(--ms-teal-200);
    background: var(--ms-teal-50);
  }
}

.guidance-popover {
  color: var(--ms-text-secondary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
}

.example-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);

  button {
    max-width: 160px;
    overflow: hidden;
    padding: 3px 7px;
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-sm);
    color: var(--ms-teal-700);
    background: var(--ms-teal-50);
    font-family: var(--ms-font-mono);
    font-size: 11px;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: pointer;
  }
}

.input-error {
  margin: 0;
  color: var(--ms-red-600);
  font-size: var(--ms-text-sm);
  line-height: 1.35;
}
</style>
