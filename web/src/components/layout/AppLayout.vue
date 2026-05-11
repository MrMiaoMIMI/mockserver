<template>
  <div class="app-layout" :class="{ 'is-collapsed': collapsed }">
    <aside class="layout-sidebar">
      <AppSidebar :collapse="collapsed" />
    </aside>
    <div class="layout-main">
      <header class="layout-header">
        <AppHeader :collapse="collapsed" @toggle-sidebar="toggleSidebar" />
      </header>
      <main class="layout-content">
        <router-view v-slot="{ Component }">
          <transition name="page-shift" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'

const SIDEBAR_KEY = 'mockserver_sidebar_collapsed'
const collapsed = ref(localStorage.getItem(SIDEBAR_KEY) === 'true')

function toggleSidebar() {
  collapsed.value = !collapsed.value
  localStorage.setItem(SIDEBAR_KEY, String(collapsed.value))
}
</script>

<style lang="scss" scoped>
.app-layout {
  position: relative;
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: linear-gradient(180deg, #f8fbff 0%, var(--ms-canvas) 320px);

  &::before {
    position: absolute;
    inset: 0;
    pointer-events: none;
    content: '';
    background: linear-gradient(90deg, rgba(37, 99, 235, 0.04), transparent 32%);
  }

  &.is-collapsed .layout-sidebar {
    width: var(--ms-sidebar-width-collapsed);
  }
}

.layout-sidebar {
  position: relative;
  z-index: 2;
  width: var(--ms-sidebar-width);
  height: 100%;
  flex-shrink: 0;
  overflow: hidden;
  background: var(--ms-sidebar-bg);
  transition: width var(--ms-transition-slow);
}

.layout-main {
  position: relative;
  z-index: 1;
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.layout-header {
  height: var(--ms-header-height);
  flex-shrink: 0;
  border-bottom: 1px solid var(--ms-header-border);
  background: var(--ms-header-bg);
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(14px);
}

.layout-content {
  position: relative;
  flex: 1;
  min-height: 0;
  padding: var(--ms-space-4);
  overflow: auto;
  overscroll-behavior: contain;
}

.page-shift-enter-active,
.page-shift-leave-active {
  transition:
    opacity var(--ms-transition-normal),
    transform var(--ms-transition-normal),
    filter var(--ms-transition-normal);
}

.page-shift-enter-from {
  opacity: 0;
  filter: blur(4px);
  transform: translateY(10px);
}

.page-shift-leave-to {
  opacity: 0;
  filter: blur(2px);
  transform: translateY(-6px);
}

@media (max-width: 860px) {
  .layout-sidebar {
    width: var(--ms-sidebar-width-collapsed);
  }

  .layout-content {
    padding: var(--ms-space-3);
  }
}
</style>
