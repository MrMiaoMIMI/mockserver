<template>
  <section class="rule-editor-workbench">
    <el-empty v-if="!ruleSet" description="Select a draft ruleset" />
    <el-empty v-else-if="mode === 'edit' && !rule" description="Select a rule to edit" />

    <div v-else class="editor-shell">
      <header class="editor-context">
        <div>
          <span>{{ mode === 'create' ? 'create rule' : 'edit rule' }}</span>
          <strong :title="form.id">{{ editorTitle }}</strong>
          <small v-if="mode === 'create' && seedLabel">{{ seedLabel }}</small>
        </div>
        <el-tag :type="mode === 'create' ? 'success' : 'info'" effect="plain">
          {{ mode }}
        </el-tag>
      </header>

      <div class="editor-status-strip">
        <el-popover
          placement="bottom-start"
          trigger="hover"
          :width="380"
          popper-class="rule-status-popover"
        >
          <template #reference>
            <button type="button" :class="['status-chip', `is-${overallTone}`]">
              <span>status</span>
              <strong>{{ overallStatusLabel }}</strong>
              <el-icon><InfoFilled /></el-icon>
            </button>
          </template>
          <div class="status-popover">
            <button
              v-for="item in authoringView.sectionReadiness"
              :key="item.section"
              type="button"
              :class="[`is-${item.tone}`, { active: activeSection === item.section }]"
              @click="activeSection = item.section"
            >
              <span>{{ item.label }}</span>
              <strong>{{ item.status }}</strong>
              <small>{{ item.detail }}</small>
            </button>
          </div>
        </el-popover>

        <el-popover
          placement="bottom-start"
          trigger="hover"
          :width="460"
          popper-class="rule-status-popover"
        >
          <template #reference>
            <button type="button" class="summary-chip">
              <span>summary</span>
              <strong>{{ compactSummaryLabel }}</strong>
              <el-icon><InfoFilled /></el-icon>
            </button>
          </template>
          <div class="summary-popover">
            <div class="summary-section">
              <span>condition</span>
              <ConditionTreePreview
                v-if="authoringView.rule"
                :condition="authoringView.rule.when"
                aria-label="Editor condition tree"
              />
              <code v-else>{{ authoringView.conditionSummary }}</code>
            </div>
            <div class="summary-section">
              <span>action</span>
              <code>{{ authoringView.actionSummary }}</code>
            </div>
          </div>
        </el-popover>
      </div>

      <div class="section-switch" role="tablist" aria-label="Rule editor sections">
        <button
          v-for="section in sections"
          :key="section.name"
          type="button"
          :class="{ active: activeSection === section.name, invalid: sectionHasError(section.name) }"
          @click="activeSection = section.name"
        >
          <span>{{ section.label }}</span>
        </button>
      </div>

      <div v-if="displayErrors.length" class="validation-summary">
        <strong>{{ displayErrors.length }} fields need attention</strong>
        <button
          v-for="error in displayErrors"
          :key="`${error.field}-${error.message}`"
          type="button"
          @click="activeSection = error.section"
        >
          {{ sectionLabel(error.section) }} · {{ error.message }}
        </button>
      </div>

      <div class="editor-body">
        <section v-show="activeSection === 'identity'" class="editor-section">
          <div class="section-heading">
            <span>identity</span>
            <strong>Rule identity and ordering</strong>
          </div>
          <el-form label-position="top" class="editor-form">
            <template v-if="mode === 'create'">
              <el-form-item label="Name" :error="fieldError('name')">
                <el-input v-model="form.name" placeholder="Debug API response" clearable />
              </el-form-item>
              <div class="generated-id-card">
                <span>Generated Rule ID</span>
                <code :title="form.id">{{ form.id }}</code>
                <small>Created from the name and checked against this ruleset.</small>
              </div>
            </template>
            <template v-else>
              <el-form-item label="Rule ID" :error="fieldError('id')">
                <el-input v-model="form.id" disabled placeholder="rule-001" />
              </el-form-item>
              <el-form-item label="Name" :error="fieldError('name')">
                <el-input v-model="form.name" placeholder="Debug API response" clearable />
              </el-form-item>
            </template>
            <div class="form-grid identity-grid">
              <el-form-item label="Priority" :error="fieldError('priority')">
                <el-input-number v-model="form.priority" :min="0" controls-position="right" />
              </el-form-item>
              <el-form-item label="Enabled">
                <el-switch v-model="form.enabled" />
              </el-form-item>
            </div>
          </el-form>
        </section>

        <section v-show="activeSection === 'condition'" class="editor-section">
          <div class="section-heading">
            <span>condition</span>
            <strong>Match request context</strong>
          </div>
          <el-radio-group v-model="form.conditionMode" class="mode-switch" size="small">
            <el-radio-button value="tree">Tree</el-radio-button>
            <el-radio-button value="expr">CEL</el-radio-button>
            <el-radio-button value="raw">Raw JSON</el-radio-button>
          </el-radio-group>

          <ConditionTreeEditor
            v-if="form.conditionMode === 'tree'"
            v-model="form.conditionTree"
            :protocol-spec="protocolSpec"
          />

          <el-form v-else-if="form.conditionMode === 'expr'" label-position="top" class="editor-form">
            <el-form-item label="CEL Expression" :error="fieldError('conditionExpr')">
              <el-input
                v-model="form.conditionExpr"
                type="textarea"
                :rows="5"
                placeholder="request.path == '/api/v1/debug'"
              />
            </el-form-item>
          </el-form>

          <div v-else class="raw-block">
            <JsonEditor v-model="form.conditionJson" :min-height="280" title="Condition JSON" />
            <p v-if="fieldError('conditionJson')" class="field-error">
              {{ fieldError('conditionJson') }}
            </p>
          </div>

          <p v-if="fieldError('conditionTree')" class="field-error">
            {{ fieldError('conditionTree') }}
          </p>
        </section>

        <section v-show="activeSection === 'action'" class="editor-section">
          <div class="section-heading">
            <span>action</span>
            <strong>{{ actionTitle }}</strong>
          </div>
          <p class="section-note">{{ authoringView.actionDescription }}</p>

          <el-form label-position="top" class="editor-form">
            <el-form-item label="Action Type" :error="fieldError('actionType')">
              <div class="action-type-grid" role="radiogroup" aria-label="Action Type">
                <button
                  v-for="profile in actionProfiles"
                  :key="profile.type"
                  type="button"
                  :class="{ active: form.actionType === profile.type }"
                  @click="form.actionType = profile.type"
                >
                  <strong>{{ profile.label }}</strong>
                  <small>{{ profile.detail }}</small>
                </button>
              </div>
            </el-form-item>

            <template v-if="requiresStatus">
              <div class="form-grid action-grid">
                <el-form-item label="Status" :error="fieldError('status')">
                  <el-input-number
                    v-model="form.status"
                    :min="100"
                    :max="599"
                    controls-position="right"
                  />
                </el-form-item>
              </div>
              <el-form-item label="Headers JSON" :error="fieldError('headersJson')">
                <el-input v-model="form.headersJson" type="textarea" :rows="4" />
              </el-form-item>
            </template>

            <el-form-item
              v-if="form.actionType === 'static'"
              label="Response Payload JSON"
              :error="fieldError('bodyJson')"
            >
              <el-input v-model="form.bodyJson" type="textarea" :rows="7" />
            </el-form-item>

            <el-form-item
              v-if="form.actionType === 'template'"
              label="Response Payload Template"
              :error="fieldError('bodyTemplate')"
            >
              <el-input v-model="form.bodyTemplate" type="textarea" :rows="7" />
            </el-form-item>

            <el-form-item
              v-if="form.actionType === 'cel'"
              label="Response Payload Expression"
              :error="fieldError('bodyExpression')"
            >
              <el-input v-model="form.bodyExpression" type="textarea" :rows="7" />
            </el-form-item>
          </el-form>

          <div v-if="form.actionType === 'sequence'" class="sequence-editor">
            <div class="sequence-toolbar">
              <el-select v-model="form.sequenceStrategy" size="small">
                <el-option label="loop" value="loop" />
                <el-option label="last" value="last" />
              </el-select>
              <el-button size="small" :icon="Plus" @click="appendSequenceStep">Add step</el-button>
            </div>
            <p v-if="fieldError('sequenceSteps')" class="field-error">
              {{ fieldError('sequenceSteps') }}
            </p>

            <article v-for="(step, index) in form.sequenceSteps" :key="step.key" class="sequence-step">
              <header>
                <strong>Step {{ index + 1 }}</strong>
                <div class="step-actions">
                  <button
                    class="icon-button"
                    type="button"
                    title="Move step up"
                    :disabled="index === 0"
                    @click="moveSequenceStep(index, -1)"
                  >
                    <el-icon><ArrowUp /></el-icon>
                  </button>
                  <button
                    class="icon-button"
                    type="button"
                    title="Move step down"
                    :disabled="index === form.sequenceSteps.length - 1"
                    @click="moveSequenceStep(index, 1)"
                  >
                    <el-icon><ArrowDown /></el-icon>
                  </button>
                  <button
                    class="icon-button danger"
                    type="button"
                    title="Delete step"
                    :disabled="form.sequenceSteps.length <= 1"
                    @click="removeSequenceStep(index)"
                  >
                    <el-icon><Delete /></el-icon>
                  </button>
                </div>
              </header>
              <el-form label-position="top" class="editor-form">
                <el-form-item label="Response Payload JSON" :error="fieldError(`sequenceSteps.${index}.bodyJson`)">
                  <el-input v-model="step.bodyJson" type="textarea" :rows="5" />
                </el-form-item>
              </el-form>
            </article>
          </div>

          <el-form
            v-if="form.actionType === 'webhook'"
            label-position="top"
            class="editor-form"
          >
            <el-form-item label="Webhook URL" :error="fieldError('webhookUrl')">
              <el-input v-model="form.webhookUrl" placeholder="http://127.0.0.1:9000/mock" />
            </el-form-item>
            <div class="form-grid webhook-grid">
              <el-form-item label="Method">
                <el-select v-model="form.webhookMethod">
                  <el-option label="POST" value="POST" />
                  <el-option label="GET" value="GET" />
                  <el-option label="PUT" value="PUT" />
                </el-select>
              </el-form-item>
              <el-form-item label="Timeout MS" :error="fieldError('webhookTimeoutMS')">
                <el-input-number v-model="form.webhookTimeoutMS" :min="0" :max="30000" />
              </el-form-item>
            </div>
            <el-form-item label="Webhook Headers JSON" :error="fieldError('webhookHeadersJson')">
              <el-input v-model="form.webhookHeadersJson" type="textarea" :rows="5" />
            </el-form-item>
          </el-form>
        </section>

        <section v-show="activeSection === 'advanced'" class="editor-section">
          <div class="section-heading">
            <span>advanced</span>
            <strong>Raw rule payload</strong>
          </div>
          <div class="raw-actions">
            <el-button size="small" :icon="Refresh" @click="refreshRawJson">Refresh JSON from form</el-button>
            <el-button size="small" :icon="Upload" @click="loadRawJson">Load form from JSON</el-button>
          </div>
          <JsonEditor v-model="form.rawRuleJson" :min-height="420" title="Rule JSON" />
          <p v-if="fieldError('rawRuleJson')" class="field-error">
            {{ fieldError('rawRuleJson') }}
          </p>
        </section>

        <section v-show="activeSection === 'preview'" class="editor-section">
          <div class="section-heading">
            <span>preview</span>
            <strong>Simulate unsaved draft</strong>
          </div>
          <p class="section-note">
            Preview uses a temporary draft override built from this editor form, so changes can be
            tested before saving.
          </p>
          <div class="preview-toolbar">
            <el-button size="small" :icon="Refresh" @click="refreshPreviewEvent">
              Use suggested event
            </el-button>
            <el-button
              type="primary"
              size="small"
              :loading="previewLoading"
              :disabled="!authoringView.canSimulate"
              @click="runAuthoringPreview"
            >
              Simulate current rule
            </el-button>
          </div>
          <JsonEditor v-model="previewEventModel" :min-height="240" title="Simulation Event" />
          <p v-if="previewError" class="field-error">{{ previewError }}</p>
          <div class="preview-result">
            <ResultInspector :raw-json="previewResultJson" />
          </div>
        </section>
      </div>

      <footer class="editor-footer">
        <el-button @click="resetForm">Reset</el-button>
        <el-button @click="$emit('cancel')">Cancel</el-button>
        <el-button :loading="previewLoading" @click="openPreviewAndRun">Preview</el-button>
        <el-button type="primary" :loading="saving" @click="submitRule">Save rule</el-button>
      </footer>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowDown, ArrowUp, Delete, InfoFilled, Plus, Refresh, Upload } from '@element-plus/icons-vue'
