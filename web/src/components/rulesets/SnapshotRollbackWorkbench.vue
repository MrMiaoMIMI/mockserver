<template>
  <div class="snapshot-rollback-workbench">
    <header class="rollback-hero">
      <div>
        <span>rollback workbench</span>
        <strong>{{ selectedItem ? selectedItem.version : 'Select a snapshot' }}</strong>
        <small>{{ selectedItem ? selectedItem.id : ruleSetId }}</small>
      </div>
      <div class="event-chip" :class="{ active: hasSimulationEvent }">
        <span>{{ hasSimulationEvent ? 'event attached' : 'diff only' }}</span>
        <small>{{ eventHint }}</small>
      </div>
    </header>

    <el-alert
      v-if="previewError"
      class="workflow-alert"
      type="error"
      show-icon
      :closable="false"
      :title="previewError"
    />
    <el-alert
      v-if="rollbackError"
      class="workflow-alert"
      type="error"
      show-icon
      :closable="false"
      :title="rollbackError"
    />
    <el-alert
      v-if="successLabel"
      class="workflow-alert"
      type="success"
      show-icon
      :closable="false"
      :title="successLabel"
    />

    <el-empty
      v-if="!historyItems.length"
      class="snapshot-empty"
      description="This ruleset has no published snapshots yet"
    />

    <template v-else>
      <section class="history-section">
        <header class="section-heading">
          <div>
            <span>snapshot history</span>
            <strong>{{ historyItems.length }} versions</strong>
          </div>
        </header>

        <div class="snapshot-list" role="listbox" aria-label="Snapshot history">
          <button
            v-for="snapshot in historyItems"
            :key="snapshot.id"
            type="button"
            class="snapshot-row"
            :class="{ active: selectedSnapshotId === snapshot.id }"
            role="option"
            :aria-selected="selectedSnapshotId === snapshot.id"
            @click="emit('select', snapshot.id)"
            @keydown.enter.prevent="emit('select', snapshot.id)"
            @keydown.space.prevent="emit('select', snapshot.id)"
          >
            <div class="row-title">
              <code :title="snapshot.id">{{ snapshot.id }}</code>
              <strong :title="snapshot.version">{{ snapshot.version }}</strong>
            </div>
            <div class="snapshot-meta">
              <span :title="snapshot.publishedAt">
                <b>published</b>
                {{ snapshot.publishedAt }}
              </span>
              <span :title="snapshot.operator">
                <b>operator</b>
                {{ snapshot.operator }}
              </span>
              <span :title="snapshot.reason">
                <b>reason</b>
                {{ snapshot.reason }}
              </span>
              <span v-if="snapshot.sourceSnapshot" :title="snapshot.sourceSnapshot">
                <b>source</b>
                {{ snapshot.sourceSnapshot }}
              </span>
            </div>
          </button>
        </div>
      </section>

      <section class="rollback-flow">
        <header class="flow-heading">
          <div>
            <span>rollback preview</span>
            <strong>{{ selectedItem ? selectedItem.id : 'No target selected' }}</strong>
          </div>
          <small>{{ selectedItem ? confirmationText : 'Select a snapshot to preview and run rollback here.' }}</small>
        </header>

        <el-input
          :model-value="rollbackReason"
          type="textarea"
          :rows="2"
          resize="none"
          placeholder="rollback reason"
          @update:model-value="updateRollbackReason"
        />

        <div class="flow-actions">
          <el-button
            :loading="previewing"
            :disabled="!selectedSnapshotId"
            @click="emit('preview')"
          >
            Preview impact
          </el-button>
          <el-button
            type="danger"
            :loading="rollingBack"
            :disabled="!selectedSnapshotId"
            @click="emit('rollback')"
          >
            Run rollback
          </el-button>
        </div>

        <div v-if="previewView" class="preview-summary">
          <div class="summary-grid">
            <div>
              <span>current</span>
              <code :title="previewView.currentSnapshot">{{ previewView.currentSnapshot }}</code>
            </div>
            <div>
              <span>target</span>
              <code :title="previewView.targetSnapshot">{{ previewView.targetSnapshot }}</code>
            </div>
            <div>
              <span>state</span>
              <strong :class="{ ok: !previewView.changed }">{{ previewView.changedLabel }}</strong>
            </div>
            <div>
              <span>validation</span>
              <strong :class="{ ok: previewResult?.result.valid }">{{ validationLabel }}</strong>
            </div>
          </div>

          <div class="message-list">
            <span v-for="message in previewView.messages" :key="message">{{ message }}</span>
          </div>

          <section v-if="previewView.ruleSetFieldDiffs.length" class="diff-section">
            <div class="diff-heading">ruleset field diffs</div>
            <div
              v-for="diff in previewView.ruleSetFieldDiffs"
              :key="`${diff.path}-${diff.message}`"
              class="field-diff"
            >
              <header>
                <code :title="diff.path">{{ diff.path }}</code>
                <span :title="diff.message">{{ diff.message }}</span>
              </header>
              <div class="value-pair">
                <div>
                  <b>current</b>
                  <pre :title="diff.current">{{ diff.current }}</pre>
                </div>
                <div>
                  <b>target</b>
                  <pre :title="diff.target">{{ diff.target }}</pre>
                </div>
              </div>
            </div>
          </section>

          <section v-if="previewView.ruleDiffs.length" class="diff-section">
            <div class="diff-heading">rule diffs</div>
            <div v-for="rule in previewView.ruleDiffs" :key="rule.ruleId" class="rule-diff">
              <header>
                <div>
                  <code :title="rule.ruleId">{{ rule.ruleId }}</code>
                  <span :title="rule.message">{{ rule.message }}</span>
                </div>
                <strong :title="rule.changeLabel">{{ rule.changeLabel }}</strong>
              </header>
              <div class="change-flags">
                <span v-if="rule.conditionChanged">condition</span>
                <span v-if="rule.actionChanged">action</span>
                <span v-if="!rule.conditionChanged && !rule.actionChanged">
                  {{ rule.changeType }}
                </span>
              </div>
              <div class="rule-field-diffs">
                <div
                  v-for="diff in rule.fieldDiffs"
                  :key="`${rule.ruleId}-${diff.path}-${diff.message}`"
                  class="field-diff"
                >
                  <header>
                    <code :title="diff.path">{{ diff.path }}</code>
                    <span :title="diff.message">{{ diff.message }}</span>
                  </header>
                  <div class="value-pair">
                    <div>
                      <b>current</b>
                      <pre :title="diff.current">{{ diff.current }}</pre>
                    </div>
                    <div>
                      <b>target</b>
                      <pre :title="diff.target">{{ diff.target }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section v-if="!previewView.hasDiffs" class="no-diff-panel">
            <strong>No structured diffs</strong>
            <span>The target snapshot has no displayable structural differences from the current published version.</span>
          </section>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type {
  PublishedRuleSetSnapshot,
  RollbackPreviewResponse,
  RollbackRuleSetResponse,
} from '@/types'
import {
  buildRollbackPreviewView,
  buildSnapshotHistory,
  rollbackResultLabel,
} from '@/utils/snapshotRollback'

const props = defineProps<{
  ruleSetId: string
  snapshots: PublishedRuleSetSnapshot[]
  selectedSnapshotId: string
  rollbackReason: string
  previewResult: RollbackPreviewResponse | null
  rollbackResult: RollbackRuleSetResponse | null
  previewError: string
  rollbackError: string
  previewing: boolean
  rollingBack: boolean
  hasSimulationEvent: boolean
}>()

const emit = defineEmits<{
  select: [snapshotId: string]
  'update:rollbackReason': [value: string]
  preview: []
  rollback: []
}>()

const historyItems = computed(() => buildSnapshotHistory(props.snapshots))
const selectedItem = computed(() => {
  return historyItems.value.find((snapshot) => snapshot.id === props.selectedSnapshotId) || null
})
const previewView = computed(() => buildRollbackPreviewView(props.previewResult))
const successLabel = computed(() => {
  const label = rollbackResultLabel(props.rollbackResult)
  return label ? `Rollback completed, ${label}` : ''
})
const eventHint = computed(() => {
  return props.hasSimulationEvent
    ? 'preview uses the current simulation payload'
    : 'preview will show config diff without request simulation'
})
const confirmationText = computed(() => {
  return `Rollback ${props.ruleSetId} to ${props.selectedSnapshotId}`
})
const validationLabel = computed(() => {
  if (!props.previewResult) return '-'
  return props.previewResult.result.valid ? 'valid target' : 'invalid target'
})

function updateRollbackReason(value: string | number) {
  emit('update:rollbackReason', String(value))
}
</script>

<style lang="scss" scoped>
.snapshot-rollback-workbench {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.rollback-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(180px, 0.42fr);
  gap: var(--ms-space-3);
  align-items: stretch;

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: var(--ms-space-1);
    padding: var(--ms-space-3);
    border: 1px solid rgba(37, 99, 235, 0.18);
    border-radius: var(--ms-radius-lg);
    background: var(--ms-teal-50);
  }

  span {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  small {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
    font-weight: var(--ms-font-bold);
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
}

.event-chip {
  background: var(--ms-control-bg) !important;

  &.active {
    border-color: rgba(22, 163, 74, 0.24);
    background: var(--ms-green-50) !important;

    span {
      color: var(--ms-green-600);
    }
  }
}

.workflow-alert {
  flex-shrink: 0;
}

.snapshot-empty {
  min-height: 320px;
}

.history-section,
.rollback-flow,
.preview-summary,
.diff-section,
.no-diff-panel {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.section-heading,
.flow-heading {
  min-width: 0;
  display: flex;
  justify-content: space-between;
  gap: var(--ms-space-3);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  small {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-md);
    line-height: 1.35;
  }

  small {
    max-width: 260px;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-align: right;
  }
}

.snapshot-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  max-height: 360px;
  overflow: auto;
}

.snapshot-row {
  min-width: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  color: var(--ms-text-primary);
  text-align: left;
  background: var(--ms-panel-bg);
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    box-shadow 0.18s ease;

  &:hover,
  &:focus-visible,
  &.active {
    border-color: rgba(37, 99, 235, 0.34);
    outline: none;
    background: var(--ms-teal-50);
    box-shadow: var(--ms-shadow-xs);
  }
}

.row-title {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--ms-space-2);
  align-items: center;

  code {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-teal-700);
    font-size: var(--ms-text-sm);
  }
}

