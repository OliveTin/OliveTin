<template>
  <Section :padding="false">
    <template #title>
      <span class="section-title-with-icon">
        Action Details:
        <ActionIconGlyph
          v-if="action"
          class="action-title-icon"
          :glyph="action.icon"
        />
        {{ actionTitle }}
      </span>
    </template>
    <template #toolbar>
      <div class="action-details-toolbar">
        <button
          v-for="dashboard in backToDashboards"
          :key="dashboard.path"
          :title="'Back to ' + dashboard.title"
          class="button neutral"
          @click="goToDashboard(dashboard.path)"
        >
          <HugeiconsIcon :icon="DashboardSquare01Icon" />
          {{ dashboard.title }}
        </button>
        <button
          v-if="action"
          title="Run this action"
          class="button neutral"
          @click="startAction"
        >
          <HugeiconsIcon :icon="WorkoutRunIcon" />
          Run
        </button>
        <router-link
          v-if="action"
          :to="{ name: 'ActionExecConditions', params: { actionId: route.params.actionId } }"
          class="button neutral"
          title="View configured automatic triggers and on-demand execution"
        >
          Execution conditions ({{ executionConditionCount }})
        </router-link>
      </div>
    </template>

    <div
      v-if="action"
      class="flex-row padding"
    >
      <div class="fg1">
        <dl>
          <dt>Timeout</dt>
          <dd>{{ action.timeout }} seconds</dd>

          <template v-if="actionGroups.length > 0">
            <dt>
              <router-link
                :to="{ name: 'LogsQueue' }"
                class="action-groups-link"
              >
                Action groups
              </router-link>
            </dt>
            <dd>
              <ul class="action-group-list">
                <li
                  v-for="group in actionGroups"
                  :key="group.name"
                  class="action-group-row"
                >
                  <router-link
                    :to="{ name: 'LogsQueue' }"
                    class="action-groups-link action-group-name"
                  >
                    {{ group.name }}
                  </router-link><template v-if="group.maxConcurrent > 0 && group.queueSize > 0">
                    -
                  </template><ActionGroupLimitsLabel
                    :max-concurrent="group.maxConcurrent"
                    :queue-size="group.queueSize"
                  />
                </li>
              </ul>
            </dd>
          </template>
        </dl>
        <p class="fg1">
          Execution history for this action. You can filter by execution tracking ID.
        </p>
      </div>
      <div style="align-self: start; text-align: right;">
        <div class="filter-container">
          <label class="input-with-icons">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1em"
              height="1em"
              viewBox="0 0 24 24"
            >
              <path
                fill="currentColor"
                d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14"
              />
            </svg>
            <input
              v-model="searchText"
              placeholder="Filter current page"
            >
            <button
              title="Clear search filter"
              :disabled="!searchText"
              @click="clearSearch"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="1em"
                height="1em"
                viewBox="0 0 24 24"
              >
                <path
                  fill="currentColor"
                  d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z"
                />
              </svg>
            </button>
          </label>
        </div>
      </div>
    </div>

    <div v-show="filteredLogs.length > 0">
      <table class="logs-table row-hover">
        <thead>
          <tr>
            <th>Timestamp</th>
            <th>Duration</th>
            <th>Execution ID</th>
            <th>Metadata</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="log in filteredLogs"
            :key="log.executionTrackingId"
            class="log-row"
            :title="log.actionTitle"
          >
            <td class="timestamp">
              {{ formatTimestamp(log.datetimeStarted) }}
            </td>
            <td class="duration">
              {{ formatExecutionDuration(log) }}
            </td>
            <td>
              <router-link :to="`/logs/${log.executionTrackingId}`">
                <LogActionTitle :justification="log.justification">
                  {{ log.executionTrackingId }}
                </LogActionTitle>
              </router-link>
            </td>
            <td class="tags">
              <span class="annotation">
                <span class="annotation-key">User:</span>
                <span class="annotation-val">{{ log.user }}</span>
              </span>
              <span
                v-if="log.tags && log.tags.length > 0"
                class="tag-list"
              >
                <span
                  v-for="tag in log.tags"
                  :key="tag"
                  class="tag"
                >{{ tag }}</span>
              </span>
            </td>
            <td class="exit-code">
              <ActionStatusDisplay
                :log-entry="log"
                :link-queued-status="true"
              />
            </td>
          </tr>
        </tbody>
      </table>

      <Pagination
        :page-size="pageSize"
        :total="totalCount"
        :current-page="currentPage"
        :page="currentPage"
        class="padding"
        item-title="execution logs"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      />
    </div>

    <div
      v-show="logs.length === 0 && !loading"
      class="empty-state"
    >
      <p>This action has no execution history.</p>
      <router-link to="/">
        Return to index
      </router-link>
    </div>
  </Section>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import Section from 'picocrank/vue/components/Section.vue'
