<template>
  <section class="page-container">
    <header v-if="title || description || $slots.actions || $slots.meta" class="page-header">
      <div class="page-heading">
        <div v-if="eyebrow" class="page-eyebrow">{{ eyebrow }}</div>
        <div class="page-title-row">
          <h1 v-if="title">{{ title }}</h1>
          <div v-if="$slots.meta" class="page-meta">
            <slot name="meta" />
          </div>
        </div>
        <p v-if="description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="page-actions">
        <slot name="actions" />
      </div>
    </header>
    <div class="page-body">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  description?: string
  eyebrow?: string
}>()
</script>

<style lang="scss" scoped>
.page-container {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.page-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: var(--ms-space-5);
  flex-shrink: 0;
}

.page-heading {
  min-width: 0;
  max-width: 100%;
}

.page-title-row {
  min-width: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--ms-space-2) var(--ms-space-3);
}

.page-eyebrow {
  margin-bottom: var(--ms-space-1);
  color: var(--ms-teal-700);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

h1 {
  min-width: 0;
  margin: 0;
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-2xl);
  font-weight: var(--ms-font-bold);
  line-height: 1.1;
  overflow-wrap: anywhere;
}

p {
  max-width: 760px;
  margin: var(--ms-space-2) 0 0;
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-base);
  line-height: 1.55;
}

.page-meta {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

.page-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--ms-space-2);
  flex-shrink: 0;
}

.page-body {
  min-height: 0;
  flex: 1;
}

@media (max-width: 900px) {
  .page-header {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .page-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  h1 {
    font-size: var(--ms-text-2xl);
  }
}
</style>