import JsonEditor from '@/components/common/JsonEditor.vue'
import ConditionTreePreview from '@/components/rulesets/ConditionTreePreview.vue'
import ConditionTreeEditor from '@/components/rulesets/ConditionTreeEditor.vue'
import ResultInspector from '@/components/rulesets/ResultInspector.vue'
import { useMockserverStore } from '@/store'
import type { MockEvent, Rule, RuleSet } from '@/types'
import {
  ACTION_AUTHORING_PROFILES,
  buildRuleAuthoringView,
  type SectionReadiness,
} from '@/utils/ruleAuthoring'
import {
  defaultRuleForm,
  formToRule,
  generateRuleIdFromName,
  newSequenceStepForm,
  ruleToForm,
  type RuleEditorMode,
  type RuleFormError,
  type RuleFormState,
} from '@/utils/ruleFormAdapter'

type EditorSection = RuleFormError['section'] | 'preview'

const props = defineProps<{
  ruleSet: RuleSet | null
  rule: Rule | null
  mode: RuleEditorMode
  seedRule?: Rule | null
  seedLabel?: string
}>()

const emit = defineEmits<{
  updated: [ruleSet: RuleSet, ruleId: string]
  cancel: []
}>()

const store = useMockserverStore()
const form = reactive<RuleFormState>(defaultRuleForm())
const activeSection = ref<EditorSection>('identity')
const formErrors = ref<RuleFormError[]>([])
const saving = ref(false)
const previewLoading = ref(false)
const previewEventJson = ref('')
const previewEventDirty = ref(false)
const previewResultJson = ref('')
const previewError = ref('')

