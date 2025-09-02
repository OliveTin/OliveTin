<template>
    <span>
        <span :class="['action-status', statusClass]">{{ statusText }}</span><span>{{ exitCodeText }}</span>
    </span>

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
    if (logEntry.executionFinished) {
        if (logEntry.blocked || logEntry.timedOut) {
            return ''
        }
        return ' Exit code: ' + logEntry.exitCode
    }
    return ''
})

const statusClass = computed(() => {
    const logEntry = props.logEntry
    if (!logEntry) return ''
    if (logEntry.executionFinished) {
        if (logEntry.blocked) {
            return 'action-blocked'
        } else if (logEntry.timedOut) {
            return 'action-timeout'
        } else if (logEntry.exitCode === 0) {
            return 'action-success'
        } else {
            return 'action-nonzero-exit'
        }
    }
    return ''
})
</script>
