<template>
  <Section id="execution-results-popup">
    <template #title>
      <span class="section-title-with-icon">
        Execution Results:
        <router-link
          v-if="actionId"
          :to="`/action/${actionId}`"
          class="action-details-title-link"
          :title="titleTooltip"
        >
          <ActionIconGlyph
            class="action-title-icon"
            :glyph="icon"
          />
          <LogActionTitle
            v-if="logEntry"
            :action-title="title"
            :justification="logEntry.justification"
          />
          <span v-else>{{ title }}</span>
        </router-link>
        <template v-else>
          <LogActionTitle
            v-if="logEntry"
            :action-title="title"
            :justification="logEntry.justification"
          />
          <span v-else>{{ title }}</span>
        </template>
      </span>
    </template>
    <template #toolbar>
      <button
        v-for="dashboard in backToDashboards"
        :key="dashboard.path"
        :title="'Back to ' + dashboard.title"
        class="button neutral"
        @click="goToDashboard(dashboard.path)"
      >
        <HugeiconsIcon :icon="DashboardSquare01Icon" />
        {{ dashboard.title }}
      </button>
      <button
        v-if="backToDashboards.length === 0"
        title="Go back"
        class="button neutral"
        @click="goBack"
      >
        <HugeiconsIcon :icon="ArrowLeftIcon" />
        Back
      </button>
    </template>

    <div
      v-if="logEntry"
      class="flex-row"
    >
      <dl class="fg1">
        <dt>Duration</dt>
        <dd><span v-html="duration" /></dd>

        <dt>Status</dt>
        <dd class="execution-dialog-status">
          <ActionStatusDisplay
            :log-entry="logEntry"
            :link-queued-status="true"
          />
        </dd>
      </dl>
    </div>

    <div
      v-if="notFound"
      class="error-message padded-content"
    >
      <h3>Execution Not Found</h3>
      <p>{{ errorMessage }}</p>
      <p>The execution with ID <code>{{ executionTrackingId }}</code> could not be found.</p>
      <router-link to="/logs">
        View all logs
      </router-link> or <router-link to="/">
        return to home
      </router-link>.
    </div>

    <div class="xterm-output-container">
      <div class="xterm-overlay-toolbar">
        <button
          type="button"
          class="xterm-overlay-button"
          title="Copy to clipboard"
          @click="copyOutput"
        >
          <HugeiconsIcon :icon="Copy01Icon" />
        </button>
        <button
          type="button"
          class="xterm-overlay-button"
          title="Toggle fullscreen"
          @click="toggleSize"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="1em"
            height="1em"
            viewBox="0 0 24 24"
          >
            <path
              fill="currentColor"
              d="M3 3h6v2H6.462l4.843 4.843l-1.415 1.414L5 6.367V9H3zm0 18h6v-2H6.376l4.929-4.928l-1.415-1.414L5 17.548V15H3zm12 0h6v-6h-2v2.524l-4.867-4.866l-1.414 1.414L17.647 19H15zm6-18h-6v2h2.562l-4.843 4.843l1.414 1.414L19 6.39V9h2z"
            />
          </svg>
        </button>
      </div>
      <div ref="xtermOutput" />
    </div>

    <br>

    <div class="flex-row g1 buttons padded-content">
      <div class="fg1" />

      <button
        :disabled="!canRerun"
        title="Rerun"
        @click="rerunAction"
      >
        <HugeiconsIcon :icon="WorkoutRunIcon" />
        Rerun
      </button>
      <button
        id="execution-dialog-kill-action"
        :disabled="!canKill"
        title="Kill"
        @click="killAction"
      >
        <HugeiconsIcon :icon="Cancel02Icon" />
        Kill
      </button>
    </div>
  </Section>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import ActionIconGlyph from '../components/ActionIconGlyph.vue'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'
import LogActionTitle from '../components/LogActionTitle.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { OutputTerminal } from '../../../js/OutputTerminal.js'
import { HugeiconsIcon } from '@hugeicons/vue'
import { WorkoutRunIcon, Cancel02Icon, ArrowLeftIcon, DashboardSquare01Icon, Copy01Icon } from '@hugeicons/core-free-icons'
import { useRouter } from 'vue-router'
import { buttonResults } from '../stores/buttonResults'
import { requestReconnectNow } from '../../../js/websocket.js'
import {
  buildRerunPrefilledArguments,
  buildRerunStartActionArgs,
  rerunNeedsArgumentForm
} from '../utils/rerunArguments.js'