const actionProfiles = ACTION_AUTHORING_PROFILES
const sections: Array<{ name: EditorSection; label: string }> = [
  { name: 'identity', label: 'Identity' },
  { name: 'condition', label: 'Condition' },
  { name: 'action', label: 'Action' },
  { name: 'advanced', label: 'Raw' },
  { name: 'preview', label: 'Preview' },
]

const requiresStatus = computed(() => {
  return false
})
const actionTitle = computed(() => {
  if (form.actionType === 'sequence') return 'Ordered response sequence'
  if (form.actionType === 'webhook') return 'Forward to external webhook'
  if (form.actionType === 'template') return 'Template response payload'
  if (form.actionType === 'cel') return 'CEL-generated response'
  return 'Static mock response'
})
const existingRuleIds = computed(() => props.ruleSet?.rules.map((rule) => rule.id) || [])
const buildOptions = computed(() => ({
  existingRuleIds: existingRuleIds.value,
  lockedRuleId: props.mode === 'edit' ? props.rule?.id : undefined,
}))
const authoringView = computed(() => buildRuleAuthoringView(form, props.ruleSet, buildOptions.value))
const protocolSpec = computed(() => {
  return props.ruleSet ? store.protocolMap.get(props.ruleSet.protocol) : undefined
})
const liveErrors = computed(() => authoringView.value.errors)
const displayErrors = computed(() => (formErrors.value.length ? formErrors.value : liveErrors.value))
const overallTone = computed(() => {
  if (displayErrors.value.length) return 'danger'
  if (authoringView.value.sectionReadiness.some((item) => item.tone === 'danger')) return 'danger'
  if (authoringView.value.sectionReadiness.some((item) => item.tone === 'warn')) return 'warn'
  return 'ok'
})
const overallStatusLabel = computed(() => {
  if (displayErrors.value.length) {
    return `${displayErrors.value.length} issue${displayErrors.value.length > 1 ? 's' : ''}`
  }
  return 'ready'
})
const compactSummaryLabel = computed(() => {
  return `${form.conditionMode.toUpperCase()} / ${form.actionType.replace(/_response$/, '')}`
})
const editorTitle = computed(() => {
  if (props.mode === 'create') return form.name.trim() || 'New rule'
  return form.name.trim() || form.id || 'Rule'
})
const previewEventModel = computed({
  get: () => previewEventJson.value,
  set: (value: string) => {
    previewEventJson.value = value
    previewEventDirty.value = value !== authoringView.value.eventJson
    previewError.value = ''
  },
})

