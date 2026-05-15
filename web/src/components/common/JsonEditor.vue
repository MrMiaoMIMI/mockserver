<template>
  <div class="json-editor" :class="{ 'is-readonly': readonly }" :style="editorStyle">
    <div v-if="showHeader" class="editor-header">
      <div class="editor-title">
        <span class="status-dot is-on" />
        <span>{{ title || 'JSON' }}</span>
      </div>
      <div class="editor-actions">
        <slot name="actions" />
        <el-button v-if="showFormat" size="small" :icon="MagicStick" :disabled="readonly" @click="formatJson">
          Format
        </el-button>
        <el-button size="small" :icon="CopyDocument" @click="copyJson">Copy</el-button>
        <el-button
          v-if="showExpand"
          size="small"
          :icon="FullScreen"
          @click="expanded = true"
        >
          Expand
        </el-button>
      </div>
    </div>
    <div class="editor-body">
      <div ref="gutterRef" class="line-gutter" aria-hidden="true">
        <span v-for="line in lineNumbers" :key="line">{{ line }}</span>
      </div>
      <el-input
        ref="inputRef"
        :model-value="modelValue"
        class="editor-input"
        :autosize="false"
        :placeholder="placeholder"
        :readonly="readonly"
        resize="none"
        type="textarea"
        @update:model-value="$emit('update:modelValue', $event)"
      />
    </div>

    <el-dialog
      v-if="showExpand"
      v-model="expanded"
      :title="title || 'JSON'"
      width="82vw"
      append-to-body
      class="json-editor-dialog"
    >
      <JsonEditor
        :model-value="modelValue"
        :title="title || 'JSON'"
        :placeholder="placeholder"
        :readonly="readonly"
        :min-height="640"
        :show-expand="false"
        @update:model-value="$emit('update:modelValue', $event)"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, FullScreen, MagicStick } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    modelValue: string
    title?: string
    placeholder?: string
    readonly?: boolean
    minHeight?: number
    showExpand?: boolean
    showFormat?: boolean
  }>(),
  {
    placeholder: 'JSON',
    minHeight: 180,
    showExpand: true,
    showFormat: true,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const expanded = ref(false)
const inputRef = ref<{ textarea?: HTMLTextAreaElement } | null>(null)
const gutterRef = ref<HTMLElement | null>(null)
const editorStyle = computed(() => ({
  minHeight: `${props.minHeight}px`,
}))
const showHeader = computed(() => Boolean(props.title) || props.showExpand)
const lineNumbers = computed(() => {
  const count = Math.max(1, props.modelValue.split('\n').length)
  return Array.from({ length: count }, (_, index) => index + 1)
})

function formatJson() {
  if (props.readonly) return
  try {
    emit('update:modelValue', JSON.stringify(JSON.parse(props.modelValue), null, 2))
  } catch (error) {
    ElMessage.error(`JSON is not valid: ${toErrorMessage(error)}`)
  }
}

async function copyJson() {
  try {
    await navigator.clipboard.writeText(props.modelValue)
    ElMessage.success('JSON copied')
  } catch {
    ElMessage.error('Copy failed')
  }
}

function syncGutterScroll() {
  if (!inputRef.value?.textarea || !gutterRef.value) return
  gutterRef.value.scrollTop = inputRef.value.textarea.scrollTop
}

function bindTextareaScroll() {
  const textarea = inputRef.value?.textarea
  if (!textarea) return
  textarea.removeEventListener('scroll', syncGutterScroll)
  textarea.addEventListener('scroll', syncGutterScroll)
  syncGutterScroll()
}

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}

onMounted(() => {
  void nextTick(bindTextareaScroll)
})

onBeforeUnmount(() => {
  inputRef.value?.textarea?.removeEventListener('scroll', syncGutterScroll)
})

watch(
  () => props.modelValue,
  () => {
    void nextTick(bindTextareaScroll)
  }
)
</script>

<style lang="scss" scoped>
.json-editor {
  width: 100%;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: #ffffff;
  box-shadow: var(--ms-shadow-xs);
}

.editor-header {
  min-height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: 0 var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);
  color: var(--ms-text-primary);
  background: var(--ms-panel-bg-soft);
}

.editor-title {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-2);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
}

.editor-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-2);
  flex-wrap: wrap;
  justify-content: flex-end;
}

.editor-body {
  min-height: 0;
  flex: 1;
  display: grid;
  grid-template-columns: 46px minmax(0, 1fr);
  overflow: hidden;
  background: #fbfdff;
}

.line-gutter {
  min-height: 0;
  padding: 10px 8px;
  overflow: hidden;
  border-right: 1px solid #d8e1ec;
  color: #8a98aa;
  background: #f2f6fa;
  font-family: var(--ms-font-mono);
  font-size: 12px;
  line-height: 1.72;
  text-align: right;
  user-select: none;

  span {
    display: block;
    height: 1.72em;
  }
}

.editor-input {
  min-height: 0;
  flex: 1;
  display: flex;

  :deep(.el-textarea__inner) {
    flex: 1;
    height: 100% !important;
    min-height: 160px !important;
    overflow: auto;
    border: none;
    border-radius: 0;
    box-shadow: none;
    color: #102033;
    background:
      linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
      #fbfdff;
    background-size: 100% 26px;
    font-family: var(--ms-font-mono);
    font-size: 13px;
    line-height: 1.72;
    tab-size: 2;

    &::placeholder {
      color: #8a98aa;
    }

    &:focus {
      box-shadow: inset 0 0 0 2px rgba(8, 145, 178, 0.2);
    }
  }
}

.is-readonly {
  :deep(.el-textarea__inner) {
    cursor: default;
  }
}

:global(.json-editor-dialog .el-dialog__body) {
  height: min(72vh, 760px);
  display: flex;
  padding-top: var(--ms-space-2);
}

:global(.json-editor-dialog .json-editor) {
  width: 100%;
}
</style>