const router = useRouter()

// Refs for DOM elements
const xtermOutput = ref(null)

const props = defineProps({
  executionTrackingId: {
    type: String,
    required: true
  }
})

const executionTrackingId = ref(props.executionTrackingId)
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
const actionId = ref('')
const backToDashboards = ref([])
const notFound = ref(false)
const errorMessage = ref('')

let executionTicker = null
let terminal = null

function initializeTerminal () {
  terminal = new OutputTerminal(executionTrackingId.value)
  terminal.open(xtermOutput.value)
  terminal.resize(80, 40)

  window.terminal = terminal
}

function toggleSize () {
  if (!xtermOutput.value) {
    return
  }

  if (xtermOutput.value.requestFullscreen) {
    xtermOutput.value.requestFullscreen()
  } else if (xtermOutput.value.webkitRequestFullscreen) {
    xtermOutput.value.webkitRequestFullscreen()
  } else if (xtermOutput.value.mozRequestFullScreen) {
    xtermOutput.value.mozRequestFullScreen()
  } else if (xtermOutput.value.msRequestFullscreen) {
    xtermOutput.value.msRequestFullscreen()
  }
}

async function copyOutput () {
  const text = terminal?.getBufferAsString?.() || logEntry.value?.output || ''
  if (!text) {
    return
  }

  try {
    await navigator.clipboard.writeText(text)
  } catch (err) {
    console.error('Failed to copy execution output:', err)
  }
}

async function reset () {
  executionSeconds.value = 0
  executionTrackingId.value = 'notset'
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
  backToDashboards.value = []
  notFound.value = false
  errorMessage.value = ''

  if (terminal) {
    await terminal.reset()
    terminal.fit()
  }
}

