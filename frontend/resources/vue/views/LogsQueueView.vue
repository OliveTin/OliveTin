<template>
  <Section :title="t('logs.queue-title')" :padding="false">
    <template #toolbar>
      <router-link to="/logs" class="button neutral">
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M20 11H7.83l5.59-5.59L12 4l-8 8l8 8l1.41-1.41L7.83 13H20z"/>
        </svg>
        {{ t('logs.back-to-list') }}
      </router-link>
    </template>

    <p class="padding">{{ t('logs.queue-page-description') }}</p>

    <div v-if="groups.length === 0 && !loading" class="empty-state padding">
      <p>{{ t('logs.queue-empty') }}</p>
      <router-link to="/logs">{{ t('logs.back-to-list') }}</router-link>
    </div>
  </Section>

  <section
    v-for="group in groups"
    :key="group.bindingId"
    class="with-header-and-content queue-group-section"
  >
    <div class="section-header flex-row">
      <div class="fg1 queue-group-heading">
        <ActionIconGlyph class="queue-group-icon" :glyph="group.actionIcon" />
        <div class="queue-group-title">
          <h2 :title="group.entityPrefix ? `${t('logs.queue-entity')}: ${group.entityPrefix}` : ''">
            {{ group.actionTitle }}
          </h2>
          <p v-if="group.entityPrefix" class="queue-entity">
            {{ t('logs.queue-entity') }}: {{ group.entityPrefix }}
          </p>
        </div>
      </div>
      <div role="toolbar" class="queue-group-toolbar">
        <router-link
          v-if="group.bindingId"
          :to="`/action/${group.bindingId}`"
          class="button neutral"
          :title="t('logs.queue-action-details')"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
            <path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm.31-8.86c-1.77-.45-2.34-.94-2.34-1.67 0-.84.79-1.43 2.1-1.43 1.38 0 1.9.66 1.94 1.64h1.71c-.05-1.34-.87-2.57-2.49-2.97V5H10.9v1.69c-1.51.32-2.72 1.3-2.72 2.81 0 1.79 1.49 2.69 3.66 3.21 1.95.46 2.34 1.22 2.34 1.8 0 .53-.39 1.39-2.1 1.39-1.6 0-2.05-.56-2.13-1.45H8.04c.08 1.5 1.18 2.37 2.82 2.69V19h2.34v-1.63c1.65-.35 2.48-1.24 2.48-2.77-.01-1.88-1.51-2.87-3.7-3.23z"/>
          </svg>
          {{ t('logs.queue-action-details') }}
        </router-link>
        <span class="queue-group-limit annotation">
          {{ t('logs.queue-group-active', { active: group.activeCount, max: group.maxConcurrent }) }}
        </span>
      </div>
    </div>

    <div class="section-content">
      <table class="logs-table row-hover">
      <thead>
        <tr>
          <th>{{ t('logs.timestamp') }}</th>
          <th>{{ t('logs.metadata') }}</th>
          <th>{{ t('logs.status') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(entry, index) in group.entries" :key="entry.executionTrackingId" class="log-row">
          <td class="timestamp">{{ formatTimestamp(entry.datetimeStarted) }}</td>
          <td class="tags">
            <span class="annotation">
              <span class="annotation-key">User:</span>
              <span class="annotation-val">{{ entry.user }}</span>
            </span>
            <span v-if="entry.tags && entry.tags.length > 0" class="tag-list">
              <span v-for="tag in entry.tags" :key="tag" class="tag">{{ tag }}</span>
            </span>
            <span class="annotation">
              <span class="annotation-key">ID:</span>
              <router-link :to="`/logs/${entry.executionTrackingId}`">
                {{ entry.executionTrackingId }}
              </router-link>
            </span>
          </td>
          <td class="exit-code">
            <span v-if="!entry.executionFinished" class="queue-position">{{ t('logs.queue-position', { position: index + 1 }) }}</span>
            <ActionStatusDisplay :logEntry="entry" />
          </td>
        </tr>
      </tbody>
    </table>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import ActionIconGlyph from '../components/ActionIconGlyph.vue'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'
import { useI18n } from 'vue-i18n'
import { getExecutionLogEntry, cloneLogEntry, updateLogEntryInGroups } from '../utils/executionLogEvents.js'

const { t } = useI18n()

const groups = ref([])
const loading = ref(false)

function collectCompletedEntries (currentGroups) {
  const completed = []
  for (const group of currentGroups || []) {
    for (const entry of group.entries || []) {
      if (entry.executionFinished) {
        completed.push(cloneLogEntry(entry))
      }
    }
  }
  return completed
}

function sortGroupEntries (entries) {
  entries.sort((left, right) => {
    if (left.executionFinished !== right.executionFinished) {
      return left.executionFinished ? 1 : -1
    }
    return (left.datetimeStarted || '').localeCompare(right.datetimeStarted || '')
  })
}

function sortGroups (groupList) {
  groupList.sort((left, right) => {
    const byTitle = (left.actionTitle || '').localeCompare(right.actionTitle || '')
    if (byTitle !== 0) {
      return byTitle
    }
    return (left.entityPrefix || '').localeCompare(right.entityPrefix || '')
  })
}

function mergeCompletedEntries (apiGroups, completedEntries) {
  const merged = (apiGroups || []).map(group => ({
    ...group,
    entries: [...(group.entries || [])]
  }))

  for (const entry of completedEntries) {
    const alreadyPresent = merged.some(group =>
      group.entries.some(item => item.executionTrackingId === entry.executionTrackingId)
    )
    if (alreadyPresent) {
      continue
    }

    let group = merged.find(item => item.bindingId === entry.bindingId)
    if (!group) {
      group = {
        bindingId: entry.bindingId,
        actionTitle: entry.actionTitle,
        actionIcon: entry.actionIcon,
        entityPrefix: '',
        maxConcurrent: 0,
        activeCount: 0,
        entries: []
      }
      merged.push(group)
    }

    group.entries.push(entry)
  }

  for (const group of merged) {
    sortGroupEntries(group.entries)
  }
  sortGroups(merged)

  return merged
}

function applyQueueEntryUpdate (logEntry, afterUpdate) {
  const result = updateLogEntryInGroups(groups.value, logEntry)
  if (!result) {
    return false
  }

  if (afterUpdate) {
    afterUpdate(result)
  }
  sortGroupEntries(result.group.entries)
  return true
}

function adjustActiveCountOnStart (group, previous, logEntry) {
  const wasActive = !previous.executionFinished
  const isActive = !logEntry.executionFinished
  if (!wasActive && isActive) {
    group.activeCount++
  }
}

function insertActiveQueueEntry (logEntry) {
  if (!logEntry?.bindingId || !logEntry.executionTrackingId || logEntry.executionFinished) {
    fetchQueue()
    return
  }

  let group = groups.value.find(item => item.bindingId === logEntry.bindingId)
  if (!group) {
    group = {
      bindingId: logEntry.bindingId,
      actionTitle: logEntry.actionTitle || '',
      actionIcon: logEntry.actionIcon || '',
      entityPrefix: logEntry.entityPrefix || '',
      maxConcurrent: 0,
      activeCount: 0,
      entries: []
    }
    groups.value.push(group)
  }

  group.entries.push(cloneLogEntry(logEntry))
  adjustActiveCountOnStart(group, { executionFinished: true }, logEntry)
  sortGroupEntries(group.entries)
  sortGroups(groups.value)
}

function onExecutionStarted (evt) {
  const logEntry = getExecutionLogEntry(evt)
  if (!logEntry) {
    return
  }

  if (!applyQueueEntryUpdate(logEntry, ({ group, previous }) => {
    adjustActiveCountOnStart(group, previous, logEntry)
  })) {
    insertActiveQueueEntry(logEntry)
  }
}

function onExecutionFinished (evt) {
  const logEntry = getExecutionLogEntry(evt)
  if (!logEntry) {
    return
  }

  applyQueueEntryUpdate(logEntry, ({ group, previous }) => {
    const wasActive = !previous.executionFinished
    if (wasActive && logEntry.executionFinished && group.activeCount > 0) {
      group.activeCount--
    }
  })
}

function formatTimestamp (timestamp) {
  if (!timestamp) {
    return 'Unknown'
  }
  try {
    return new Date(timestamp).toLocaleString()
  } catch (err) {
    return timestamp
  }
}

async function fetchQueue () {
  loading.value = true
  try {
    const completedEntries = collectCompletedEntries(groups.value)
    const response = await window.client.getExecutionQueue({})
    groups.value = mergeCompletedEntries(response.groups || [], completedEntries)
  } catch (err) {
    console.error('Failed to fetch execution queue:', err)
    window.showBigError('fetch-queue', 'getting execution queue', err, false)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchQueue()
  window.addEventListener('EventExecutionStarted', onExecutionStarted)
  window.addEventListener('EventExecutionFinished', onExecutionFinished)
})

onUnmounted(() => {
  window.removeEventListener('EventExecutionStarted', onExecutionStarted)
  window.removeEventListener('EventExecutionFinished', onExecutionFinished)
})
</script>

<style scoped>
.queue-group-heading {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.queue-group-title h2 {
  margin: 0;
}

.queue-group-toolbar {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.queue-group-limit {
  white-space: nowrap;
}

.queue-entity {
  margin: 0.25rem 0 0;
  font-size: 0.85rem;
  color: #666;
}

.queue-group-icon {
  font-size: 1.5em;
  flex-shrink: 0;
}

.timestamp {
  font-family: monospace;
  font-size: 0.875rem;
  color: #666;
}

.annotation {
  font-weight: 500;
  font-size: smaller;
}

.exit-code {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.queue-position {
  white-space: nowrap;
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