import ActionIconGlyph from '../components/ActionIconGlyph.vue'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'
import ActionGroupLimitsLabel from '../components/ActionGroupLimitsLabel.vue'
import LogActionTitle from '../components/LogActionTitle.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { DashboardSquare01Icon, WorkoutRunIcon } from '@hugeicons/core-free-icons'
import { requestReconnectNow } from '../../../js/websocket.js'
import { needsArgumentForm } from '../utils/needsArgumentForm.js'
import { getExecutionLogEntry, updateLogEntryInList } from '../utils/executionLogEvents.js'
import { countExecutionConditions } from '../utils/executionConditionCount.js'

const route = useRoute()
const router = useRouter()
const logs = ref([])
const action = ref(null)
const backToDashboards = ref([])
const actionTitle = ref('Action Details')
const searchText = ref('')
const pageSize = ref(10)
const currentPage = ref(1)
const loading = ref(false)
const totalCount = ref(0)
const durationClock = ref(Date.now())
let durationTicker = null

const filteredLogs = computed(() => {
  if (!searchText.value) {
    return logs.value
  }
  const searchLower = searchText.value.toLowerCase()
  return logs.value.filter(log =>
    log.executionTrackingId.toLowerCase().includes(searchLower) ||
    log.actionTitle.toLowerCase().includes(searchLower) ||
    (log.justification || '').toLowerCase().includes(searchLower)
  )
})

const executionConditionCount = computed(() => countExecutionConditions(action.value))

const actionGroups = computed(() => action.value?.groups ?? [])

async function fetchActionLogs () {
  loading.value = true
  try {
    const actionId = route.params.actionId
    const startOffset = (currentPage.value - 1) * pageSize.value

    const args = {
      actionId,
      startOffset: BigInt(startOffset),
      pageSize: BigInt(Number(pageSize.value))
    }

    const response = await window.client.getActionLogs(args)

    logs.value = response.logs
    const serverPageSize = Number(response.pageSize)
    if (Number.isFinite(serverPageSize) && serverPageSize > 0) {
      pageSize.value = serverPageSize
    }
    totalCount.value = Number(response.totalCount) || 0
    syncDurationTicker()
  } catch (err) {
    console.error('Failed to fetch action logs:', err)
    window.showBigError('fetch-action-logs', 'getting action logs', err, false)
  } finally {
    loading.value = false
  }
}

async function fetchAction () {
  try {
    const actionId = route.params.actionId
    const args = {
      bindingId: actionId
    }
    const response = await window.client.getActionBinding(args)
    action.value = response.action
    backToDashboards.value = (response.backToDashboards || []).slice(0, 3)
    actionTitle.value = action.value?.title || 'Unknown Action'
  } catch (err) {
    console.error('Failed to fetch action:', err)
    window.showBigError('fetch-action', 'getting action details', err, false)
  }
}

function goToDashboard (path) {
  router.push(path)
}

function resetState () {
  action.value = null
  backToDashboards.value = []
  actionTitle.value = 'Action Details'
  logs.value = []
  totalCount.value = 0
  currentPage.value = 1
  searchText.value = ''
  loading.value = true
  syncDurationTicker()
}

function clearSearch () {
  searchText.value = ''
}

function formatTimestamp (timestamp) {
  if (!timestamp) return 'Unknown'
  try {
    const date = new Date(timestamp)
    return date.toLocaleString()
  } catch (err) {
    return timestamp
  }
}

function plural (n, singular, pluralForm) {
  return n === 1 ? `1 ${singular}` : `${n} ${pluralForm}`
}

function formatDurationSimple (ms) {
  if (!Number.isFinite(ms) || ms < 0) {
    return '—'
  }
  const totalSec = Math.round(ms / 1000)
  if (totalSec === 0) {
    return '0 seconds'
  }
  const days = Math.floor(totalSec / 86400)
  const hours = Math.floor((totalSec % 86400) / 3600)
  const minutes = Math.floor((totalSec % 3600) / 60)
  const seconds = totalSec % 60

  const parts = []
  if (days > 0) parts.push(plural(days, 'day', 'days'))
  if (hours > 0) parts.push(plural(hours, 'hour', 'hours'))
  if (minutes > 0) parts.push(plural(minutes, 'minute', 'minutes'))
  if (seconds > 0) parts.push(plural(seconds, 'second', 'seconds'))
  return parts.join(' ')
}

