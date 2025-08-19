<template>
  <section class="with-header-and-content">
    <div class="section-header">
      <h2>Execution Results: {{ title }}</h2>

      <button @click="toggleSize" title="Toggle dialog size">
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor"
            d="M3 3h6v2H6.462l4.843 4.843l-1.415 1.414L5 6.367V9H3zm0 18h6v-2H6.376l4.929-4.928l-1.415-1.414L5 17.548V15H3zm12 0h6v-6h-2v2.524l-4.867-4.866l-1.414 1.414L17.647 19H15zm6-18h-6v2h2.562l-4.843 4.843l1.414 1.414L19 6.39V9h2z" />
        </svg>
      </button>

    </div>
    <div class="section-content">
      <div v-if="logEntry">
        <div class="action-header padded-content" style="float: right">
          <span class="icon" role="img" v-html="icon"></span>
        </div>

        <dl>
          <dt>Duration</dt>
          <dd><span v-html="duration"></span></dd>

          <dt>Status</dt>
          <dd>
            <ActionStatusDisplay :log-entry="logEntry" />
          </dd>
        </dl>
      </div>

      <div ref="xtermOutput"></div>

      <br />

      <div class="flex-row g1 buttons padded-content">
        <button @click="goBack" title="Go back">
          <HugeiconsIcon :icon="ArrowLeftIcon" />
          Back
        </button>

        <div class = "fg1" />

        <button :disabled="!canRerun" @click="rerunAction" title="Rerun">
          <HugeiconsIcon :icon="WorkoutRunIcon" />
          Rerun
        </button>
        <button :disabled="!canKill" @click="killAction" title="Kill">
          <HugeiconsIcon :icon="Cancel02Icon" />
          Kill
        </button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick } from 'vue'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'
import { OutputTerminal } from '../../../js/OutputTerminal.js'
import { HugeiconsIcon } from '@hugeicons/vue'
import { WorkoutRunIcon, Cancel02Icon, ArrowLeftIcon } from '@hugeicons/core-free-icons'
import { useRouter } from 'vue-router'

const router = useRouter()

// Refs for DOM elements
const xtermOutput = ref(null)
const dialog = ref(null)

const props = defineProps({
  executionTrackingId: {
    type: String,
    required: true
  }
})

const executionTrackingId = ref(props.executionTrackingId)
const isBig = ref(false)
const hideBasics = ref(false)
const hideDetails = ref(false)
const hideDetailsOnResult = ref(false)
const executionSeconds = ref(0)
const icon = ref('')
const title = ref('Waiting for result...')
const titleTooltip = ref('')
const duration = ref('')
const logEntry = ref(null)
const canRerun = ref(false)
const canKill = ref(false)

let executionTicker = null
let terminal = null

function initializeTerminal() {
  terminal = new OutputTerminal()

  console.log('initializeTerminal', xtermOutput.value)

  terminal.open(xtermOutput.value)
  terminal.resize(80, 24)
  window.terminal = terminal
}

function toggleSize() {
  isBig.value = !isBig.value
  if (isBig.value) {
    terminal.fit()
  } else {
    terminal.resize(80, 24)
  }
}

async function reset() {
  executionSeconds.value = 0
  executionTrackingId.value = 'notset'
  isBig.value = false
  hideBasics.value = false
  hideDetails.value = false
  hideDetailsOnResult.value = false

  icon.value = ''
  title.value = 'Waiting for result...'
  titleTooltip.value = ''
  duration.value = ''

  canRerun.value = false
  canKill.value = false
  logEntry.value = null

  if (terminal) {
    await terminal.reset()
    terminal.fit()
  }
}

function show(actionButton) {
  if (actionButton) {
    icon.value = actionButton.domIcon.innerText
  }

  canKill.value = true

  // Clear existing ticker
  if (executionTicker) {
    clearInterval(executionTicker)
  }

  executionSeconds.value = 0
  executionTick()
  executionTicker = setInterval(() => {
    executionTick()
  }, 1000)
}

