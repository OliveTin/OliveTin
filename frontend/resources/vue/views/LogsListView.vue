<template>
  <Section
    :title="t('logs.title')"
    :icon="LeftToRightListDashIcon"
    :padding="false"
  >
    <template #toolbar>
      <router-link
        to="/logs/queue"
        class="button neutral inline-icon"
      >
        <HugeiconsIcon :icon="Queue01Icon" />
        {{ t('logs.queue') }}
      </router-link>
      <router-link
        to="/logs/calendar"
        class="button neutral inline-icon"
      >
        <HugeiconsIcon :icon="Calendar01Icon" />
        {{ t('logs.calendar') }}
      </router-link>
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
          :placeholder="t('logs.filter-placeholder')"
          list="logs-filter-suggestions"
          :aria-invalid="filterError ? 'true' : 'false'"
        >
        <datalist id="logs-filter-suggestions">
          <option
            v-for="suggestion in filterSuggestions"
            :key="suggestion"
            :value="suggestion"
          />
        </datalist>
        <button
          :title="t('logs.clear-filter')"
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
    </template>

    <div class="padding logs-intro">
      <p>{{ t('logs.page-description') }}</p>
      <details class="filter-help">
        <summary>{{ t('logs.filter-help-title') }}</summary>
        <p>{{ t('logs.filter-help-intro') }}</p>
        <p>{{ t('logs.filter-help-fields') }}</p>
        <p><code>{{ t('logs.filter-help-examples') }}</code></p>
      </details>
      <p
        v-if="filterError"
        class="filter-error"
        role="alert"
      >
        {{ filterError }}
      </p>
      <p
        v-if="selectedDate"
        class="date-filter-banner"
      >
        <span>{{ formatDateFilter(selectedDate) }}</span>
        <button
          type="button"
          class="button neutral inline-icon"
          :title="t('logs.clear-date-filter')"
          @click="clearDateFilter"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="1em"
            height="1em"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              fill="currentColor"
              d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z"
            />
          </svg>
          {{ t('logs.clear-date-filter') }}
        </button>
      </p>
    </div>

    <div v-show="logs.length > 0">
      <ExecutionLogsTable
        :logs="logs"
        :loading="loading"
      />

      <Pagination
        :page-size="pageSize"
        :total="totalCount"
        :current-page="currentPage"
        class="padding"
        item-title="execution logs"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      />
    </div>

    <div
      v-show="logs.length === 0 && !loading && searchText && !filterError"
      class="empty-state padding"
    >
      <p>{{ t('logs.no-logs-for-filter') }}</p>
      <button
        class="button neutral"
        @click="clearSearch"
      >
        {{ t('logs.clear-filter') }}
      </button>
    </div>

    <div
      v-show="selectedDate && logs.length === 0 && !loading && !searchText"
      class="empty-state padding"
    >
      <p>{{ t('logs.no-logs-to-display') }} {{ formatDateFilter(selectedDate) }}.</p>
      <button
        class="button neutral"
        @click="clearDateFilter"
      >
        {{ t('logs.clear-date-filter') }}
      </button>
    </div>

    <div
      v-show="logs.length === 0 && !loading && !selectedDate && !searchText"
      class="empty-state padding"
    >
      <p>{{ t('logs.no-logs-to-display') }}</p>
      <router-link to="/">
        {{ t('return-to-index') }}
      </router-link>
    </div>
  </Section>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ConnectError, Code } from '@connectrpc/connect'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useI18n } from 'vue-i18n'
import { HugeiconsIcon } from '@hugeicons/vue'
import { Calendar01Icon, LeftToRightListDashIcon, Queue01Icon } from '@hugeicons/core-free-icons'
import ExecutionLogsTable from '../components/ExecutionLogsTable.vue'
import { getExecutionLogEntry, updateLogEntryInList } from '../utils/executionLogEvents.js'
import { loadStoredLogsFilter, storeLogsFilter } from '../utils/logsFilterStorage.js'
const route = useRoute()
const router = useRouter()

function readInitialFilter () {
  // Prefer ?filter= when present (e.g. calendar keeps it while adding ?date=).
  // Otherwise restore from sessionStorage so sidebar / breadcrumb returns keep the filter.
  if (typeof route.query.filter === 'string' && route.query.filter !== '') {
    storeLogsFilter(route.query.filter)
    return route.query.filter
  }

  return loadStoredLogsFilter()
}

const logs = ref([])
const searchText = ref(readInitialFilter())
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

function updateDateFromRoute () {
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

watch(searchText, (value) => {
  currentPage.value = 1
  storeLogsFilter(value)
  syncFilterToRoute(value)
  scheduleFetchLogs()
})

watch(() => route.query.filter, (filter) => {
  const next = typeof filter === 'string' ? filter : ''
  if (searchText.value === next) {
    return
  }
  searchText.value = next
})

function syncFilterToRoute (value) {
  const next = value || ''
  const current = typeof route.query.filter === 'string' ? route.query.filter : ''
  if (next === current) {
    return
  }

  const query = { ...route.query }
  if (next) {
    query.filter = next
  } else {
    delete query.filter
  }
  router.replace({ path: route.path, query })
}

async function fetchLogs () {
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

function scheduleFetchLogs () {
  if (fetchTimer) {
    clearTimeout(fetchTimer)
  }
  fetchTimer = setTimeout(() => {
    fetchLogs()
  }, 400)
}

function clearSearch () {
  searchText.value = ''
}

function clearDateFilter () {
  selectedDate.value = null
  const query = { ...route.query }
  delete query.date
  router.push({ path: route.path, query })
}

function formatDateFilter (dateString) {
  try {
    const date = new Date(dateString + 'T00:00:00')
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch (err) {
    return dateString
  }
}

function handlePageChange (page) {
  currentPage.value = page
  fetchLogs()
}

function handlePageSizeChange (newPageSize) {
  pageSize.value = newPageSize
  currentPage.value = 1
  fetchLogs()
}

function onExecutionEvent (evt) {
  const logEntry = getExecutionLogEntry(evt)
  if (!logEntry) {
    return
  }

  if (!updateLogEntryInList(logs.value, logEntry)) {
    fetchLogs()
  }
}

onMounted(() => {
  updateDateFromRoute()
  window.addEventListener('EventExecutionStarted', onExecutionEvent)
  window.addEventListener('EventExecutionFinished', onExecutionEvent)
})

onUnmounted(() => {
  window.removeEventListener('EventExecutionStarted', onExecutionEvent)
  window.removeEventListener('EventExecutionFinished', onExecutionEvent)
  if (fetchTimer) {
    clearTimeout(fetchTimer)
    fetchTimer = null
  }
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

.date-filter-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  margin: 0;
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