function formatExecutionDuration (log) {
  // Reading durationClock keeps this column reactive while executions are in progress.
  const clock = durationClock.value

  if (!log?.datetimeStarted) {
    return '—'
  }
  const started = new Date(log.datetimeStarted)
  if (Number.isNaN(started.getTime())) {
    return '—'
  }

  let endMs
  if (log.executionFinished) {
    const finished = new Date(log.datetimeFinished)
    if (Number.isNaN(finished.getTime())) {
      return '—'
    }
    endMs = finished.getTime()
  } else {
    endMs = clock
  }

  return formatDurationSimple(endMs - started.getTime())
}

function syncDurationTicker () {
  if (durationTicker != null) {
    clearInterval(durationTicker)
    durationTicker = null
  }
  const hasRunning = logs.value.some(l => !l.executionFinished)
  if (!hasRunning) {
    return
  }
  durationTicker = window.setInterval(() => {
    durationClock.value = Date.now()
  }, 1000)
}

function handlePageChange (page) {
  currentPage.value = page
  fetchActionLogs()
}

function handlePageSizeChange (newPageSize) {
  pageSize.value = newPageSize
  currentPage.value = 1
  fetchActionLogs()
}

async function startAction () {
  if (!action.value || !action.value.bindingId) {
    console.error('Cannot start action: no binding ID')
    return
  }

  if (needsArgumentForm(action.value)) {
    router.push(`/actionBinding/${action.value.bindingId}/argumentForm`)
    return
  }

  try {
    requestReconnectNow()
    const args = {
      bindingId: action.value.bindingId,
      arguments: []
    }

    const response = await window.client.startAction(args)
    router.push(`/logs/${response.executionTrackingId}`)
  } catch (err) {
    console.error('Failed to start action:', err)
    window.showBigError('start-action', 'starting action', err, false)
  }
}

function onExecutionEvent (evt) {
  const logEntry = getExecutionLogEntry(evt)
  if (!logEntry || logEntry.bindingId !== route.params.actionId) {
    return
  }

  if (!updateLogEntryInList(logs.value, logEntry)) {
    fetchActionLogs()
    return
  }
  syncDurationTicker()
}

onMounted(() => {
  fetchAction()
  fetchActionLogs()
  window.addEventListener('EventExecutionStarted', onExecutionEvent)
  window.addEventListener('EventExecutionFinished', onExecutionEvent)
})

watch(
  () => route.params.actionId,
  () => {
    resetState()
    fetchAction()
    fetchActionLogs()
  },
  { immediate: false }
)

onUnmounted(() => {
  window.removeEventListener('EventExecutionStarted', onExecutionEvent)
  window.removeEventListener('EventExecutionFinished', onExecutionEvent)
  if (durationTicker != null) {
    clearInterval(durationTicker)
    durationTicker = null
  }
})
</script>

<style scoped>
.section-title-with-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.action-title-icon {
  font-size: 1.5rem;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
}

.logs-table th {
  background-color: var(--section-background);
  padding: 0.5rem;
  text-align: left;
  font-weight: 600;
}

.logs-table td {
  padding: 0.5rem;
  border-top: 1px solid var(--border-color);
}

.log-row:hover {
  background-color: var(--hover-background);
}

.timestamp {
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.duration {
  font-size: 0.9rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
}

.filter-container {
  display: flex;
  justify-content: flex-end;
  padding: 0.5rem 1rem;
}

.input-with-icons {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 0.25rem;
  background: var(--section-background);
  width: 100%;
  max-width: 300px;
}

.input-with-icons input {
  border: none;
  outline: none;
  background: transparent;
  flex: 1;
  color: var(--text-primary);
}

.input-with-icons button {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary);
}

.input-with-icons button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.annotation {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.85rem;
}

.annotation-key {
  font-weight: 600;
  color: var(--text-secondary);
}

.annotation-val {
  color: var(--text-primary);
}

.tag-list {
  display: inline-flex;
  gap: 0.25rem;
}

.tag {
  background-color: var(--accent-color);
  color: var(--accent-text);
  padding: 0.1rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.85rem;
}

.padding {
  padding: 1rem;
}

.action-details-toolbar {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.action-group-list {
  margin: 0;
  padding-left: 0;
  list-style: none;
}

.action-group-row {
  padding: 0.25rem 0;
}

.action-group-name {
  font-family: monospace;
}

.action-groups-link {
  color: inherit;
  text-decoration: none;
}

.action-groups-link:hover {
  text-decoration: underline;
  color: var(--link-color, #007bff);
}
</style>
