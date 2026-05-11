<template>
  <span :class="['state-chip', `is-${tone}`, { 'has-dot': dot }]">
    <span v-if="dot" class="chip-dot" />
    <slot>{{ label }}</slot>
  </span>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    label?: string
    tone?: 'neutral' | 'ok' | 'warn' | 'caution' | 'danger' | 'accent' | 'primary'
    dot?: boolean
  }>(),
  {
    tone: 'neutral',
    dot: false,
  }
)
</script>

<style lang="scss" scoped>
.state-chip {
  max-width: 100%;
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 4px 9px;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-secondary);
  background: var(--ms-panel-bg-soft);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chip-dot {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: currentcolor;
}

.is-ok {
  color: var(--ms-green-600);
  background: var(--ms-green-50);
}

.is-warn {
  color: var(--ms-amber-600);
  background: var(--ms-amber-50);
}

.is-caution {
  color: #c2410c;
  background: #fff7ed;
}

.is-danger {
  color: var(--ms-red-600);
  background: var(--ms-red-50);
}

.is-accent {
  color: var(--ms-blue-500);
  background: var(--ms-blue-50);
}

.is-primary {
  color: var(--ms-teal-700);
  background: var(--ms-teal-50);
}
</style>
