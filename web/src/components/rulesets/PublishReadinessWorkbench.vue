<template>
  <section class="publish-readiness-workbench">
    <el-empty v-if="!ruleSet" description="Select a draft ruleset" />

    <template v-else>
      <header class="readiness-hero" :class="`is-${view.validationState}`">
        <div>
          <span>publish readiness</span>
          <strong>{{ view.validationLabel }} · {{ view.draftStateLabel }}</strong>
          <small>{{ view.validationDetail }} / {{ view.draftStateDetail }}</small>
        </div>
        <el-tag :type="view.canPublish ? 'success' : 'warning'" effect="plain">
          {{ view.canPublish ? 'ready' : 'blocked' }}
        </el-tag>
      </header>

      <div class="readiness-metrics">
        <div
          v-for="metric in view.metrics"
          :key="metric.label"
          class="metric-card"
          :class="`is-${metric.tone}`"
        >
          <span>{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
        </div>
      </div>

      <section class="readiness-panel">
        <header>
          <div>
            <span>validation</span>
            <strong>{{ view.validationLabel }}</strong>
          </div>
          <el-button size="small" :icon="Check" :loading="validating" @click="$emit('validate')">
            Run validation
          </el-button>
        </header>

        <el-alert
          v-if="validateError"
          type="error"
          show-icon
          :closable="false"
          :title="validateError"
        />

        <div v-if="view.issues.length" class="issue-list">
          <article v-for="issue in view.issues" :key="`${issue.path}-${issue.message}`" class="issue-row">
            <span>{{ issue.section }}</span>
            <div>
              <strong :title="issue.scope">{{ issue.scope }}</strong>
              <code :title="issue.path">{{ issue.path }}</code>
              <small :title="issue.message">{{ issue.message }}</small>
            </div>
          </article>
        </div>

        <div v-else class="empty-state" :class="{ ok: view.validationState === 'valid' }">
          <strong>{{ view.validationState === 'valid' ? 'No validation issues' : 'Validation not run' }}</strong>
          <small>{{ view.validationDetail }}</small>
        </div>

        <div v-if="view.warnings.length" class="warning-list">
          <header>
            <span>runtime warnings</span>
            <strong>{{ view.warnings.length }}</strong>
          </header>
          <article
            v-for="warning in view.warnings"
            :key="`${warning.path}-${warning.message}`"
            class="issue-row is-warning"
          >
            <span>{{ warning.section }}</span>
            <div>
              <strong :title="warning.scope">{{ warning.scope }}</strong>
              <code :title="warning.path">{{ warning.path }}</code>
              <small :title="warning.message">{{ warning.message }}</small>
            </div>
          </article>
        </div>
      </section>

      <section class="readiness-panel">
        <header>
          <div>
            <span>publication</span>
            <strong>{{ view.draftStateLabel }}</strong>
          </div>
        </header>

        <div class="publish-facts">
          <div>
            <span>current published</span>
            <code :title="view.publishedSnapshotLabel">{{ view.publishedSnapshotLabel }}</code>
          </div>
          <div>
            <span>last publish result</span>
            <code :title="view.publishResultLabel">{{ view.publishResultLabel }}</code>
          </div>
        </div>

        <el-alert
          v-if="publishError"
          type="error"
          show-icon
          :closable="false"
          :title="publishError"
        />

        <div v-if="publishResult" class="publish-success">
          <span class="status-dot is-on" />
          <div>
            <strong>Published {{ publishResult.snapshot.snapshot_id }}</strong>
            <small>
              version v{{ publishResult.ruleset.version || publishResult.snapshot.ruleset.version || 0 }}
              · {{ publishResult.snapshot.published_at }}
            </small>
          </div>
        </div>

        <div class="publish-actions">
          <el-button
            type="success"
            :icon="Upload"
            :loading="publishing"
            :disabled="!view.canPublish"
            @click="$emit('publish')"
          >
            Publish draft
          </el-button>
          <small v-if="!view.canPublish">Validation must pass first.</small>
          <small v-else-if="view.warnings.length">Publish is allowed, but runtime warnings should be reviewed.</small>
          <small v-else>A publish reason is required before release.</small>
        </div>
      </section>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Check, Upload } from '@element-plus/icons-vue'