function rerunAction() {
  if (logEntry.value && logEntry.value.actionId) {
    const actionButton = document.getElementById('actionButton-' + logEntry.value.actionId)
    if (actionButton && actionButton.btn) {
      actionButton.btn.click()
    }
  }
}

async function killAction() {
  if (!executionTrackingId.value || executionTrackingId.value === 'notset') {
    return
  }

  const killActionArgs = {
    executionTrackingId: executionTrackingId.value
  }

  try {
    await window.client.killAction(killActionArgs)
  } catch (err) {
    console.error('Failed to kill action:', err)
  }
}

function executionTick() {
  executionSeconds.value++
  updateDuration(null)
}

function hideEverythingApartFromOutput() {
  hideDetailsOnResult.value = true
  hideBasics.value = true
  hideDetailsOnResult.value = true
  hideBasics.value = true
}

async function fetchExecutionResult(executionTrackingIdParam) {
  console.log("fetchExecutionResult", executionTrackingIdParam)

  executionTrackingId.value = executionTrackingIdParam

  const executionStatusArgs = {
    executionTrackingId: executionTrackingId.value
  }

  try {
    const logEntryResult = await window.client.executionStatus(executionStatusArgs)

    await renderExecutionResult(logEntryResult)
  } catch (err) {
    renderError(err)
    throw err
  }
}

function updateDuration(logEntryParam) {
  logEntry.value = logEntryParam
  if (logEntry.value == null) {
    duration.value = executionSeconds.value + ' seconds'
    duration.value = duration.value
  } else if (!logEntry.value.executionStarted) {
    duration.value = logEntry.value.datetimeStarted + ' (request time). Not executed.'
  } else if (logEntry.value.executionStarted && !logEntry.value.executionFinished) {
    duration.value = logEntry.value.datetimeStarted
  } else {
    let delta = ''
    try {
      delta = (new Date(logEntry.value.datetimeStarted) - new Date(logEntry.value.datetimeStarted)) / 1000
      delta = new Intl.RelativeTimeFormat().format(delta, 'seconds').replace('in ', '').replace('ago', '')
    } catch (e) {
      console.warn('Failed to calculate delta', e)
    }
    duration.value = logEntry.value.datetimeStarted + ' &rarr; ' + logEntry.value.datetimeFinished
    if (delta !== '') {
      duration.value += ' (' + delta + ')'
    }
  }
}

async function renderExecutionResult(res) {
  logEntry.value = res.logEntry

  // Clear ticker
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  executionTicker = null

  if (hideDetailsOnResult.value) {
    hideDetails.value = true
  }

  executionTrackingId.value = res.logEntry.executionTrackingId
  canRerun.value = res.logEntry.executionFinished
  canKill.value = res.logEntry.canKill

  icon.value = res.logEntry.actionIcon
  title.value = res.logEntry.actionTitle
  titleTooltip.value = 'Action ID: ' + res.logEntry.actionId + '\nExecution ID: ' + res.logEntry.executionTrackingId

  updateDuration(res.logEntry)

  if (terminal) {
    await terminal.reset()
    await terminal.write(res.logEntry.output, () => {
      terminal.fit()
    })
  }
}

function renderError(err) {
  window.showBigError('execution-dlg-err', 'in the execution dialog', 'Failed to fetch execution result. ' + err, false)
}

function handleClose() {
  if (executionTicker) {
    clearInterval(executionTicker)
  }

  executionTicker = null
}

function cleanup() {
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  executionTicker = null
  if (terminal != null) {
    terminal.close()
  }
  terminal = null
}

function goBack() {
  router.back()
}

onMounted(() => {
  initializeTerminal()
  fetchExecutionResult(props.executionTrackingId)
})

onBeforeUnmount(() => {
  cleanup()
})

// Expose methods for parent/imperative use
defineExpose({
  reset,
  show,
  rerunAction,
  killAction,
  fetchExecutionResult,
  renderExecutionResult,
  hideEverythingApartFromOutput,
  handleClose
})

</script>
