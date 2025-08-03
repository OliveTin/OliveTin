<template>
  <dialog 
    ref="dialog" 
    title="Execution Results" 
    :class="{ big: isBig }"
    @close="handleClose"
  >
    <div class="action-header padded-content">
      <span class="icon" role="img" v-html="icon"></span>

      <h2>
        <span :title="titleTooltip">{{ title }}</span>
      </h2>

      <button 
        v-show="!hideToggleButton"
        @click="toggleSize" 
        title="Toggle dialog size"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M3 3h6v2H6.462l4.843 4.843l-1.415 1.414L5 6.367V9H3zm0 18h6v-2H6.376l4.929-4.928l-1.415-1.414L5 17.548V15H3zm12 0h6v-6h-2v2.524l-4.867-4.866l-1.414 1.414L17.647 19H15zm6-18h-6v2h2.562l-4.843 4.843l1.414 1.414L19 6.39V9h2z"/>
        </svg>
      </button>
    </div>
    
    <div v-show="!hideBasics" class="padded-content-sides">
      <strong>Duration: </strong><span v-html="duration"></span>
    </div>
    
    <div v-show="!hideDetails && logEntry" class="padded-content-sides">
      <p>
        <strong>Status: </strong>
        <ActionStatusDisplay :log-entry="logEntry" v-if="logEntry"/>
      </p>
    </div>

    <div ref="xtermOutput" v-show="!isHtmlOutput"></div>
    <div 
      v-show="isHtmlOutput" 
      class="padded-content"
      v-html="htmlOutput"
    ></div>

    <div class="buttons padded-content">
      <button 
        :disabled="!canRerun" 
        @click="rerunAction" 
        title="Rerun"
      >
        Rerun
      </button>
      <button 
        :disabled="!canKill" 
        @click="killAction" 
        title="Kill"
      >
        Kill
      </button>

      <form method="dialog">
        <button name="Cancel" title="Close">Close</button>
      </form>
    </div>
  </dialog>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick } from 'vue'
import ActionStatusDisplay from './components/ActionStatusDisplay.vue'
import { OutputTerminal } from '../../js/OutputTerminal.js'

// Refs for DOM elements
const xtermOutput = ref(null)
const dialog = ref(null)

// State
const state = reactive({
  isBig: false,
  hideToggleButton: false,
  hideBasics: false,
  hideDetails: false,
  hideDetailsOnResult: false,

  // Execution data
  executionSeconds: 0,
  executionTrackingId: 'notset',
  executionTicker: null,

  // Display data
  icon: '',
  title: 'Waiting for result...',
  titleTooltip: '',
  duration: '',
  htmlOutput: '',
  isHtmlOutput: false,

  // Action data
  logEntry: null,
  canRerun: false,
  canKill: false,

  // Terminal
  terminal: null
})

// Expose for template
const isBig = ref(false)
const hideToggleButton = ref(false)
const hideBasics = ref(false)
const hideDetails = ref(false)
const hideDetailsOnResult = ref(false)
const executionSeconds = ref(0)
const executionTrackingId = ref('notset')
const icon = ref('')
const title = ref('Waiting for result...')
const titleTooltip = ref('')
const duration = ref('')
const htmlOutput = ref('')
const isHtmlOutput = ref(false)
const logEntry = ref(null)
const canRerun = ref(false)
const canKill = ref(false)

let executionTicker = null
let terminal = null

function syncStateToRefs() {
  isBig.value = state.isBig
  hideToggleButton.value = state.hideToggleButton
  hideBasics.value = state.hideBasics
  hideDetails.value = state.hideDetails
  hideDetailsOnResult.value = state.hideDetailsOnResult
  executionSeconds.value = state.executionSeconds
  executionTrackingId.value = state.executionTrackingId
  icon.value = state.icon
  title.value = state.title
  titleTooltip.value = state.titleTooltip
  duration.value = state.duration
  htmlOutput.value = state.htmlOutput
  isHtmlOutput.value = state.isHtmlOutput
  logEntry.value = state.logEntry
  canRerun.value = state.canRerun
  canKill.value = state.canKill
}

function syncRefsToState() {
  state.isBig = isBig.value
  state.hideToggleButton = hideToggleButton.value
  state.hideBasics = hideBasics.value
  state.hideDetails = hideDetails.value
  state.hideDetailsOnResult = hideDetailsOnResult.value
  state.executionSeconds = executionSeconds.value
  state.executionTrackingId = executionTrackingId.value
  state.icon = icon.value
  state.title = title.value
  state.titleTooltip = titleTooltip.value
  state.duration = duration.value
  state.htmlOutput = htmlOutput.value
  state.isHtmlOutput = isHtmlOutput.value
  state.logEntry = logEntry.value
  state.canRerun = canRerun.value
  state.canKill = canKill.value
}

function initializeTerminal() {
  terminal = new OutputTerminal()
  terminal.open(xtermOutput.value)
  terminal.resize(80, 24)
  window.terminal = terminal
  state.terminal = terminal
}

function toggleSize() {
  state.isBig = !state.isBig
  isBig.value = state.isBig
  if (state.isBig) {
    terminal.fit()
  } else {
    terminal.resize(80, 24)
  }
}

