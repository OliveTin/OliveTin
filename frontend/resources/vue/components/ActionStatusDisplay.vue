<template>
  <div :class="statusClass + ' annotation'">
    <router-link
      v-if="showQueueLink"
      to="/logs/queue"
      class="queue-status-link"
    >
      {{ statusText }}
    </router-link>
    <span v-else>{{ statusText }}</span><span>{{ exitCodeText }}</span>
  </div>
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

const statusClass = computed(() => {
  const logEntry = props.logEntry
  if (!logEntry) return ''
  if (logEntry.executionFinished) {
    if (logEntry.blocked) {
      return 'status-blocked'
    } else if (logEntry.timedOut) {
      return 'status-timeout'
    } else if (logEntry.exitCode === 0) {
      return 'status-success'
    } else {
      return 'status-nonzero-exit'
    }
  }
  return ''
})
</script>

<style scoped>
.status-success {
  color: var(--karma-good-fg);
}

.status-nonzero-exit {
  color: var(--karma-bad-fg);
}

.status-timeout {
  color: var(--karma-warning-fg);
}

.status-blocked {
  color: #ca79ff;
}

.queue-status-link {
  color: #0d6efd;
  text-decoration: none;
}

.queue-status-link:hover {
  text-decoration: underline;
}

</style>