.snapshot-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ms-space-2);

  span {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  b {
    margin-right: 6px;
    color: var(--ms-text-secondary);
    font-weight: var(--ms-font-bold);
  }
}

.rollback-flow {
  padding-top: var(--ms-space-2);
  border-top: 1px solid var(--ms-border-light);
}

.flow-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--ms-space-2);

  :deep(.el-button) {
    margin-left: 0;
  }
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ms-space-2);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--ms-space-1);
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    background: var(--ms-control-bg);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code,
  strong {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-amber-600);

    &.ok {
      color: var(--ms-green-600);
    }
  }
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  span {
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-secondary);
    background: var(--ms-green-50);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
}

.diff-heading {
  color: var(--ms-text-tertiary);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

.field-diff,
.rule-diff,
.no-diff-panel {
  min-width: 0;
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);
}

.field-diff {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  header {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  code,
  span {
    min-width: 0;
    overflow: hidden;
    line-height: 1.45;
    text-overflow: ellipsis;
  }

  code {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.value-pair {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--ms-space-2);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  b {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-xs);
    text-transform: uppercase;
  }

  pre {
    min-height: 38px;
    max-height: 128px;
    overflow: auto;
    margin: 0;
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-secondary);
    background: var(--ms-control-bg);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-xs);
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
  }
}

