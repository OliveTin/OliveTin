<template>
  <Table
    :data="logs"
    :headers="headers"
    :table-id="tableId"
    :show-pagination="false"
    :filterable="false"
    :loading="loading"
    row-key="executionTrackingId"
  >
    <template #cell-timestamp="{ row }">
      <span class="timestamp">{{ formatTimestamp(row.datetimeStarted) }}</span>
    </template>

    <template
      v-if="variant === 'action-history'"
      #cell-duration="{ row }"
    >
      <span class="duration">
        <slot
          name="duration"
          :row="row"
        >
          {{ row.duration }}
        </slot>
      </span>
    </template>

    <template #cell-executionId="{ row }">
      <router-link
        :to="`/logs/${row.executionTrackingId}`"
        class="execution-id-link"
      >
        <LogActionTitle :justification="row.justification">
          {{ row.executionTrackingId }}
        </LogActionTitle>
      </router-link>
    </template>

    <template
      v-if="variant === 'standard'"
      #cell-action="{ row }"
    >
      <ActionIconGlyph
        class="icon"
        :glyph="row.actionIcon"
      />
      <router-link
        v-if="row.bindingId"
        :to="{ name: 'ActionDetails', params: { actionId: row.bindingId } }"
      >
        <LogActionTitle
          :action-title="row.actionTitle"
          :justification="row.justification"
        />
      </router-link>
      <LogActionTitle
        v-else
        :action-title="row.actionTitle"
        :justification="row.justification"
      />
      <span
        v-if="row.entityPrefix"
        class="queue-entity annotation"
      >
        {{ row.entityPrefix }}
      </span>
    </template>

    <template #cell-metadata="{ row }">
      <span class="tags">
        <span class="annotation">
          <span class="annotation-key">User:</span>
          <span class="annotation-val">{{ row.user }}</span>
        </span>
        <span
          v-if="row.tags && row.tags.length > 0"
          class="tag-list"
        >
          <span
            v-for="tag in row.tags"
            :key="tag"
            class="tag"
          >{{ tag }}</span>
        </span>
      </span>
    </template>

    <template #cell-status="{ row }">
      <span class="exit-code">
        <span
          v-if="variant === 'standard' && row.queuePosition != null && !row.executionFinished"
          class="queue-position"
        >
          {{ t('logs.queue-position', { position: row.queuePosition }) }}
        </span>
        <ActionStatusDisplay
          :log-entry="row"
          :link-queued-status="true"
        />
      </span>
    </template>
  </Table>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Table from 'picocrank/vue/components/Table.vue'
import ActionIconGlyph from './ActionIconGlyph.vue'
import ActionStatusDisplay from './ActionStatusDisplay.vue'
import LogActionTitle from './LogActionTitle.vue'

const props = defineProps({
  logs: {
    type: Array,
    required: true
  },
  variant: {
    type: String,
    default: 'standard',
    validator: value => ['standard', 'action-history'].includes(value)
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const { t } = useI18n()

const tableId = computed(() =>
  props.variant === 'action-history'
    ? 'olivetin-execution-logs-action-history'
    : 'olivetin-execution-logs-standard'
)

const headers = computed(() => {
  if (props.variant === 'action-history') {
    return [
      { key: 'timestamp', label: 'Timestamp', sortable: false },
      { key: 'duration', label: 'Duration', sortable: false },
      { key: 'executionId', label: 'Execution ID', sortable: false },
      { key: 'metadata', label: 'Metadata', sortable: false },
      { key: 'status', label: 'Status', sortable: false }
    ]
  }

  return [
    { key: 'timestamp', label: t('logs.timestamp'), sortable: false },
    { key: 'action', label: t('logs.action'), sortable: false },
    { key: 'executionId', label: t('logs.execution-id'), sortable: false },
    { key: 'metadata', label: t('logs.metadata'), sortable: false },
    { key: 'status', label: t('logs.status'), sortable: false }
  ]
})

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
</script>

<style scoped>
.timestamp {
  font-family: monospace;
  font-size: 0.875rem;
  color: var(--muted-text-color, #666);
}

.duration {
  font-size: 0.9rem;
  color: var(--muted-text-color, #666);
  white-space: nowrap;
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
  color: var(--muted-text-color, #666);
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.exit-code {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.queue-position {
  white-space: nowrap;
}

.execution-id-link {
  font-family: monospace;
  font-size: 0.875rem;
  text-decoration: none;
}

.execution-id-link:hover {
  text-decoration: underline;
}
</style>
