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

    <div v-if="groups.length > 0" class="queue-groups padding">
      <section v-for="group in groups" :key="group.bindingId" class="queue-group">
        <header class="queue-group-header">
          <ActionIconGlyph class="icon" :glyph="group.actionIcon" />
          <div class="queue-group-title">
            <h3>{{ group.actionTitle }}</h3>
            <p v-if="group.entityPrefix" class="queue-entity">
              {{ t('logs.queue-entity') }}: {{ group.entityPrefix }}
            </p>
          </div>
          <span class="queue-group-limit annotation">
            {{ t('logs.queue-group-active', { active: group.activeCount, max: group.maxConcurrent }) }}
          </span>
        </header>

        <table class="logs-table">
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
                <span class="annotation">
                  <span class="queue-position">{{ t('logs.queue-position', { position: index + 1 }) }}</span>
                  <span :class="queueStatusClass(entry)">{{ queueStatusText(entry) }}</span>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>

    <div v-else-if="!loading" class="empty-state padding">
      <p>{{ t('logs.queue-empty') }}</p>
      <router-link to="/logs">{{ t('logs.back-to-list') }}</router-link>
    </div>
  </Section>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import ActionIconGlyph from '../components/ActionIconGlyph.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const groups = ref([])
const loading = ref(false)

function queueStatusText (entry) {
  if (entry.executionStarted) {
    return t('logs.queue-running')
  }
  return t('logs.queue-waiting')
}

function queueStatusClass (entry) {
  return entry.executionStarted ? 'queue-status-running' : 'queue-status-waiting'
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
    const response = await window.client.getExecutionQueue({})
    groups.value = response.groups || []
  } catch (err) {
    console.error('Failed to fetch execution queue:', err)
    window.showBigError('fetch-queue', 'getting execution queue', err, false)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchQueue()
  window.addEventListener('EventExecutionStarted', fetchQueue)
  window.addEventListener('EventExecutionFinished', fetchQueue)
})

onUnmounted(() => {
  window.removeEventListener('EventExecutionStarted', fetchQueue)
  window.removeEventListener('EventExecutionFinished', fetchQueue)
})
</script>

<style scoped>
.queue-groups {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.queue-group {
  border: 1px solid var(--border-color, #ccc);
  border-radius: 0.5rem;
  overflow: hidden;
}

.queue-group-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--section-background, #f8f9fa);
  border-bottom: 1px solid var(--border-color, #ccc);
}

.queue-group-title {
  flex: 1;
}

.queue-group-title h3 {
  margin: 0;
  font-size: 1rem;
}

.queue-entity {
  margin: 0.25rem 0 0;
  font-size: 0.85rem;
  color: #666;
}

.queue-group-limit {
  white-space: nowrap;
}

.icon {
  font-size: 1.5em;
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

.queue-position {
  margin-right: 0.5rem;
}

.queue-status-running {
  color: var(--karma-warning-fg, #856404);
}

.queue-status-waiting {
  color: #0d6efd;
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