watch(
  () => [props.mode, props.rule?.id, props.ruleSet?.id, props.seedRule?.id],
  () => resetForm(),
  { immediate: true }
)

watch(
  () => authoringView.value.eventJson,
  (eventJson) => {
    if (!eventJson) return
    if (!previewEventDirty.value || !previewEventJson.value) {
      previewEventJson.value = eventJson
      previewEventDirty.value = false
    }
  },
  { immediate: true }
)

watch(
  form,
  () => {
    if (formErrors.value.length) {
      formErrors.value = []
    }
  },
  { deep: true }
)

watch(
  () => [form.name, props.mode, props.ruleSet?.rules.map((rule) => rule.id).join('\u0000') || ''],
  () => {
    if (props.mode !== 'create' || activeSection.value === 'advanced') return
    form.id = generateRuleIdFromName(form.name, props.ruleSet)
  }
)

function resetForm() {
  const nextForm = props.mode === 'edit' && props.rule
    ? ruleToForm(props.rule)
    : props.mode === 'create' && props.seedRule
      ? ruleToForm(props.seedRule)
      : defaultRuleForm(props.ruleSet)
  Object.assign(form, nextForm)
  const nextView = buildRuleAuthoringView(nextForm, props.ruleSet, {
    existingRuleIds: existingRuleIds.value,
    lockedRuleId: props.mode === 'edit' ? props.rule?.id : undefined,
  })
  formErrors.value = []
  previewEventJson.value = nextView.eventJson
  previewEventDirty.value = false
  previewResultJson.value = ''
  previewError.value = ''
  activeSection.value = 'identity'
}

