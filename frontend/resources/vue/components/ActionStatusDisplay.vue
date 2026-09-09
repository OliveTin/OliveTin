<template>
  <router-link
    v-if="showQueueLink"
    to="/logs/queue"
    class="tag"
    :class="statusTagClass"
  >
    {{ statusText }}
  </router-link>
  <span
    v-else
    class="tag"
    :class="statusTagClass"
  >{{ statusText }}{{ exitCodeText }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  logEntry: {
    type: Object,
    required: true
  },
  linkQueuedStatus: {
    type: Boolean,
    default: false
  }
})

function isWaitingInQueue (logEntry) {
  return logEntry &&
        !logEntry.executionFinished &&
        !logEntry.executionStarted
}

const statusText = computed(() => {
  const logEntry = props.logEntry
  if (!logEntry) return 'unknown'

  if (logEntry.executionFinished) {
    if (logEntry.blocked) {
      return 'Blocked'
    } else if (logEntry.timedOut) {
      return 'Timed out'
    } else {
      return 'Completed'
    }
  }

  if (isWaitingInQueue(logEntry)) {
    return 'Queued'
  }

  return 'Still running...'
})

const exitCodeText = computed(() => {
  const logEntry = props.logEntry
  if (!logEntry) return ''
  if (logEntry.exitCode === 0) {
    return ''
  }
  if (logEntry.executionFinished) {
    if (logEntry.blocked || logEntry.timedOut) {
      return ''
    }
    return ' (Exit code: ' + logEntry.exitCode + ')'
  }
  return ''
})

const showQueueLink = computed(() => {
  return props.linkQueuedStatus && isWaitingInQueue(props.logEntry)
})

const statusTagClass = computed(() => {
  const logEntry = props.logEntry
  if (!logEntry) {
    return ''
  }

  if (!logEntry.executionFinished) {
    if (isWaitingInQueue(logEntry)) {
      return 'note'
    }
    return 'info'
  }

  if (logEntry.blocked) {
    return 'status-blocked'
  }
  if (logEntry.timedOut) {
    return ['warning', 'status-timeout']
  }
  if (logEntry.exitCode === 0) {
    return ['good', 'status-success']
  }
  return ['error', 'status-nonzero-exit']
})
</script>

<style scoped>
.tag {
  text-transform: none;
}

.tag.status-blocked {
  border-color: transparent;
  background-color: color-mix(in srgb, #ca79ff 30%, var(--standout-bg-color));
  color: var(--text-color);
}

a.tag {
  text-decoration: none;
}
</style>