async function reset() {
  state.executionSeconds = 0
  state.executionTrackingId = 'notset'
  state.isBig = false
  state.hideToggleButton = false
  state.hideBasics = false
  state.hideDetails = false
  state.hideDetailsOnResult = false

  state.icon = ''
  state.title = 'Waiting for result...'
  state.titleTooltip = ''
  state.duration = ''
  state.htmlOutput = ''
  state.isHtmlOutput = false

  state.canRerun = false
  state.canKill = false
  state.logEntry = null

  syncStateToRefs()

  if (terminal) {
    await terminal.reset()
    terminal.fit()
  }
}

function show(actionButton) {
  if (actionButton) {
    state.icon = actionButton.domIcon.innerText
    icon.value = state.icon
  }

  state.canKill = true
  canKill.value = true

  // Clear existing ticker
  if (executionTicker) {
    clearInterval(executionTicker)
  }

  state.executionSeconds = 0
  executionSeconds.value = 0
  executionTick()
  executionTicker = setInterval(() => {
    executionTick()
  }, 1000)
  state.executionTicker = executionTicker

  // Close if already open
  if (dialog.value && dialog.value.open) {
    dialog.value.close()
  }

  dialog.value && dialog.value.showModal()
}

function rerunAction() {
  if (state.logEntry && state.logEntry.actionId) {
    const actionButton = document.getElementById('actionButton-' + state.logEntry.actionId)
    if (actionButton && actionButton.btn) {
      actionButton.btn.click()
    }
  }
  dialog.value && dialog.value.close()
}

async function killAction() {
  if (!state.executionTrackingId || state.executionTrackingId === 'notset') {
    return
  }

  const killActionArgs = {
    executionTrackingId: state.executionTrackingId
  }

  try {
    await window.client.killAction(killActionArgs)
  } catch (err) {
    console.error('Failed to kill action:', err)
  }
}

function executionTick() {
  state.executionSeconds++
  executionSeconds.value = state.executionSeconds
  updateDuration(null)
}

function hideEverythingApartFromOutput() {
  state.hideDetailsOnResult = true
  state.hideBasics = true
  hideDetailsOnResult.value = true
  hideBasics.value = true
}

async function fetchExecutionResult(executionTrackingIdParam) {
  state.executionTrackingId = executionTrackingIdParam
  executionTrackingId.value = executionTrackingIdParam

  const executionStatusArgs = {
    executionTrackingId: state.executionTrackingId
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
  let logEntry = logEntryParam
  if (logEntry == null) {
    duration.value = state.executionSeconds + ' seconds'
    state.duration = duration.value
  } else if (!logEntry.executionStarted) {
    duration.value = logEntry.datetimeStarted + ' (request time). Not executed.'
    state.duration = duration.value
  } else if (logEntry.executionStarted && !logEntry.executionFinished) {
    duration.value = logEntry.datetimeStarted
    state.duration = duration.value
  } else {
    let delta = ''
    try {
      delta = (new Date(logEntry.datetimeStarted) - new Date(logEntry.datetimeStarted)) / 1000
      delta = new Intl.RelativeTimeFormat().format(delta, 'seconds').replace('in ', '').replace('ago', '')
    } catch (e) {
      console.warn('Failed to calculate delta', e)
    }
    duration.value = logEntry.datetimeStarted + ' &rarr; ' + logEntry.datetimeFinished
    if (delta !== '') {
      duration.value += ' (' + delta + ')'
    }
    state.duration = duration.value
  }
}

async function renderExecutionResult(res) {
  // Clear ticker
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  state.executionTicker = null

  if (res.type === 'execution-dialog-output-html') {
    state.isHtmlOutput = true
    state.htmlOutput = res.logEntry.output
    state.hideDetailsOnResult = true
    isHtmlOutput.value = true
    htmlOutput.value = res.logEntry.output
    hideDetailsOnResult.value = true
  } else {
    state.isHtmlOutput = false
    state.htmlOutput = ''
    isHtmlOutput.value = false
    htmlOutput.value = ''
  }

  if (state.hideDetailsOnResult) {
    state.hideDetails = true
    hideDetails.value = true
  }

  state.executionTrackingId = res.logEntry.executionTrackingId
  executionTrackingId.value = res.logEntry.executionTrackingId
  state.canRerun = res.logEntry.executionFinished
  canRerun.value = res.logEntry.executionFinished
  state.canKill = res.logEntry.canKill
  canKill.value = res.logEntry.canKill
  state.logEntry = res.logEntry
  logEntry.value = res.logEntry

  state.icon = res.logEntry.actionIcon
  icon.value = res.logEntry.actionIcon
  state.title = res.logEntry.actionTitle
  title.value = res.logEntry.actionTitle
  state.titleTooltip = 'Action ID: ' + res.logEntry.actionId + '\nExecution ID: ' + res.logEntry.executionTrackingId
  titleTooltip.value = state.titleTooltip

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
  // Clean up when dialog is closed
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  state.executionTicker = null
}

function cleanup() {
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  state.executionTicker = null
  if (terminal) {
    terminal.close()
  }
  state.terminal = null
}

onMounted(() => {
  nextTick(() => {
    initializeTerminal()
  })
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