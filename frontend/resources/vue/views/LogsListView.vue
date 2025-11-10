<template>
  <Section :title="t('logs.title')" :padding="false">
      <template #toolbar>
        <label class="input-with-icons">
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
            <path fill="currentColor"
              d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14" />
          </svg>
          <input :placeholder="t('search-filter')" v-model="searchText" />
          <button title="Clear search filter" :disabled="!searchText" @click="clearSearch">
            <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
              <path fill="currentColor"
                d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z" />
            </svg>
          </button>
        </label>
      </template>

      <p class = "padding">{{ t('logs.page-description') }}</p>
      <div v-show="filteredLogs.length > 0">
        <table class="logs-table">
          <thead>
            <tr>
              <th>{{ t('logs.timestamp') }}</th>
              <th>{{ t('logs.action') }}</th>
              <th>{{ t('logs.metadata') }}</th>
              <th>{{ t('logs.status') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in filteredLogs" :key="log.executionTrackingId" class="log-row" :title="log.actionTitle">
              <td class="timestamp">{{ formatTimestamp(log.datetimeStarted) }}</td>
                <span class="icon" v-html="log.actionIcon"></span>
                <router-link :to="`/logs/${log.executionTrackingId}`">
                  {{ log.actionTitle }}
                </router-link>
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

        <Pagination :pageSize="pageSize" :total="totalCount" :currentPage="currentPage" @page-change="handlePageChange" class = "padding"
          @page-size-change="handlePageSizeChange" itemTitle="execution logs" />
      </div>

      <div v-show="logs.length === 0" class="empty-state">
        <p>{{ t('logs.no-logs-to-display') }}</p>
        <router-link to="/">{{ t('return-to-index') }}</router-link>
      </div>
  </Section>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Pagination from '../components/Pagination.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useI18n } from 'vue-i18n'

const logs = ref([])
const searchText = ref('')
const pageSize = ref(10)
const currentPage = ref(1)
const loading = ref(false)
const totalCount = ref(0)

const { t } = useI18n()

const filteredLogs = computed(() => {
  if (!searchText.value) {
    return logs.value
  }
  const searchLower = searchText.value.toLowerCase()
  return logs.value.filter(log =>
    log.actionTitle.toLowerCase().includes(searchLower)
  )
})

async function fetchLogs() {
  loading.value = true
  try {
    const startOffset = (currentPage.value - 1) * pageSize.value

    const args = {
      "startOffset": BigInt(startOffset),
    }

    const response = await window.client.getLogs(args)

    logs.value = response.logs
    pageSize.value = Number(response.pageSize) || 0
    totalCount.value = Number(response.totalCount) || 0
  } catch (err) {
    console.error('Failed to fetch logs:', err)
    window.showBigError('fetch-logs', 'getting logs', err, false)
  } finally {
    loading.value = false
  }
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
  fetchLogs()
}

function handlePageSizeChange(newPageSize) {
  pageSize.value = newPageSize
  currentPage.value = 1 // Reset to first page
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.logs-view {
  padding: 1rem;
}

.input-with-icons {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 0.5rem;
}

.input-with-icons input {
  border: none;
  outline: none;
  flex: 1;
  font-size: 1rem;
}

.input-with-icons button {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 3px;
}

.input-with-icons button:hover:not(:disabled) {
}

.input-with-icons button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.timestamp {
  font-family: monospace;
  font-size: 0.875rem;
  color: #666;
}

.icon {
  margin-right: 0.5rem;
  font-size: 1.2em;
}

.content {
  color: #007bff;
  text-decoration: none;
  cursor: pointer;
}

.content:hover {
  text-decoration: underline;
}

.status-success {
  color: #28a745;
  font-weight: 500;
}

.status-error {
  color: #dc3545;
  font-weight: 500;
}

.status-timeout {
  color: #ffc107;
  font-weight: 500;
}

.status-blocked {
  color: #6c757d;
  font-weight: 500;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.empty-state a {
  color: #007bff;
  text-decoration: none;
}

.empty-state a:hover {
  text-decoration: underline;
}

</style>