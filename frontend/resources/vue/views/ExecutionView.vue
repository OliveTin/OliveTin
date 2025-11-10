<template>
  <Section :title="'Execution Results: ' + title" id = "execution-results-popup">
    <template #toolbar>
			<router-link v-if="actionId" :to="`/action/${actionId}`" title="View all executions for this action" class="button neutral">
				<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
					<path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm.31-8.86c-1.77-.45-2.34-.94-2.34-1.67 0-.84.79-1.43 2.1-1.43 1.38 0 1.9.66 1.94 1.64h1.71c-.05-1.34-.87-2.57-2.49-2.97V5H10.9v1.69c-1.51.32-2.72 1.3-2.72 2.81 0 1.79 1.49 2.69 3.66 3.21 1.95.46 2.34 1.22 2.34 1.8 0 .53-.39 1.39-2.1 1.39-1.6 0-2.05-.56-2.13-1.45H8.04c.08 1.5 1.18 2.37 2.82 2.69V19h2.34v-1.63c1.65-.35 2.48-1.24 2.48-2.77-.01-1.88-1.51-2.87-3.7-3.23z"/>
				</svg>
				Action Details
			</router-link>
			<button @click="toggleSize" title="Toggle dialog size" class = "neutral">
				<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
					<path fill="currentColor"
						  d="M3 3h6v2H6.462l4.843 4.843l-1.415 1.414L5 6.367V9H3zm0 18h6v-2H6.376l4.929-4.928l-1.415-1.414L5 17.548V15H3zm12 0h6v-6h-2v2.524l-4.867-4.866l-1.414 1.414L17.647 19H15zm6-18h-6v2h2.562l-4.843 4.843l1.414 1.414L19 6.39V9h2z" />
				</svg>
			</button>
    </template>

		<div v-if="logEntry" class = "flex-row">
				<dl class = "fg1">
					<dt>Duration</dt>
					<dd><span v-html="duration"></span></dd>

					<dt>Status</dt>
					<dd>
						<ActionStatusDisplay :log-entry="logEntry" id = "execution-dialog-status" />
					</dd>
				</dl>
        <span class="icon" role="img" v-html="icon" style = "align-self: start"></span>
    </div>

		<div v-if="notFound" class="error-message padded-content">
			<h3>Execution Not Found</h3>
			<p>{{ errorMessage }}</p>
			<p>The execution with ID <code>{{ executionTrackingId }}</code> could not be found.</p>
			<router-link to="/logs">View all logs</router-link> or <router-link to="/">return to home</router-link>.
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
					<button :disabled="!canKill" @click="killAction" title="Kill" id = "execution-dialog-kill-action">
						<HugeiconsIcon :icon="Cancel02Icon" />
						Kill
					</button>
				</div>

	</Section>
</template>

<script setup>
	import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import ActionStatusDisplay from '../components/ActionStatusDisplay.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { OutputTerminal } from '../../../js/OutputTerminal.js'
import { HugeiconsIcon } from '@hugeicons/vue'
import { WorkoutRunIcon, Cancel02Icon, ArrowLeftIcon } from '@hugeicons/core-free-icons'
import { useRouter } from 'vue-router'
import { buttonResults } from '../stores/buttonResults'

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
const notFound = ref(false)
const errorMessage = ref('')

let executionTicker = null
let terminal = null

function initializeTerminal() {
  terminal = new OutputTerminal(executionTrackingId.value)
  terminal.open(xtermOutput.value)
  terminal.resize(80, 40)

  window.terminal = terminal
}

function toggleSize() {
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

async function reset() {
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
  notFound.value = false
  errorMessage.value = ''

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

async function rerunAction() {
  if (!logEntry.value || !logEntry.value.actionId) {
    console.error('Cannot rerun: no action ID available')
    return
  }

  try {
    const startActionArgs = {
      "bindingId": logEntry.value.actionId,
      "arguments": []
    }

    const res = await window.client.startAction(startActionArgs)
    router.push(`/logs/${res.executionTrackingId}`)
  } catch (err) {
    console.error('Failed to rerun action:', err)
    window.showBigError('rerun-action', 'rerunning action', err, false)
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
  notFound.value = false
  errorMessage.value = ''

  const executionStatusArgs = {
	executionTrackingId: executionTrackingId.value
  }

  try {
	const logEntryResult = await window.client.executionStatus(executionStatusArgs)

	await renderExecutionResult(logEntryResult)
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
  actionId.value = res.logEntry.actionId

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
  document.addEventListener('fullscreenchange', (e) => {
    setTimeout(() => { // Wait for the DOM to settle
      if (document.fullscreenElement) {
        window.terminal.fit()
      } else {
        window.terminal.resize(80, 40)
        window.terminal.fit()
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