import type {
  PublishedRuleSetSnapshot,
  PublishRuleSetResponse,
  RuleSet,
  ValidateRuleSetResponse,
} from '@/types'
import { buildPublishReadinessView } from '@/utils/publishReadiness'

const props = defineProps<{
  ruleSet: RuleSet | null
  currentPublished: PublishedRuleSetSnapshot | null
  validation: ValidateRuleSetResponse | null
  publishResult: PublishRuleSetResponse | null
  validateError: string
  publishError: string
  validating: boolean
  publishing: boolean
}>()

defineEmits<{
  validate: []
  publish: []
}>()

const view = computed(() => {
  return buildPublishReadinessView(
    props.ruleSet,
    props.currentPublished,
    props.validation,
    props.publishResult
  )
})
</script>

<style lang="scss" scoped>
.publish-readiness-workbench {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.readiness-hero {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid rgba(22, 163, 74, 0.22);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-green-50);

  &.is-invalid {
    border-color: rgba(248, 113, 113, 0.22);
    background: var(--ms-red-50);
  }

  &.is-not_run {
    border-color: var(--ms-border-light);
    background: var(--ms-control-bg);
  }

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  span:not(.el-tag__content) {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.readiness-metrics,
.publish-facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ms-space-2);
}

.metric-card,
.publish-facts > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
  padding: var(--ms-space-2);
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  code {
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    font-family: var(--ms-font-mono);
  }
}

.metric-card {
  &.is-ok {
    background: var(--ms-green-50);
  }

  &.is-warn {
    background: var(--ms-amber-50);
  }

  &.is-blocked {
    background: var(--ms-red-50);
  }
}

.readiness-panel {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);

  > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-3);
  }

  header div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  header span {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  header strong {
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
  }
}

.issue-list,
.warning-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.warning-list {
  padding-top: var(--ms-space-2);
  border-top: 1px solid var(--ms-border-light);

  > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-2);

    span {
      color: var(--ms-amber-600);
      font-size: var(--ms-text-sm);
      font-weight: var(--ms-font-bold);
      text-transform: uppercase;
    }

    strong {
      color: var(--ms-amber-600);
      font-size: var(--ms-text-sm);
    }
  }
}

.issue-row {
  min-width: 0;
  display: grid;
  grid-template-columns: 80px minmax(0, 1fr);
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border-radius: var(--ms-radius-md);
  background: var(--ms-red-50);

  > span {
    color: var(--ms-red-600);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  strong,
  code,
  small {
    min-width: 0;
    overflow: hidden;
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
  }

  code {
    color: var(--ms-text-secondary);
  }

  small {
    color: var(--ms-text-tertiary);
  }
}

.issue-row.is-warning {
  background: var(--ms-amber-50);

  > span {
    color: var(--ms-amber-600);
  }
}

.empty-state,
.publish-success {
  display: flex;
  align-items: flex-start;
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);

  &.ok,
  & {
    &.ok {
      background: var(--ms-green-50);
    }
  }

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  strong,
  small {
    overflow: hidden;
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    color: var(--ms-text-tertiary);
  }
}

.publish-success {
  background: var(--ms-green-50);
}

.publish-actions {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  :deep(.el-button) {
    width: 100%;
    margin-left: 0;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
}

@media (max-width: 620px) {
  .readiness-hero,
  .readiness-panel > header {
    flex-direction: column;
  }

  .readiness-hero {
    :deep(.el-tag) {
      align-self: flex-start;
    }

    strong,
    small {
      white-space: normal;
    }
  }

  .readiness-metrics,
  .publish-facts,
  .issue-row {
    grid-template-columns: 1fr;
  }
}
</style>
