<template>
  <Section :title="t('logs.calendar-title')" :padding="false">
    <template #toolbar>
      <router-link to="/logs" class="button neutral">
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M20 11H7.83l5.59-5.59L12 4l-8 8l8 8l1.41-1.41L7.83 13H20z"/>
        </svg>
        {{ t('logs.back-to-list') }}
      </router-link>
    </template>

    <div class="padding">
      <Calendar
        :events="calendarEvents"
        :loading="loading"
        :error="error"
        :current-month="currentMonthIndex"
        :current-year="currentYear"
        @event-click="handleEventClick"
        @date-click="handleDayClick"
        @month-change="handleMonthChange"
      />
    </div>
  </Section>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Calendar from 'picocrank/vue/components/Calendar.vue'
import Section from 'picocrank/vue/components/Section.vue'

const router = useRouter()
const { t } = useI18n()

const logs = ref([])
const loading = ref(false)
const error = ref(null)
const currentMonthIndex = ref(new Date().getMonth())
const currentYear = ref(new Date().getFullYear())

// Convert logs to calendar events format
const calendarEvents = computed(() => {
  return logs.value
    .filter(log => {
      // Only include logs with valid start dates
      if (!log.datetimeStarted) return false
      const startDate = new Date(log.datetimeStarted)
      return !isNaN(startDate.getTime())
    })
    .map(log => {
      const startDate = new Date(log.datetimeStarted)
      let endDate = log.datetimeFinished ? new Date(log.datetimeFinished) : null
      
      // Validate end date
      if (endDate && isNaN(endDate.getTime())) {
        endDate = null
      }

      return {
        id: log.executionTrackingId,
        title: log.actionTitle || 'Untitled Action',
        date: startDate,
        startDate: startDate,
        endDate: endDate,
        actionIcon: log.actionIcon,
        user: log.user,
        tags: log.tags,
        logEntry: log
      }
    })
})

async function fetchLogs() {
  loading.value = true
  error.value = null
  
  try {
    // Fetch a large number of logs to populate the calendar
    // We'll fetch more than a single page to get better calendar coverage
    const args = {
      "startOffset": BigInt(0),
    }

    const response = await window.client.getLogs(args)
    logs.value = response.logs || []
  } catch (err) {
    console.error('Failed to fetch logs:', err)
    error.value = 'Failed to load logs'
    window.showBigError('fetch-logs-calendar', 'getting logs for calendar', err, false)
  } finally {
    loading.value = false
  }
}

function handleEventClick(event) {
  // Navigate to the execution view when clicking on a calendar event
  if (event.id) {
    router.push(`/logs/${event.id}`)
  }
}

function handleDayClick(date) {
  // Navigate to logs list filtered by the selected date
  // Format date as YYYY-MM-DD for the query parameter
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const dateString = `${year}-${month}-${day}`
  router.push(`/logs?date=${dateString}`)
}

function handleMonthChange(month, year) {
  currentMonthIndex.value = month
  currentYear.value = year
  // Optionally fetch logs for the new month if needed
  // For now, we'll keep all logs loaded
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.padding {
  padding: 1rem;
}

@media (prefers-color-scheme: dark) {
  :deep(div.calendar-header-nav) {
    background-color: var(--bg, #111);
    color: var(--text-color, #fff);
    border-color: var(--border-color, #333);
  }

  :deep(div.calendar-header-nav h2.calendar-title) {
    color: #fff !important;
  }
}
</style>