function show (actionButton) {
  if (actionButton) {
    icon.value = actionButton.glyph ?? ''
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

async function rerunAction () {
  const bindingId = logEntry.value?.bindingId
  if (!logEntry.value || !bindingId) {
    console.error('Cannot rerun: no action ID available')
    return
  }

  try {
    const binding = await window.client.getActionBinding({ bindingId })
    if (rerunNeedsArgumentForm(binding.action, logEntry.value)) {
      router.push({
        path: `/actionBinding/${bindingId}/argumentForm`,
        state: { prefilledArguments: buildRerunPrefilledArguments(logEntry.value) }
      })
      return
    }

    requestReconnectNow()
    const startActionArgs = buildRerunStartActionArgs(bindingId, logEntry.value)

    const res = await window.client.startAction(startActionArgs)
    router.push(`/logs/${res.executionTrackingId}`)
  } catch (err) {
    console.error('Failed to rerun action:', err)
    window.showBigError('rerun-action', 'rerunning action', err, false)
  }
}

async function killAction () {
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

function executionTick () {
  executionSeconds.value++
  updateDuration(null)
}

function hideEverythingApartFromOutput () {
  hideDetailsOnResult.value = true
  hideBasics.value = true
  hideDetailsOnResult.value = true
  hideBasics.value = true
}

async function fetchExecutionResult (executionTrackingIdParam) {
  console.log('fetchExecutionResult', executionTrackingIdParam)

  executionTrackingId.value = executionTrackingIdParam
  notFound.value = false
  errorMessage.value = ''
  backToDashboards.value = []

  const executionStatusArgs = {
    executionTrackingId: executionTrackingId.value
  }

  try {
    const executionStatusResult = await window.client.executionStatus(executionStatusArgs)

    await renderExecutionResult(executionStatusResult)
  } catch (err) {
    // Check if it's a "not found" error (404 or similar)
    if (err.status === 404 || err.code === 'NotFound' || err.message?.includes('not found')) {
	  notFound.value = true
	  errorMessage.value = err.message || 'The execution could not be found in the system.'
    } else {
	  renderError(err)
    }
    throw err
  }
}

function updateDuration (logEntryParam) {
  logEntry.value = logEntryParam
  if (logEntry.value == null) {
    duration.value = executionSeconds.value + ' seconds'
  } else if (!logEntry.value.executionStarted) {
    duration.value = logEntry.value.datetimeStarted + ' (request time). Not executed.'
  } else if (logEntry.value.executionStarted && !logEntry.value.executionFinished) {
    duration.value = logEntry.value.datetimeStarted
  } else {
    let delta = ''
    try {
		  delta = (new Date(logEntry.value.datetimeFinished) - new Date(logEntry.value.datetimeStarted)) / 1000
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

async function renderExecutionResult (res) {
  logEntry.value = res.logEntry
  if (res.backToDashboards) {
    backToDashboards.value = res.backToDashboards.slice(0, 3)
  }

  // Clear ticker
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  executionTicker = null

  if (hideDetailsOnResult.value) {
    hideDetails.value = true
  }

  executionTrackingId.value = res.logEntry.executionTrackingId
  canRerun.value = res.logEntry.executionFinished && !!res.logEntry.bindingId
  canKill.value = res.logEntry.canKill

  icon.value = res.logEntry.actionIcon
  title.value = res.logEntry.actionTitle
  titleTooltip.value = 'Action ID: ' + res.logEntry.bindingId + '\nExecution ID: ' + res.logEntry.executionTrackingId
  actionId.value = res.logEntry.bindingId

  updateDuration(res.logEntry)

  if (terminal) {
    await terminal.reset()
    await terminal.write(res.logEntry.output, () => {
	  terminal.fit()
    })
  }
}

function renderError (err) {
  window.showBigError('execution-dlg-err', 'in the execution dialog', 'Failed to fetch execution result. ' + err, false)
}

function handleClose () {
  if (executionTicker) {
    clearInterval(executionTicker)
  }

  executionTicker = null
}

function cleanup () {
  if (executionTicker) {
    clearInterval(executionTicker)
  }
  executionTicker = null
  if (terminal != null) {
    terminal.close()
  }
  terminal = null
}

function goBack () {
  router.back()
}

function goToDashboard (path) {
  router.push(path)
}

onMounted(() => {
  document.addEventListener('fullscreenchange', (e) => {
    setTimeout(() => { // Wait for the DOM to settle
      if (document.fullscreenElement) {
        terminal.fit()
      } else {
        terminal.resize(80, 40)
        terminal.fit()
      }
    }, 100)
  })

  initializeTerminal()
  fetchExecutionResult(props.executionTrackingId)

  watch(
    () => buttonResults[props.executionTrackingId],
    (newResult, oldResult) => {
	  if (newResult) {
        renderExecutionResult({
		  logEntry: newResult
        })
	  }
    }
  )
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

<style scoped>
.section-title-with-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.action-title-icon {
  font-size: 1.5rem;
}

.action-details-title-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: inherit;
  text-decoration: none;
}

.action-details-title-link:hover {
  text-decoration: underline;
}

.xterm-output-container {
  position: relative;
}

.xterm-overlay-toolbar {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  z-index: 2;
  display: flex;
  gap: 0.35rem;
}

.xterm-overlay-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.35rem;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 0.25rem;
  background: rgba(30, 30, 30, 0.85);
  color: #f0f0f0;
  cursor: pointer;
  line-height: 1;
}

.xterm-overlay-button:hover {
  background: rgba(50, 50, 50, 0.95);
  border-color: rgba(255, 255, 255, 0.45);
  color: #fff;
}

.xterm-overlay-button:focus-visible {
  outline: 2px solid rgba(255, 255, 255, 0.6);
  outline-offset: 2px;
}

.action-history-link {
  color: var(--link-color, #007bff);
  text-decoration: none;
  display: inline-block;
  font-size: 0.9rem;
}

.error-message {
  background-color: #f8d7da;
  border: 1px solid #f5c2c7;
  border-radius: 0.25rem;
  padding: 1.5rem;
  margin: 1rem 0;
}

.error-message h3 {
  margin: 0 0 0.5rem 0;
  color: #721c24;
}

.error-message p {
  margin: 0.5rem 0;
  color: #721c24;
}

.error-message code {
  background-color: #f8d7da;
  padding: 0.125rem 0.25rem;
  border-radius: 0.125rem;
  font-family: monospace;
}

.error-message a {
  color: #721c24;
  text-decoration: underline;
  font-weight: 500;
}
</style>
