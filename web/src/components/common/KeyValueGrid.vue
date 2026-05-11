<template>
  <div class="key-value-grid">
    <div v-for="item in items" :key="item.label" :class="['kv-item', `is-${item.tone || 'neutral'}`]">
      <span>{{ item.label }}</span>
      <code v-if="item.code" :title="String(item.value)">{{ item.value }}</code>
      <strong v-else :title="String(item.value)">{{ item.value }}</strong>
      <small v-if="item.caption">{{ item.caption }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface KeyValueItem {
  label: string
  value: string | number
  caption?: string
  code?: boolean
  tone?: 'neutral' | 'ok' | 'warn' | 'danger' | 'accent' | 'primary'
}

defineProps<{
  items: KeyValueItem[]
}>()
</script>

<style lang="scss" scoped>
.key-value-grid {
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: var(--ms-space-2);
}

.kv-item {
  min-width: 0;
  padding: var(--ms-space-2);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-panel-bg-soft);

  span,
  small {
    display: block;
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong,
  code {
    display: block;
    min-width: 0;
    margin-top: 2px;
    overflow: hidden;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-base);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    font-family: var(--ms-font-mono);
  }

  small {
    margin-top: 2px;
    font-weight: var(--ms-font-medium);
  }
}

.is-ok strong,
.is-ok code {
  color: var(--ms-green-600);
}

.is-warn strong,
.is-warn code {
  color: var(--ms-amber-600);
}

.is-danger strong,
.is-danger code {
  color: var(--ms-red-600);
}

.is-accent strong,
.is-accent code {
  color: var(--ms-blue-500);
}

.is-primary strong,
.is-primary code {
  color: var(--ms-teal-700);
}
</style>