async function submitRule() {
  if (!props.ruleSet) return
  const source = activeSection.value === 'advanced' ? 'raw' : 'form'
  const result = formToRule(form, {
    source,
    ...buildOptions.value,
  })
  if (handleBuildErrors(result.errors) || !result.rule) return

  saving.value = true
  try {
    const updated =
      props.mode === 'edit' && props.rule
        ? await store.updateRule(props.ruleSet.id, props.rule.id, result.rule)
        : await store.addRule(props.ruleSet.id, result.rule)
    emit('updated', updated, result.rule.id)
  } finally {
    saving.value = false
  }
}

function refreshRawJson() {
  const result = formToRule(form, {
    source: 'form',
    ...buildOptions.value,
  })
  if (handleBuildErrors(result.errors) || !result.rule) return
  form.rawRuleJson = JSON.stringify(result.rule, null, 2)
  formErrors.value = []
  ElMessage.success('Raw JSON refreshed from form')
}

function loadRawJson() {
  const result = formToRule(form, {
    source: 'raw',
    ...buildOptions.value,
  })
  if (handleBuildErrors(result.errors) || !result.rule) return
  Object.assign(form, ruleToForm(result.rule))
  formErrors.value = []
  activeSection.value = 'identity'
  ElMessage.success('Form loaded from Raw JSON')
}

function appendSequenceStep() {
  form.sequenceSteps.push(newSequenceStepForm(form.sequenceSteps.length + 1))
}

function removeSequenceStep(index: number) {
  if (form.sequenceSteps.length <= 1) return
  form.sequenceSteps.splice(index, 1)
}

