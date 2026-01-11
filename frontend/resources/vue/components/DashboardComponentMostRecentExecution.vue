<template>
  <div class="mre-container" :class="component.cssClass">   
    <router-link 
        v-if="executionTrackingId" 
        :to="`/logs/${executionTrackingId}`" 
        class="mre-link"
    >
        <pre class="mre-output">{{ output }}</pre>
    </router-link>
    <pre v-else class="mre-output fg-important">{{ output }}</pre>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { buttonResults } from '../stores/buttonResults'

const props = defineProps({
  component: {
    type: Object,
    required: true
  }
})

const output = ref('Waiting...')
const executionTrackingId = ref(null)
let unwatchButtonResults = null

function updateFromLogEntry(logEntry) {
  if (logEntry) {
    if (logEntry.output !== undefined) {
      output.value = logEntry.output
    } else {
      output.value = 'No output available'
    }
    if (logEntry.executionTrackingId) {
      executionTrackingId.value = logEntry.executionTrackingId
    }
  }
}

async function fetchMostRecentExecution() {
  if (!props.component.title) {
    output.value = 'Error: No action ID specified'
    executionTrackingId.value = null
    return
  }

  if (!window.client) {
    output.value = 'Error: Client not initialized'
    executionTrackingId.value = null
    return
  }

  try {
    const executionStatusArgs = {
      actionId: props.component.title
    }

    const result = await window.client.executionStatus(executionStatusArgs)
    
    if (result.logEntry) {
      updateFromLogEntry(result.logEntry)
    } else {
      output.value = 'No output available'
      executionTrackingId.value = null
    }
  } catch (err) {
    if (err.code === 'NotFound' || err.status === 404) {
      output.value = 'No execution found'
      executionTrackingId.value = null
    } else {
      output.value = 'Error: ' + (err.message || 'Failed to fetch execution')
      console.error('Failed to fetch most recent execution:', err)
      executionTrackingId.value = null
    }
  }
}

onMounted(() => {
  fetchMostRecentExecution()
  
  unwatchButtonResults = watch(
    buttonResults,
    () => {
      // Find the most recent finished execution for this bindingId
      const bindingId = props.component.title
      let mostRecent = null
      let mostRecentTime = null
      
      for (const trackingId in buttonResults) {
        const logEntry = buttonResults[trackingId]
        if (logEntry && logEntry.bindingId === bindingId && logEntry.executionFinished) {
          const finishedTime = new Date(logEntry.datetimeFinished)
          if (!mostRecent || finishedTime > mostRecentTime) {
            mostRecent = logEntry
            mostRecentTime = finishedTime
          }
        }
      }
      
      if (mostRecent) {
        updateFromLogEntry(mostRecent)
      }
    },
    { deep: true }
  )
})

onBeforeUnmount(() => {
  if (unwatchButtonResults) {
    unwatchButtonResults()
  }
})
</script>

<style scoped>
.mre-container {
  display: grid;
  grid-column: span 2;
}

.mre-link {
  text-decoration: none;
  color: inherit;
  display: grid;
  cursor: pointer;
  grid-column: span 2;
}

.mre-link:hover .mre-output {
  border-color: #999;
}

.mre-output {
  box-shadow: 0 0 .6em #aaa;
  border: 1px dashed #ccc;
  border-radius: .7em;
  padding: 1em;
  margin: 0;
  min-height: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: monospace;
  font-size: 0.9em;
  overflow-x: auto;
  overflow-y: auto;
  transition: border-color 0.2s ease;
  max-height: 20em;
}

@media (prefers-color-scheme: dark) {
  .mre-output {
    border: 1px dashed #444;
    box-shadow: 0 0 .6em #444;
  }
  
  .mre-link:hover .mre-output {
    border-color: #666;
  }
}
</style>

