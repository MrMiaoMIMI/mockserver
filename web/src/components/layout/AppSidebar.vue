<template>
  <div class="sidebar" :class="{ 'is-collapsed': collapse }">
    <div class="sidebar-brand">
      <div class="brand-mark">
        <el-icon><Connection /></el-icon>
      </div>
      <transition name="brand-copy">
        <div v-if="!collapse" class="brand-copy">
          <strong>MockServer</strong>
          <span>traffic console</span>
        </div>
      </transition>
    </div>

    <nav class="sidebar-nav">
      <RouterLink
        v-for="item in navItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: item.active }"
        :to="item.path"
        :title="collapse ? item.label : undefined"
      >
        <el-icon><component :is="item.icon" /></el-icon>
        <transition name="brand-copy">
          <span v-if="!collapse" class="nav-label">{{ item.label }}</span>
        </transition>
        <transition name="brand-copy">
          <span v-if="!collapse && item.count !== undefined" class="nav-count">{{ item.count }}</span>
        </transition>
      </RouterLink>
    </nav>

    <div class="sidebar-footer">
      <div class="signal-card">
        <span class="status-dot is-on" />
        <transition name="brand-copy">
          <div v-if="!collapse" class="signal-copy">
            <strong>admin</strong>
            <span>/mockserver/api/v1</span>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, markRaw } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { Box, Connection, Tickets, TrendCharts } from '@element-plus/icons-vue'
import { useMockserverStore } from '@/store'

defineProps<{
  collapse: boolean
}>()

const route = useRoute()
const store = useMockserverStore()

const navItems = computed(() => [
  {
    path: '/rulesets',
    label: 'Rulesets',
    icon: markRaw(Tickets),
    count: store.drafts.length,
    active: route.path.startsWith('/rulesets'),
  },
  {
    path: '/namespaces',
    label: 'Namespaces',
    icon: markRaw(Box),
    count: store.namespaces.length,
    active: route.path.startsWith('/namespaces'),
  },
  {
    path: '/dashboard',
    label: 'Traffic',
    icon: markRaw(TrendCharts),
    count: store.trafficTotal,
    active: route.path.startsWith('/dashboard'),
  },
])
</script>

<style lang="scss" scoped>
.sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  color: var(--ms-sidebar-text);
  background:
    linear-gradient(180deg, rgba(37, 99, 235, 0.12), transparent 260px),
    var(--ms-sidebar-bg);
}

.sidebar-brand {
  height: var(--ms-header-height);
  display: flex;
  align-items: center;
  gap: var(--ms-space-3);
  padding: 0 var(--ms-space-4);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.brand-mark {
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: var(--ms-radius-lg);
  color: var(--ms-text-inverse);
  background: linear-gradient(135deg, #2563eb, #06b6d4 82%);
  box-shadow: 0 16px 30px rgba(15, 23, 42, 0.32);
}

.brand-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  line-height: 1.1;

  strong {
    color: #ffffff;
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
    font-weight: var(--ms-font-bold);
  }

  span {
    margin-top: 3px;
    color: var(--ms-sidebar-text-muted);
    font-size: var(--ms-text-sm);
  }
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  flex: 1;
  padding: var(--ms-space-4) var(--ms-space-3);
}

.nav-item {
  min-width: 0;
  height: 42px;
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--ms-space-2);
  padding: 0 var(--ms-space-2);
  border: 1px solid transparent;
  border-radius: var(--ms-radius-lg);
  color: var(--ms-sidebar-text);
  transition:
    color var(--ms-transition-fast),
    background var(--ms-transition-fast),
    border-color var(--ms-transition-fast);

  .el-icon {
    justify-self: center;
    font-size: 18px;
  }

  &:hover {
    color: #ffffff;
    background: var(--ms-sidebar-bg-hover);
  }

  &.active {
    border-color: rgba(96, 165, 250, 0.42);
    color: #ffffff;
    background: rgba(37, 99, 235, 0.24);
  }
}

.nav-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: var(--ms-font-semibold);
  white-space: nowrap;
}

.nav-count {
  min-width: 28px;
  padding: 3px 8px;
  border-radius: var(--ms-radius-pill);
  color: #dbeafe;
  background: rgba(255, 255, 255, 0.11);
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
  text-align: center;
}

.sidebar-footer {
  padding: var(--ms-space-3);
}

.signal-card {
  min-height: 44px;
  display: flex;
  align-items: center;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--ms-radius-lg);
  background: rgba(255, 255, 255, 0.055);
}

.signal-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;

  strong {
    color: #ffffff;
    font-size: var(--ms-text-sm);
    text-transform: uppercase;
  }

  span {
    overflow: hidden;
    color: var(--ms-sidebar-text-muted);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.is-collapsed {
  .sidebar-brand,
  .sidebar-footer {
    padding-right: var(--ms-space-3);
    padding-left: var(--ms-space-3);
  }

  .brand-mark,
  .signal-card {
    margin: 0 auto;
  }

  .nav-item {
    grid-template-columns: 1fr;
    padding: 0;
  }
}

@media (max-width: 860px) {
  .sidebar-brand,
  .sidebar-footer {
    padding-right: var(--ms-space-3);
    padding-left: var(--ms-space-3);
  }

  .brand-copy,
  .nav-label,
  .nav-count,
  .signal-copy {
    display: none;
  }

  .brand-mark,
  .signal-card {
    margin: 0 auto;
  }

  .nav-item {
    grid-template-columns: 1fr;
    padding: 0;
  }
}

.brand-copy-enter-active,
.brand-copy-leave-active {
  transition: opacity var(--ms-transition-fast), transform var(--ms-transition-fast);
}

.brand-copy-enter-from,
.brand-copy-leave-to {
  opacity: 0;
  transform: translateX(-6px);
}
</style>