function moveSequenceStep(index: number, direction: -1 | 1) {
  const target = index + direction
  if (target < 0 || target >= form.sequenceSteps.length) return
  const [step] = form.sequenceSteps.splice(index, 1)
  form.sequenceSteps.splice(target, 0, step)
}

function refreshPreviewEvent() {
  if (!authoringView.value.eventJson) {
    const firstError = authoringView.value.errors[0]
    if (firstError) {
      activeSection.value = firstError.section
      ElMessage.error(firstError.message)
    }
    return
  }
  previewEventJson.value = authoringView.value.eventJson
  previewEventDirty.value = false
  previewError.value = ''
}

async function openPreviewAndRun() {
  activeSection.value = 'preview'
  await runAuthoringPreview()
}

async function runAuthoringPreview() {
  if (!props.ruleSet) return
  const view = authoringView.value
  if (handleBuildErrors(view.errors) || !view.draftOverride) return

  previewLoading.value = true
  previewError.value = ''
  try {
    const event = parsePreviewEvent()
    const result = await store.simulateDraft(props.ruleSet.id, {
      event,
      draft_override: view.draftOverride,
      explain_only: false,
      explain_summary: false,
      explain_max_depth: 3,
    })
    previewResultJson.value = JSON.stringify({ result }, null, 2)
  } catch (error) {
    previewError.value = toErrorMessage(error)
  } finally {
    previewLoading.value = false
  }
}

function parsePreviewEvent(): MockEvent {
  try {
    return JSON.parse(previewEventJson.value) as MockEvent
  } catch (error) {
    throw new Error(`Simulation Event is not valid JSON: ${toErrorMessage(error)}`)
  }
}

function handleBuildErrors(errors: RuleFormError[]) {
  formErrors.value = errors
  if (!errors.length) return false
  activeSection.value = errors[0].section
  ElMessage.error(errors[0].message)
  return true
}

function fieldError(field: string) {
  return displayErrors.value.find((error) => error.field === field)?.message || ''
}

function sectionHasError(section: EditorSection) {
  return displayErrors.value.some((error) => error.section === section)
}

function sectionLabel(section: SectionReadiness['section']) {
  return sections.find((item) => item.name === section)?.label || section
}

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}
</script>

<style lang="scss" scoped>
.rule-editor-workbench {
  height: 100%;
  min-height: 0;
}

.editor-shell {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.editor-context {
  flex-shrink: 0;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: 10px var(--ms-space-3);
  border: 1px solid rgba(37, 99, 235, 0.22);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-teal-50);

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
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-md);
    line-height: 1.25;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.editor-status-strip {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: minmax(0, 0.62fr) minmax(0, 1fr);
  gap: var(--ms-space-2);
}

.status-chip,
.summary-chip {
  min-width: 0;
  width: 100%;
  height: 34px;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--ms-space-2);
  padding: 0 10px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  color: var(--ms-text-secondary);
  background: var(--ms-control-bg);
  font: inherit;
  cursor: pointer;

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  .el-icon {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    min-width: 0;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-align: right;
  }

  .el-icon {
    color: var(--ms-text-tertiary);
    font-size: 14px;
  }
}

.status-chip {
  &.is-ok {
    border-color: rgba(22, 163, 74, 0.24);
    background: var(--ms-green-50);
  }

  &.is-warn {
    border-color: rgba(245, 158, 11, 0.3);
    background: var(--ms-amber-50);

    strong {
      color: var(--ms-amber-600);
    }
  }


  &.is-danger {
    border-color: rgba(248, 113, 113, 0.28);
    background: var(--ms-red-50);

    strong {
      color: var(--ms-red-600);
    }
  }
}

