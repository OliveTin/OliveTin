<template>
  <Section :title="'Action Details: ' + actionTitle" :padding="false">
      <template #toolbar>
        <button v-if="action" @click="startAction" title="Start this action" class="button neutral">
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
            <path fill="currentColor" d="M8 6v12l8-6z" />
          </svg>
          Start
        </button>
      </template>

      <div class = "flex-row padding" v-if="action">
        <div class = "fg1">
          <dl>
            <dt>Title</dt>
            <dd>{{ action.title }}</dd>
            <dt>Timeout</dt>
            <dd>{{ action.timeout }} seconds</dd>
          </dl>
          <p v-if="action" class = "fg1">
            Execution history for this action. You can filter by execution tracking ID.
          </p>
        </div>
        <div style = "align-self: start; text-align: right;">
          <span class="icon" v-html="action.icon"></span>

          <div class="filter-container">
            <label class="input-with-icons">
              <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
                <path fill="currentColor"
                  d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14" />
              </svg>
              <input placeholder="Filter current page" v-model="searchText" />
              <button title="Clear search filter" :disabled="!searchText" @click="clearSearch">
                <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
                  <path fill="currentColor"
                    d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z" />
                </svg>
              </button>
            </label>
          </div>
        </div>
      </div>

      <div v-show="filteredLogs.length > 0">
        <table class="logs-table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Execution ID</th>
              <th>Metadata</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in filteredLogs" :key="log.executionTrackingId" class="log-row" :title="log.actionTitle">
              <td class="timestamp">{{ formatTimestamp(log.datetimeStarted) }}</td>
              <td>
                <router-link :to="`/logs/${log.executionTrackingId}`">
                  {{ log.executionTrackingId }}
                </router-link>
              </td>
              <td class="tags">
                <span class="annotation">
                  <span class="annotation-key">User:</span>
                  <span class="annotation-val">{{ log.user }}</span>
                </span>
                <span v-if="log.tags && log.tags.length > 0" class="tag-list">
                  <span v-for="tag in log.tags" :key="tag" class="tag">{{ tag }}</span>
                </span>
              </td>
              <td class="exit-code">
                <span :class="getStatusClass(log) + ' annotation'">
                  {{ getStatusText(log) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>

        <Pagination :pageSize="pageSize" :total="totalCount" :currentPage="currentPage" @page-change="handlePageChange" class="padding"
          @page-size-change="handlePageSizeChange" itemTitle="execution logs" />
      </div>

      <div v-show="logs.length === 0 && !loading" class="empty-state">
        <p>This action has no execution history.</p>
        <router-link to="/">Return to index</router-link>
      </div>
  </Section>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Pagination from '../components/Pagination.vue'
import Section from 'picocrank/vue/components/Section.vue'

const route = useRoute()
const router = useRouter()
const logs = ref([])
const action = ref(null)
const actionTitle = ref('Action Details')
const searchText = ref('')
const pageSize = ref(10)
const currentPage = ref(1)
const loading = ref(false)
const totalCount = ref(0)

const filteredLogs = computed(() => {
  if (!searchText.value) {
    return logs.value
  }
  const searchLower = searchText.value.toLowerCase()
  return logs.value.filter(log =>
    log.executionTrackingId.toLowerCase().includes(searchLower) ||
    log.actionTitle.toLowerCase().includes(searchLower)
  )
})

async function fetchActionLogs() {
  loading.value = true
  try {
    const actionId = route.params.actionId
    const startOffset = (currentPage.value - 1) * pageSize.value

    const args = {
      "actionId": actionId,
      "startOffset": BigInt(startOffset),
    }

    const response = await window.client.getActionLogs(args)

    logs.value = response.logs
    pageSize.value = Number(response.pageSize) || 0
    totalCount.value = Number(response.totalCount) || 0
  } catch (err) {
    console.error('Failed to fetch action logs:', err)
    window.showBigError('fetch-action-logs', 'getting action logs', err, false)
  } finally {
    loading.value = false
  }
}

async function fetchAction() {
  try {
    const actionId = route.params.actionId
    const args = {
      "bindingId": actionId
    }
    const response = await window.client.getActionBinding(args)
    action.value = response.action
    actionTitle.value = action.value?.title || 'Unknown Action'
  } catch (err) {
    console.error('Failed to fetch action:', err)
  }
}

function resetState() {
  action.value = null
  actionTitle.value = 'Action Details'
  logs.value = []
  totalCount.value = 0
  currentPage.value = 1
  searchText.value = ''
  loading.value = true
}

function clearSearch() {
  searchText.value = ''
}

function formatTimestamp(timestamp) {
  if (!timestamp) return 'Unknown'
  try {
    const date = new Date(timestamp)
    return date.toLocaleString()
  } catch (err) {
    return timestamp
  }
}

function getStatusClass(log) {
  if (log.timedOut) return 'status-timeout'
  if (log.blocked) return 'status-blocked'
  if (log.exitCode !== 0) return 'status-error'
  return 'status-success'
}

function getStatusText(log) {
  if (log.timedOut) return 'Timed out'
  if (log.blocked) return 'Blocked'
  if (log.exitCode !== 0) return `Exit code ${log.exitCode}`
  return 'Completed'
}

function handlePageChange(page) {
  currentPage.value = page
  fetchActionLogs()
}

function handlePageSizeChange(newPageSize) {
  pageSize.value = newPageSize
  currentPage.value = 1
  fetchActionLogs()
}

async function startAction() {
  if (!action.value || !action.value.bindingId) {
    console.error('Cannot start action: no binding ID')
    return
  }

  try {
    const args = {
      "bindingId": action.value.bindingId,
      "arguments": []
    }

    const response = await window.client.startAction(args)
    router.push(`/logs/${response.executionTrackingId}`)
  } catch (err) {
    console.error('Failed to start action:', err)
    window.showBigError('start-action', 'starting action', err, false)
  }
}

onMounted(() => {
  fetchAction()
  fetchActionLogs()
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
</script>

<style scoped>
.action-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.action-header h2 {
  margin: 0;
}

.icon {
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

.exit-code .status-success {
  color: #28a745;
}

.exit-code .status-error {
  color: #dc3545;
}

.exit-code .status-timeout {
  color: #ffc107;
}

.exit-code .status-blocked {
  color: #6c757d;
}

.padding {
  padding: 1rem;
}
</style>

