<template>
  <Section :title="t('logs.title')" :padding="false">
      <template #toolbar>
        <router-link to="/logs/queue" class="button neutral">
          {{ t('logs.queue') }}
        </router-link>
        <router-link to="/logs/calendar" class="button neutral">
          {{ t('logs.calendar') }}
        </router-link>
        <label class="input-with-icons">
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
            <path fill="currentColor"
              d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14" />
          </svg>
          <input
            :placeholder="t('logs.filter-placeholder')"
            v-model="searchText"
            list="logs-filter-suggestions"
            :aria-invalid="filterError ? 'true' : 'false'"
          />
          <datalist id="logs-filter-suggestions">
            <option v-for="suggestion in filterSuggestions" :key="suggestion" :value="suggestion" />
          </datalist>
          <button :title="t('logs.clear-filter')" :disabled="!searchText" @click="clearSearch">
            <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
              <path fill="currentColor"
                d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z" />
            </svg>
          </button>
        </label>
      </template>

      <div class="padding logs-intro">
        <p>{{ t('logs.page-description') }}</p>
        <details class="filter-help">
          <summary>{{ t('logs.filter-help-title') }}</summary>
          <p>{{ t('logs.filter-help-intro') }}</p>
          <p>{{ t('logs.filter-help-fields') }}</p>
          <p><code>{{ t('logs.filter-help-examples') }}</code></p>
        </details>
        <p v-if="filterError" class="filter-error" role="alert">{{ filterError }}</p>
      </div>

      <div v-show="logs.length > 0">
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
            <tr v-for="log in logs" :key="log.executionTrackingId" class="log-row" :title="log.actionTitle">
              <td class="timestamp">{{ formatTimestamp(log.datetimeStarted) }}</td>
              <td>
                <ActionIconGlyph class="icon" :glyph="log.actionIcon" />
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

        <Pagination :pageSize="pageSize" :total="totalCount" :currentPage="currentPage" :page="currentPage" @page-change="handlePageChange" class = "padding"
          @page-size-change="handlePageSizeChange" itemTitle="execution logs" />
      </div>

      <div v-show="logs.length === 0 && !loading && searchText && !filterError" class="empty-state padding">
        <p>{{ t('logs.no-logs-for-filter') }}</p>
        <button @click="clearSearch" class="button neutral">
          {{ t('logs.clear-filter') }}
        </button>
      </div>

      <div v-show="selectedDate && logs.length === 0 && !loading && !searchText" class="empty-state padding">
        <p>{{ t('logs.no-logs-to-display') }} {{ formatDateFilter(selectedDate) }}.</p>
        <button @click="clearDateFilter" class="button neutral">
          {{ t('logs.clear-date-filter') }}
        </button>
      </div>

      <div v-show="logs.length === 0 && !loading && !selectedDate && !searchText" class="empty-state padding">
        <p>{{ t('logs.no-logs-to-display') }}</p>
        <router-link to="/">{{ t('return-to-index') }}</router-link>
      </div>
  </Section>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ConnectError, Code } from '@connectrpc/connect'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useI18n } from 'vue-i18n'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'
import ActionIconGlyph from '../components/ActionIconGlyph.vue'

const route = useRoute()
const router = useRouter()

const logs = ref([])
const searchText = ref('')
const pageSize = ref(10)
const currentPage = ref(1)
const loading = ref(false)
const totalCount = ref(0)
const selectedDate = ref(null)
const filterError = ref('')
let fetchTimer = null

const filterSuggestions = [
  '!Update',
  'Status != Completed',
  'Status == Blocked',
  'Status == Running',
  'Action contains backup',
  'User == guest'
]

const { t } = useI18n()

function updateDateFromRoute() {
  const dateParam = route.query.date
  if (dateParam) {
    selectedDate.value = dateParam
  } else {
    selectedDate.value = null
  }
  fetchLogs()
}

watch(() => route.query.date, () => {
  updateDateFromRoute()
})

watch(searchText, () => {
  currentPage.value = 1
  scheduleFetchLogs()
})

async function fetchLogs() {
  loading.value = true
  filterError.value = ''
  try {
    const startOffset = (currentPage.value - 1) * pageSize.value

    const args = {
      startOffset: BigInt(startOffset),
      pageSize: BigInt(pageSize.value)
    }

    if (selectedDate.value) {
      args.dateFilter = selectedDate.value
    }

    if (searchText.value.trim()) {
      args.filter = searchText.value.trim()
    }

    const response = await window.client.getLogs(args)

    logs.value = response.logs
    totalCount.value = Number(response.totalCount) || 0
  } catch (err) {
    console.error('Failed to fetch logs:', err)
    if (err instanceof ConnectError && err.code === Code.InvalidArgument && searchText.value.trim()) {
      filterError.value = `${t('logs.filter-error')} ${err.message}`
      logs.value = []
      totalCount.value = 0
      return
    }
    window.showBigError('fetch-logs', 'getting logs', err, false)
  } finally {
    loading.value = false
  }
}

function scheduleFetchLogs() {
  if (fetchTimer) {
    clearTimeout(fetchTimer)
  }
  fetchTimer = setTimeout(() => {
    fetchLogs()
  }, 400)
}

function clearSearch() {
  searchText.value = ''
  currentPage.value = 1
  fetchLogs()
}

function clearDateFilter() {
  selectedDate.value = null
  const query = { ...route.query }
  delete query.date
  router.push({ path: route.path, query })
}

function formatDateFilter(dateString) {
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
  currentPage.value = 1
  fetchLogs()
}

onMounted(() => {
  updateDateFromRoute()
})
</script>

<style scoped>
.logs-intro {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.logs-intro p {
  margin: 0;
}

.filter-help summary {
  cursor: pointer;
  font-weight: 600;
}

.filter-help code {
  display: block;
  white-space: pre-wrap;
  margin-top: 0.5rem;
}

.filter-error {
  color: var(--karma-bad-fg, #b00020);
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
  max-width: 360px;
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