:global(.rule-status-popover .status-popover),
:global(.rule-status-popover .summary-popover) {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

:global(.rule-status-popover .status-popover button) {
  min-width: 0;
  display: grid;
  grid-template-columns: 84px minmax(80px, auto) minmax(0, 1fr);
  align-items: center;
  gap: var(--ms-space-2);
  padding: 8px 10px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  color: var(--ms-text-secondary);
  background: var(--ms-panel-bg);
  text-align: left;
  cursor: pointer;
}

:global(.rule-status-popover .status-popover button.active),
:global(.rule-status-popover .status-popover button:hover) {
  border-color: rgba(37, 99, 235, 0.28);
  background: var(--ms-teal-50);
}

:global(.rule-status-popover .status-popover button.is-danger) {
  border-color: rgba(248, 113, 113, 0.28);
  background: var(--ms-red-50);
}

:global(.rule-status-popover .status-popover span),
:global(.rule-status-popover .summary-popover span) {
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

:global(.rule-status-popover .status-popover strong) {
  color: var(--ms-text-primary);
  font-size: var(--ms-text-sm);
}

:global(.rule-status-popover .status-popover small),
:global(.rule-status-popover .summary-popover code) {
  min-width: 0;
  overflow: hidden;
  color: var(--ms-text-secondary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(.rule-status-popover .summary-popover .summary-section) {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
  padding: 8px 10px;
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);
}

.section-switch {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);

  button {
    height: 32px;
    border: none;
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-tertiary);
    background: transparent;
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    cursor: pointer;
    transition:
      color var(--ms-transition-fast),
      background var(--ms-transition-fast);

    &.active {
      color: var(--ms-text-inverse);
      background: var(--ms-teal-600);
      box-shadow: var(--ms-shadow-xs);
    }

    &:hover:not(.active) {
      color: var(--ms-teal-700);
      background: var(--ms-teal-50);
    }

    &.invalid {
      color: var(--ms-red-600);
    }
  }
}

.section-note {
  margin: 0;
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  line-height: 1.5;
}

.validation-summary {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border: 1px solid rgba(248, 113, 113, 0.22);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-red-50);

  strong {
    color: var(--ms-red-600);
    font-size: var(--ms-text-sm);
  }

  button {
    min-width: 0;
    border: none;
    color: var(--ms-red-600);
    background: transparent;
    text-align: left;
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    cursor: pointer;
  }
}

.editor-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding-right: 2px;
  padding-bottom: var(--ms-space-3);
}

.editor-section {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.section-heading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--ms-space-2);

  span {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    line-height: 1.3;
  }
}

.editor-form {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

.generated-id-card {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);

  span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
}

.form-grid {
  display: grid;
  gap: var(--ms-space-3);
}

.identity-grid,
.webhook-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.action-grid {
  grid-template-columns: 180px;
}

.action-type-grid {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: var(--ms-space-2);

  button {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-height: 78px;
    padding: var(--ms-space-2);
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-lg);
    color: var(--ms-text-secondary);
    background: var(--ms-panel-bg);
    text-align: left;
    cursor: pointer;
    transition:
      border-color var(--ms-transition-fast),
      background var(--ms-transition-fast),
      box-shadow var(--ms-transition-fast);

    &.active {
      border-color: rgba(37, 99, 235, 0.36);
      background: var(--ms-teal-50);
      box-shadow: var(--ms-shadow-xs);
    }
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    line-height: 1.2;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.35;
  }
}

.mode-switch,
.raw-actions,
.preview-toolbar,
.sequence-toolbar,
.step-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--ms-space-2);
}

.preview-toolbar {
  justify-content: space-between;
}

.preview-result {
  min-height: 220px;
  padding: var(--ms-space-2);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-bg-base);
}

.raw-block {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.sequence-editor {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.sequence-toolbar {
  justify-content: space-between;
}

.sequence-step {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-2);
  }

  strong {
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
  }
}

.field-error {
  margin: 0;
  color: var(--ms-red-600);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
}

.danger {
  color: var(--ms-red-600);
}

.editor-footer {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: 1fr 1fr 1.2fr 1.4fr;
  gap: var(--ms-space-2);
  padding-top: var(--ms-space-3);
  border-top: 1px solid var(--ms-border-light);
  background: var(--ms-panel-bg);

  :deep(.el-button) {
    margin-left: 0;
  }
}

@media (max-width: 620px) {
  .editor-status-strip,
  .section-switch,
  .action-type-grid,
  .identity-grid,
  .webhook-grid,
  .editor-footer {
    grid-template-columns: 1fr;
  }
}
</style>
