<template>
  <Section :title="'Execution conditions: ' + actionTitle" :padding="false">
    <template #toolbar>
      <router-link :to="{ name: 'ActionDetails', params: { actionId: route.params.actionId } }" class="button neutral">
        Back to action details
      </router-link>
    </template>

    <div v-if="action" class="padding content">
      <p>
        These entries mirror the automatic triggers from your OliveTin configuration for this action.
        You can always run the action manually as well.
      </p>

      <h3 class="exec-type-heading">
        On demand
        <a class="doc-link" :href="execConditionDocs.onDemand" target="_blank" rel="noopener noreferrer">Documentation</a>
      </h3>
      <p>
        Manual execution from the web UI (dashboard or action details), or via the API (for example StartAction),
        is always available when your user is allowed to execute the action.
      </p>

      <template v-if="action.execOnStartup">
        <h3 class="exec-type-heading">
          <code>execOnStartup</code>
          <a class="doc-link" :href="execConditionDocs.startup" target="_blank" rel="noopener noreferrer">Documentation</a>
        </h3>
        <p>Runs once when OliveTin starts.</p>
      </template>

      <template v-if="nonEmptyList(action.execOnCron)">
        <h3 class="exec-type-heading">
          <code>execOnCron</code>
          <a class="doc-link" :href="execConditionDocs.cron" target="_blank" rel="noopener noreferrer">Documentation</a>
        </h3>
        <ul>
          <li v-for="(line, idx) in action.execOnCron" :key="'cron-' + idx"><code>{{ line }}</code></li>
        </ul>
      </template>

      <template v-if="nonEmptyList(action.execOnFileCreatedInDir)">
        <h3 class="exec-type-heading">
          <code>execOnFileCreatedInDir</code>
          <a class="doc-link" :href="execConditionDocs.fileCreated" target="_blank" rel="noopener noreferrer">Documentation</a>
        </h3>
        <ul>
          <li v-for="(dir, idx) in action.execOnFileCreatedInDir" :key="'created-' + idx"><code>{{ dir }}</code></li>
        </ul>
      </template>

      <template v-if="nonEmptyList(action.execOnFileChangedInDir)">
        <h3 class="exec-type-heading">
          <code>execOnFileChangedInDir</code>
          <a class="doc-link" :href="execConditionDocs.fileChanged" target="_blank" rel="noopener noreferrer">Documentation</a>
        </h3>
        <ul>
          <li v-for="(dir, idx) in action.execOnFileChangedInDir" :key="'changed-' + idx"><code>{{ dir }}</code></li>
        </ul>
      </template>

      <template v-if="action.execOnCalendarFile">
        <h3 class="exec-type-heading">
          <code>execOnCalendarFile</code>
          <a class="doc-link" :href="execConditionDocs.calendar" target="_blank" rel="noopener noreferrer">Documentation</a>
        </h3>
        <p><code>{{ action.execOnCalendarFile }}</code></p>
      </template>

      <template v-if="nonEmptyList(action.execOnWebhooks)">
        <h3 class="exec-type-heading">
          <code>execOnWebhook</code>
          <a class="doc-link" :href="execConditionDocs.webhook" target="_blank" rel="noopener noreferrer">Documentation</a>
        </h3>
        <ul class="webhook-list">
          <li v-for="(wh, idx) in action.execOnWebhooks" :key="'wh-' + idx">
            <span v-if="wh.template">template: <code>{{ wh.template }}</code></span>
            <span v-if="wh.matchPath"> · matchPath: <code>{{ wh.matchPath }}</code></span>
            <span v-if="!wh.template && !wh.matchPath">Webhook trigger (no template or match path in response)</span>
          </li>
        </ul>
      </template>

      <p v-if="!hasConfiguredTriggers" class="muted">
        This action has no automatic triggers in configuration besides on-demand execution.
      </p>
    </div>

    <div v-else-if="!loading" class="padding empty-state">
      <p>Could not load this action.</p>
      <router-link :to="{ name: 'Actions' }">Return to index</router-link>
    </div>
  </Section>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'

const route = useRoute()
const action = ref(null)
const actionTitle = ref('Action')
const loading = ref(true)

const execConditionDocs = {
  onDemand: 'https://docs.olivetin.app/action_execution/ondemand.html',
  startup: 'https://docs.olivetin.app/action_execution/onstartup.html',
  cron: 'https://docs.olivetin.app/action_execution/oncron.html',
  fileCreated: 'https://docs.olivetin.app/action_execution/onfilecreated.html',
  fileChanged: 'https://docs.olivetin.app/action_execution/onfilechanged.html',
  calendar: 'https://docs.olivetin.app/action_execution/oncalendar.html',
  webhook: 'https://docs.olivetin.app/action_execution/onwebhook.html',
}

function nonEmptyList(list) {
  return Array.isArray(list) && list.length > 0
}

const hasConfiguredTriggers = computed(() => {
  const a = action.value
  if (!a) {
    return false
  }
  if (a.execOnStartup) {
    return true
  }
  if (nonEmptyList(a.execOnCron) || nonEmptyList(a.execOnFileCreatedInDir) || nonEmptyList(a.execOnFileChangedInDir)) {
    return true
  }
  if (a.execOnCalendarFile) {
    return true
  }
  if (nonEmptyList(a.execOnWebhooks)) {
    return true
  }
  return false
})

async function fetchAction() {
  loading.value = true
  try {
    const actionId = route.params.actionId
    const response = await window.client.getActionBinding({ bindingId: actionId })
    action.value = response.action
    actionTitle.value = response.action?.title || 'Action'
  } catch (err) {
    console.error('Failed to fetch action:', err)
    window.showBigError('fetch-action-exec-conditions', 'getting action', err, false)
    action.value = null
  } finally {
    loading.value = false
  }
}

onMounted(fetchAction)

watch(
  () => route.params.actionId,
  () => {
    action.value = null
    actionTitle.value = 'Action'
    fetchAction()
  }
)
</script>

<style scoped>
.content h3 {
  margin-top: 1.25rem;
  margin-bottom: 0.35rem;
  font-size: 1rem;
}

.exec-type-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.35rem 0.75rem;
}

.exec-type-heading .doc-link {
  font-size: 0.85rem;
  font-weight: normal;
}

.content p,
.content ul {
  margin: 0.35rem 0 0;
}

.webhook-list li {
  margin-bottom: 0.35rem;
}

.muted {
  color: var(--text-secondary);
  margin-top: 1.5rem;
}

.empty-state {
  text-align: center;
  color: var(--text-secondary);
}

.padding {
  padding: 1rem;
}
</style>