.rule-diff {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  > header {
    min-width: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: var(--ms-space-2);
    align-items: start;

    div {
      min-width: 0;
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    code,
    span {
      min-width: 0;
      overflow: hidden;
      line-height: 1.45;
      text-overflow: ellipsis;
    }

    code {
      color: var(--ms-text-primary);
      font-size: var(--ms-text-sm);
      font-weight: var(--ms-font-semibold);
    }

    span {
      color: var(--ms-text-tertiary);
      font-size: var(--ms-text-sm);
    }

    strong {
      padding: 4px 8px;
      border-radius: var(--ms-radius-pill);
      color: var(--ms-teal-700);
      background: var(--ms-teal-50);
      font-size: var(--ms-text-xs);
      white-space: nowrap;
    }
  }
}

.change-flags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);

  span {
    padding: 3px 7px;
    border-radius: var(--ms-radius-pill);
    color: var(--ms-amber-600);
    background: var(--ms-amber-50);
    font-size: var(--ms-text-xs);
    font-weight: var(--ms-font-bold);
  }
}

.rule-field-diffs {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.no-diff-panel {
  background: var(--ms-green-50);

  strong {
    color: var(--ms-green-600);
    font-size: var(--ms-text-sm);
  }

  span {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
}

@media (max-width: 720px) {
  .rollback-hero,
  .summary-grid,
  .value-pair {
    grid-template-columns: 1fr;
  }

  .snapshot-meta {
    grid-template-columns: 1fr;
  }

  .flow-heading {
    flex-direction: column;

    small {
      max-width: none;
      text-align: left;
      white-space: normal;
    }
  }
}
</style>
