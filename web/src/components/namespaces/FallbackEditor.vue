<template>
  <section class="fallback-editor">
    <header class="editor-header">
      <strong>{{ title }}</strong>
      <el-radio-group :model-value="modelValue.type" size="small" @update:model-value="setType">
        <el-radio-button value="respond">Response</el-radio-button>
        <el-radio-button value="forward">Forward</el-radio-button>
      </el-radio-group>
    </header>

    <div v-if="modelValue.type === 'respond'" class="editor-fields">
      <ProtocolResponseEditor
        :model-value="modelValue.responsePayload"
        :drafts="modelValue.responseFieldDrafts"
        :protocol-spec="protocolSpec"
        :json-editor-height="150"
        @update:model-value="patch({ responsePayload: $event })"
        @update:drafts="patch({ responseFieldDrafts: $event })"
      />
    </div>

    <div v-else class="editor-fields">
      <el-form-item label="Timeout MS">
        <el-input-number
          :model-value="modelValue.forwardTimeoutMs"
          :min="0"
          :max="30000"
          controls-position="right"
          @update:model-value="patch({ forwardTimeoutMs: Number($event || 0) })"
        />
      </el-form-item>
      <div class="forward-target">
        <span>target</span>
        <code>original scheme / host / path / query</code>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import ProtocolResponseEditor from '@/components/rulesets/ProtocolResponseEditor.vue'
import type { NamespaceFallbackForm, NamespaceFallbackType, ProtocolSpec } from '@/types'

const props = defineProps<{
  modelValue: NamespaceFallbackForm
  title: string
  protocol?: string
  protocolSpec?: ProtocolSpec
}>()

const emit = defineEmits<{
  'update:modelValue': [value: NamespaceFallbackForm]
}>()

function setType(value: string | number | boolean | undefined) {
  patch({ type: String(value || 'respond') as NamespaceFallbackType })
}

function patch(values: Partial<NamespaceFallbackForm>) {
  emit('update:modelValue', {
    ...props.modelValue,
    ...values,
  })
}
</script>

<style lang="scss" scoped>
.fallback-editor {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding: var(--ms-space-4);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);

  strong {
    min-width: 0;
    overflow: hidden;
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    font-weight: var(--ms-font-bold);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.editor-fields {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.forward-target {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-panel-bg);

  span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@media (max-width: 760px) {
  .editor-header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
