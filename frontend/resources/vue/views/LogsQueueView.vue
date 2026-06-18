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
    v-for="actionGroup in groups"
    :key="actionGroup.name"
    class="with-header-and-content queue-action-group-section"
  >
    <div class="section-header flex-row">
      <div class="fg1 queue-action-group-heading">
        <ActionIconGlyph class="icon" :glyph="actionGroup.icon" />
        <h2>{{ displayActionGroupName(actionGroup.name) }}</h2>
      </div>
      <span
        v-if="actionGroup.maxConcurrent > 0"
        class="queue-action-group-limit annotation"
      >
        {{ t('logs.queue-group-limit', { max: actionGroup.maxConcurrent, queued: actionGroup.queuedCount }) }}
      </span>
    </div>

    <div class="section-content">
      <table class="logs-table row-hover">
        <thead>
          <tr>
            <th>{{ t('logs.timestamp') }}</th>
            <th>{{ t('logs.action') }}</th>
            <th>{{ t('logs.metadata') }}</th>
            <th>{{ t('logs.status') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="action in actionGroup.actions" :key="`${actionGroup.name}:${action.bindingId}`">
            <tr
              v-for="(entry, index) in action.entries"
              :key="entry.executionTrackingId"
              class="log-row"
              :title="action.actionTitle"
            >
              <td class="timestamp">{{ formatTimestamp(entry.datetimeStarted) }}</td>
              <td>
                <ActionIconGlyph class="icon" :glyph="action.actionIcon" />
                <router-link :to="`/logs/${entry.executionTrackingId}`">
                  <LogActionTitle
                    :action-title="action.actionTitle"
                    :justification="entry.justification"
                  />
                </router-link>
                <span v-if="action.entityPrefix" class="queue-entity annotation">
                  {{ action.entityPrefix }}
                </span>
              </td>
              <td class="tags">
                <span class="annotation">
                  <span class="annotation-key">User:</span>
                  <span class="annotation-val">{{ entry.user }}</span>
                </span>
                <span v-if="entry.tags && entry.tags.length > 0" class="tag-list">
                  <span v-for="tag in entry.tags" :key="tag" class="tag">{{ tag }}</span>
                </span>
              </td>
              <td class="exit-code">
                <span v-if="!entry.executionFinished" class="queue-position">
                  {{ t('logs.queue-position', { position: index + 1 }) }}
                </span>
                <ActionStatusDisplay :logEntry="entry" link-queued-status />
              </td>
            </tr>
          </template>
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
import LogActionTitle from '../components/LogActionTitle.vue'
import { useI18n } from 'vue-i18n'
import { getExecutionLogEntry, cloneLogEntry, updateLogEntryInGroups } from '../utils/executionLogEvents.js'

const defaultActionGroupName = 'default'

const { t } = useI18n()

const groups = ref([])
const loading = ref(false)

function displayActionGroupName (name) {
  if (name === defaultActionGroupName) {
    return t('logs.queue-default-group')
  }

  return name
}

function collectCompletedEntries (currentGroups) {
  const completed = []
  for (const group of currentGroups || []) {
    for (const action of group.actions || []) {
      for (const entry of action.entries || []) {
        if (entry.executionFinished) {
          completed.push(cloneLogEntry(entry))
        }
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
    if (left.name === defaultActionGroupName) {
      return 1
    }
    if (right.name === defaultActionGroupName) {
      return -1
    }
    return (left.name || '').localeCompare(right.name || '')
  })
}

function sumActionEntries (actions) {
  return (actions || []).reduce((total, action) => total + (action.entries || []).length, 0)
}

function cloneActionGroup (group) {
  return {
    ...group,
    actions: (group.actions || []).map(action => ({
      ...action,
      entries: [...(action.entries || [])]
    }))
  }
}

function findActionInGroups (groupList, bindingId) {
  const matches = []

  for (const group of groupList) {
    for (const action of group.actions || []) {
      if (action.bindingId === bindingId) {
        matches.push({ group, action })
      }
    }
  }

  return matches
}

function mergeCompletedEntries (apiGroups, completedEntries) {
  const merged = (apiGroups || []).map(cloneActionGroup)

  for (const entry of completedEntries) {
    const alreadyPresent = merged.some(group =>
      (group.actions || []).some(action =>
        (action.entries || []).some(item => item.executionTrackingId === entry.executionTrackingId)
      )
    )
    if (alreadyPresent) {
      continue
    }

    const matches = findActionInGroups(merged, entry.bindingId)
    if (matches.length === 0) {
      continue
    }

    for (const { action } of matches) {
      action.entries.push(entry)
    }
  }

  for (const group of merged) {
    for (const action of group.actions || []) {
      sortGroupEntries(action.entries)
      action.activeCount = action.entries.length
    }
    refreshGroupCounts(group)
  }
  sortGroups(merged)

  return merged
}

function forEachActionWithEntry (groupList, executionTrackingId, callback) {
  for (const group of groupList) {
    for (const action of group.actions || []) {
      const hasEntry = (action.entries || []).some(
        item => item.executionTrackingId === executionTrackingId
      )
      if (hasEntry) {
        callback(group, action)
      }
    }
  }
}

function refreshGroupCounts (group) {
  let queued = 0

  for (const action of group.actions || []) {
    action.activeCount = (action.entries || []).length
    for (const entry of action.entries || []) {
      if (entry.queued) {
        queued++
      }
    }
  }

  group.activeCount = sumActionEntries(group.actions)
  group.queuedCount = queued
}

function applyQueueEntryUpdate (logEntry, afterUpdate) {
  const result = updateLogEntryInGroups(groups.value, logEntry)
  if (!result) {
    return false
  }

  if (afterUpdate) {
    afterUpdate(result)
  }

  forEachActionWithEntry(groups.value, logEntry.executionTrackingId, (group, action) => {
    sortGroupEntries(action.entries)
    refreshGroupCounts(group)
  })

  return true
}

function insertActiveQueueEntry (logEntry) {
  if (!logEntry?.bindingId || !logEntry.executionTrackingId || logEntry.executionFinished) {
    fetchQueue()
    return
  }

  const matches = findActionInGroups(groups.value, logEntry.bindingId)
  if (matches.length === 0) {
    fetchQueue()
    return
  }

  const touchedGroups = new Set()
  for (const { group, action } of matches) {
    action.entries.push(cloneLogEntry(logEntry))
    touchedGroups.add(group)
    sortGroupEntries(action.entries)
  }

  for (const group of touchedGroups) {
    refreshGroupCounts(group)
  }
  sortGroups(groups.value)
}

function onExecutionStarted (evt) {
  const logEntry = getExecutionLogEntry(evt)
  if (!logEntry) {
    return
  }

  if (!applyQueueEntryUpdate(logEntry)) {
    insertActiveQueueEntry(logEntry)
  }
}

function onExecutionFinished (evt) {
  const logEntry = getExecutionLogEntry(evt)
  if (!logEntry) {
    return
  }

  applyQueueEntryUpdate(logEntry)
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
.queue-action-group-heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.queue-action-group-heading h2 {
  margin: 0;
}

.queue-action-group-limit {
  white-space: nowrap;
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

.queue-entity {
  display: block;
  margin-top: 0.25rem;
  color: #666;
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
