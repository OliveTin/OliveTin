<template>
    <div :class = "statusClass + ' annotation'">
        <span>{{ statusText }}</span><span>{{ exitCodeText }}</span>
    </div>

</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
    logEntry: {
        type: Object,
        required: true
    }
})

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
    } else {
        return 'Still running...'
    }
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


</style>