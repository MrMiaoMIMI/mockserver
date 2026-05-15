<template>
  <section class="protocol-response-editor">
    <div class="response-grid">
      <template v-for="field in fields" :key="field.path">
        <el-form-item
          v-if="field.type === 'number'"
          :label="fieldLabel(field)"
          :error="fieldErrors[field.path]"
        >
          <el-input-number
            :model-value="numberValue(field)"
            :min="field.min"
            :max="field.max"
            controls-position="right"
            @update:model-value="updateField(field.path, Number($event || 0))"
          />
        </el-form-item>

        <el-form-item
          v-else-if="field.type === 'bool'"
          :label="fieldLabel(field)"
          :error="fieldErrors[field.path]"
        >
          <el-switch
            :model-value="Boolean(fieldValue(field.path))"
            @update:model-value="updateField(field.path, $event)"
          />
        </el-form-item>

        <el-form-item
          v-else-if="field.type === 'string'"
          :label="fieldLabel(field)"
          :error="fieldErrors[field.path]"
        >
          <el-input
            :model-value="stringValue(field)"
            clearable
            @update:model-value="updateField(field.path, $event)"
          />
        </el-form-item>

        <el-form-item
          v-else-if="isHeaderField(field)"
          :label="fieldLabel(field)"
          :error="fieldErrors[field.path]"
          class="headers-field"
        >
          <div class="headers-editor">
            <div v-for="entry in headerEntries(field)" :key="entry.key" class="header-row">
              <el-input
                :model-value="entry.key"
                placeholder="content-type"
                @update:model-value="updateHeaderKey(field.path, entry.key, $event)"
              />
              <el-input
                :model-value="entry.value"
                placeholder="application/json"
                @update:model-value="updateHeaderValue(field.path, entry.key, $event)"
              />
              <el-button :icon="Delete" circle title="Remove header" @click="removeHeader(field.path, entry.key)" />
            </div>
            <el-button size="small" :icon="Plus" @click="addHeader(field.path)">Add header</el-button>
          </div>
        </el-form-item>

        <el-form-item v-else :label="fieldLabel(field)" :error="fieldErrors[field.path]" class="json-field">
          <JsonEditor
            :model-value="draftValue(field)"
            :min-height="jsonEditorHeight"
            :title="`${field.path} JSON`"
            @update:model-value="updateDraft(field.path, $event)"
          />
        </el-form-item>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Delete, Plus } from '@element-plus/icons-vue'
import JsonEditor from '@/components/common/JsonEditor.vue'
import type { ProtocolFieldSpec, ProtocolSpec } from '@/types'
import {
  defaultResponsePayload,
  getResponsePath,
  isHeaderField,
  prettyJSON,
  responseFields,
  setResponsePath,
  type ResponseFieldDrafts,
} from '@/utils/responseSpec'

const props = withDefaults(
  defineProps<{
    protocolSpec?: ProtocolSpec
    modelValue: Record<string, unknown>
    drafts: ResponseFieldDrafts
    fieldErrors?: Record<string, string>
    jsonEditorHeight?: number
  }>(),
  {
    fieldErrors: () => ({}),
    jsonEditorHeight: 180,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, unknown>]
  'update:drafts': [value: ResponseFieldDrafts]
}>()

const fields = computed(() => responseFields(props.protocolSpec))

function fieldValue(path: string) {
  const value = getResponsePath(props.modelValue, path)
  if (value !== undefined) return value
  return getResponsePath(defaultResponsePayload(props.protocolSpec), path)
}

function numberValue(field: ProtocolFieldSpec) {
  const value = fieldValue(field.path)
  return typeof value === 'number' ? value : Number(field.default || 0)
}

function stringValue(field: ProtocolFieldSpec) {
  const value = fieldValue(field.path)
  return typeof value === 'string' ? value : String(field.default || '')
}

function draftValue(field: ProtocolFieldSpec) {
  return props.drafts[field.path] ?? prettyJSON(fieldValue(field.path))
}

function updateField(path: string, value: unknown) {
  emit('update:modelValue', setResponsePath(props.modelValue, path, value))
}

function updateDraft(path: string, value: string) {
  emit('update:drafts', {
    ...props.drafts,
    [path]: value,
  })
}

function headerEntries(field: ProtocolFieldSpec) {
  const raw = fieldValue(field.path)
  const headers = isRecord(raw) ? raw : {}
  return Object.entries(headers).map(([key, value]) => ({
    key,
    value: headerValueText(value),
  }))
}

function addHeader(path: string) {
  const headers = headerObject(path)
  let key = 'x-mock-header'
  for (let index = 2; key in headers; index += 1) {
    key = `x-mock-header-${index}`
  }
  updateField(path, {
    ...headers,
    [key]: [''],
  })
}

function removeHeader(path: string, key: string) {
  const headers = headerObject(path)
  delete headers[key]
  updateField(path, headers)
}

function updateHeaderKey(path: string, oldKey: string, nextKey: string) {
  const headers = headerObject(path)
  const value = headers[oldKey]
  delete headers[oldKey]
  headers[nextKey.trim()] = value
  updateField(path, headers)
}

function updateHeaderValue(path: string, key: string, value: string) {
  updateField(path, {
    ...headerObject(path),
    [key]: value.split(',').map((item) => item.trim()),
  })
}

function headerObject(path: string): Record<string, unknown> {
  const raw = fieldValue(path)
  return isRecord(raw) ? { ...raw } : {}
}

function headerValueText(value: unknown) {
  if (Array.isArray(value)) return value.map(String).join(', ')
  if (typeof value === 'string') return value
  if (value === undefined || value === null) return ''
  return String(value)
}

function fieldLabel(field: ProtocolFieldSpec) {
  return field.required ? `${field.path} *` : field.path
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
</script>

<style lang="scss" scoped>
.protocol-response-editor {
  width: 100%;
  min-width: 0;
}

.response-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ms-space-3);
}

.response-grid > :deep(.el-form-item) {
  min-width: 0;
}

.response-grid :deep(.el-form-item__content) {
  min-width: 0;
}

.json-field,
.headers-field {
  grid-column: 1 / -1;
}

.json-field :deep(.json-editor) {
  width: 100%;
}

.headers-editor {
  width: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.header-row {
  display: grid;
  grid-template-columns: minmax(140px, 0.42fr) minmax(180px, 1fr) auto;
  gap: var(--ms-space-2);
  align-items: center;
}

@media (max-width: 760px) {
  .response-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .header-row {
    grid-template-columns: minmax(0, 1fr) auto;

    .el-input:first-child {
      grid-column: 1 / -1;
    }
  }
}
</style>
