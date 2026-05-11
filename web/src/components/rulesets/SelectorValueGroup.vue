<template>
  <div class="selector-value-group">
    <template v-if="normalizedValues.length">
      <el-popover
        v-for="value in visibleValues"
        :key="value"
        placement="top-start"
        trigger="hover"
        :width="520"
        popper-class="selector-value-popover"
      >
        <template #reference>
          <button
            class="selector-chip"
            type="button"
            :aria-label="`copy selector value ${value}`"
            @click.stop="copyValue(value)"
          >
            <code>{{ value }}</code>
            <el-icon><DocumentCopy /></el-icon>
          </button>
        </template>
        <div class="value-popover">
          <span>full value</span>
          <code>{{ value }}</code>
          <small>Click the chip to copy</small>
        </div>
      </el-popover>

      <el-popover
        v-if="hiddenValues.length"
        placement="bottom-start"
        trigger="click"
        :width="560"
        popper-class="selector-value-popover"
      >
        <template #reference>
          <button class="more-chip" type="button">+{{ hiddenValues.length }}</button>
        </template>
        <div class="hidden-values">
          <header>
            <span>hidden values</span>
            <strong>{{ hiddenValues.length }}</strong>
          </header>
          <button
            v-for="value in hiddenValues"
            :key="value"
            class="hidden-value"
            type="button"
            @click="copyValue(value)"
          >
            <code>{{ value }}</code>
            <el-icon><DocumentCopy /></el-icon>
          </button>
        </div>
      </el-popover>
    </template>
    <small v-else class="empty-text">{{ emptyText }}</small>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    values?: string[]
    maxVisible?: number
    emptyText?: string
  }>(),
  {
    maxVisible: 2,
    emptyText: 'any',
  }
)

const normalizedValues = computed(() => {
  return Array.from(new Set((props.values || []).map((item) => item.trim()).filter(Boolean)))
})
const visibleValues = computed(() => normalizedValues.value.slice(0, props.maxVisible))
const hiddenValues = computed(() => normalizedValues.value.slice(props.maxVisible))

async function copyValue(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    ElMessage.success('Selector value copied')
  } catch {
    ElMessage.error('Copy failed. Copy the full value from the popover manually.')
  }
}
</script>

<style lang="scss" scoped>
.selector-value-group {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);
}

.selector-chip,
.more-chip,
.hidden-value {
  min-width: 0;
  border: none;
  cursor: pointer;
  font: inherit;
}

.selector-chip {
  max-width: 100%;
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 4px 8px;
  border-radius: var(--ms-radius-sm);
  color: var(--ms-text-secondary);
  background: var(--ms-control-bg);
  transition:
    color var(--ms-transition-fast),
    background var(--ms-transition-fast);

  code {
    min-width: 0;
    max-width: min(320px, 100%);
    overflow: hidden;
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .el-icon {
    flex: 0 0 auto;
    color: var(--ms-text-tertiary);
    font-size: 14px;
    opacity: 0;
    transition: opacity var(--ms-transition-fast);
  }

  &:hover,
  &:focus-visible {
    color: var(--ms-teal-700);
    background: var(--ms-teal-50);

    .el-icon {
      opacity: 1;
    }
  }
}

.more-chip {
  padding: 4px 9px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-teal-700);
  background: var(--ms-teal-50);
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
}

.empty-text {
  padding: 4px 8px;
  border-radius: var(--ms-radius-sm);
  color: var(--ms-text-tertiary);
  background: var(--ms-control-bg);
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
}

.value-popover {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  span,
  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }

  span {
    font-family: var(--ms-font-display);
    text-transform: uppercase;
  }

  code {
    max-height: 160px;
    overflow: auto;
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-primary);
    background: var(--ms-control-bg);
    font-size: var(--ms-text-sm);
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
  }
}

.hidden-values {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  max-height: 340px;
  overflow: auto;

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-3);
  }

  header span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  header strong {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
  }
}

.hidden-value {
  width: 100%;
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border-radius: var(--ms-radius-md);
  color: var(--ms-text-primary);
  background: var(--ms-control-bg);
  text-align: left;

  code {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: var(--ms-text-sm);
    line-height: 1.5;
  }

  .el-icon {
    flex: 0 0 auto;
    margin-top: 2px;
    color: var(--ms-text-tertiary);
  }

  &:hover,
  &:focus-visible {
    color: var(--ms-teal-700);
    background: var(--ms-teal-50);
  }
}
</style>
