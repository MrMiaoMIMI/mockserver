<template>
  <article :class="['metric-card', `is-${tone}`]">
    <span class="metric-label">{{ label }}</span>
    <strong>{{ value }}</strong>
    <small v-if="caption">{{ caption }}</small>
  </article>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string
    value: string | number
    caption?: string
    tone?: 'neutral' | 'ok' | 'warn' | 'danger' | 'accent' | 'primary'
  }>(),
  {
    tone: 'neutral',
  }
)
</script>

<style lang="scss" scoped>
.metric-card {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);
  box-shadow: var(--ms-shadow-xs);

  .metric-label,
  small {
    display: block;
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    display: block;
    margin-top: 3px;
    overflow: hidden;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
    line-height: 1.1;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    margin-top: 3px;
    font-weight: var(--ms-font-medium);
  }

  &.is-ok strong {
    color: var(--ms-green-600);
  }

  &.is-warn strong {
    color: var(--ms-amber-600);
  }

  &.is-danger strong {
    color: var(--ms-red-600);
  }

  &.is-accent strong {
    color: var(--ms-blue-500);
  }

  &.is-primary strong {
    color: var(--ms-teal-700);
  }
}
</style>
