<template>
  <Section :title="t('logs.title')" :padding="false">
      <template #toolbar>
        <router-link to="/logs/calendar" class="button neutral">
          Calendar
        </router-link>
        <label class="input-with-icons">
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
            <path fill="currentColor"
              d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14" />
          </svg>
          <input :placeholder="t('search-filter')" v-model="searchText" />
          <button :title="t('logs.clear-filter')" :disabled="!searchText" @click="clearSearch">
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
              <th>
                <div class="timestamp-header">
                  <span>{{ t('logs.timestamp') }}</span>
                  <span v-if="selectedDate" class="date-filter-indicator">
                    <span class="date-filter-text">{{ formatDateFilter(selectedDate) }}</span>
                    <button :title="t('logs.clear-date-filter')" @click="clearDateFilter" class="clear-date-button">
                      <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
                        <path fill="currentColor"
                          d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z" />
                      </svg>
                    </button>
                  </span>
                </div>
              </th>
              <th>{{ t('logs.action') }}</th>
              <th>{{ t('logs.metadata') }}</th>
              <th>{{ t('logs.status') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in filteredLogs" :key="log.executionTrackingId" class="log-row" :title="log.actionTitle">
              <td class="timestamp">{{ formatTimestamp(log.datetimeStarted) }}</td>
              <td>
                <span class="icon" v-html="log.actionIcon"></span>
                <router-link :to="`/logs/${log.executionTrackingId}`">
                  {{ log.actionTitle }}
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
                <ActionStatusDisplay :logEntry="log" />
              </td>
            </tr>
          </tbody>
        </table>

        <Pagination :pageSize="pageSize" :total="totalCount" :currentPage="currentPage" @page-change="handlePageChange" class = "padding"
          @page-size-change="handlePageSizeChange" itemTitle="execution logs" />
      </div>

      <div v-show="selectedDate && filteredLogs.length === 0" class="empty-state">
        <p>No logs found for {{ formatDateFilter(selectedDate) }}.</p>
        <button @click="clearDateFilter" class="button neutral">
          Clear date filter
        </button>
      </div>

      <div v-show="logs.length === 0 && !selectedDate" class="empty-state">
        <p>{{ t('logs.no-logs-to-display') }}</p>
        <router-link to="/">{{ t('return-to-index') }}</router-link>
      </div>
  </Section>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useI18n } from 'vue-i18n'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'

const route = useRoute()
const router = useRouter()

const logs = ref([])
const searchText = ref('')
const pageSize = ref(10)
const currentPage = ref(1)
const loading = ref(false)
const totalCount = ref(0)
const selectedDate = ref(null)

const { t } = useI18n()

// Read date query parameter from route
function updateDateFromRoute() {
  const dateParam = route.query.date
  if (dateParam) {
    selectedDate.value = dateParam
  } else {
    selectedDate.value = null
  }
}

// Watch for route changes to update date filter
watch(() => route.query.date, () => {
  updateDateFromRoute()
})

const filteredLogs = computed(() => {
  let result = logs.value
  
  // Filter by selected date if present
  if (selectedDate.value) {
    result = result.filter(log => {
      if (!log.datetimeStarted) return false
      const logDate = new Date(log.datetimeStarted)
      const logDateString = `${logDate.getFullYear()}-${String(logDate.getMonth() + 1).padStart(2, '0')}-${String(logDate.getDate()).padStart(2, '0')}`
      return logDateString === selectedDate.value
    })
  }
  
  if (searchText.value) {
    const searchLower = searchText.value.toLowerCase()
    result = result.filter(log =>
      log.actionTitle.toLowerCase().includes(searchLower)
    )
  }
  
  // Sort by timestamp with most recent first
  return [...result].sort((a, b) => {
    const dateA = a.datetimeStarted ? new Date(a.datetimeStarted).getTime() : 0
    const dateB = b.datetimeStarted ? new Date(b.datetimeStarted).getTime() : 0
    return dateB - dateA // Descending order (most recent first)
  })
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

function clearDateFilter() {
  selectedDate.value = null
  // Remove date query parameter from URL
  const query = { ...route.query }
  delete query.date
  router.push({ path: route.path, query })
}

function formatDateFilter(dateString) {
  // Format YYYY-MM-DD to a short format (e.g., "Jan 15, 2024")
  try {
    const date = new Date(dateString + 'T00:00:00')
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch (err) {
    return dateString
  }
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

function handlePageChange(page) {
  currentPage.value = page
  fetchLogs()
}

function handlePageSizeChange(newPageSize) {
  pageSize.value = newPageSize
  currentPage.value = 1 // Reset to first page
}

onMounted(() => {
  updateDateFromRoute()
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
  padding: 0.25rem;
  border-radius: 3px;
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

.annotation {
  font-weight: 500;
  font-size: smaller;
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

.timestamp-header {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.date-filter-indicator {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  font-weight: normal;
  color: var(--text-secondary, #666);
  white-space: nowrap;
}

.date-filter-text {
  font-style: italic;
}

.timestamp-header .clear-date-button {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.125rem;
  border-radius: 3px;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.timestamp-header .clear-date-button:hover {
  opacity: 1;
  background: var(--hover-background, rgba(0, 0, 0, 0.05));
}

.timestamp-header .clear-date-button svg {
  width: 0.75rem;
  height: 0.75rem;
}

</style>
