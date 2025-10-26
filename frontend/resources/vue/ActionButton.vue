<template>
	<div :id="`actionButton-${actionId}`" role="none" class="action-button">
		<button :id="`actionButtonInner-${actionId}`" :title="title" :disabled="!canExec || isDisabled"
													  :class="buttonClasses" @click="handleClick">

			<div class="navigate-on-start-container">
				<div v-if="navigateOnStart == 'pop'" class="navigate-on-start" title="Opens a popup dialog on start">
					<HugeiconsIcon :icon="ComputerTerminal01Icon" />
				</div>
				<div v-if="navigateOnStart == 'arg'" class="navigate-on-start" title="Opens an argument form on start">
					<HugeiconsIcon :icon="TypeCursorIcon" />
				</div>
				<div v-if="navigateOnStart == ''" class="navigate-on-start" title="Run in the background">
					<HugeiconsIcon :icon="WorkoutRunIcon" />
				</div>
			</div>

			<span class="icon" v-html="unicodeIcon"></span>
			<span class="title" aria-live="polite">{{ displayTitle }}
			</span>
		</button>
	</div>
</template>

<script setup>
import ArgumentForm from './views/ArgumentForm.vue'
import { buttonResults } from './stores/buttonResults'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { WorkoutRunIcon, TypeCursorIcon, ComputerTerminal01Icon } from '@hugeicons/core-free-icons'

import { ref, watch, onMounted, inject } from 'vue'

const router = useRouter()
const navigateOnStart = ref('')

const props = defineProps({
  actionData: {
	type: Object,
	required: true
  }
})

const actionId = ref('')
const title = ref('')
const canExec = ref(true)
const popupOnStart = ref('')

// Display properties
const unicodeIcon = ref('&#x1f4a9;')
const displayTitle = ref('')

// State
const isDisabled = ref(false)
const showArgumentForm = ref(false)

// Animation classes
const buttonClasses = ref([])

// Timestamps
const updateIterationTimestamp = ref(0)

function getUnicodeIcon(icon) {
  if (icon === '') {
	console.log('icon not found	', icon)

	return '&#x1f4a9;'
  } else {
	return unescape(icon)
  }
}

function constructFromJson(json) {
  updateIterationTimestamp.value = 0

  updateFromJson(json)

  actionId.value = json.bindingId
  title.value = json.title
  canExec.value = json.canExec
  popupOnStart.value = json.popupOnStart

  if (popupOnStart.value.includes('execution-dialog')) {
	navigateOnStart.value = 'pop'
  } else if (props.actionData.arguments.length > 0) {
	navigateOnStart.value = 'arg'
  }

  isDisabled.value = !json.canExec
  displayTitle.value = title.value
  unicodeIcon.value = getUnicodeIcon(json.icon)
}

function updateFromJson(json) {
  // Fields that should not be updated
  // title - as the callback URL relies on it

  unicodeIcon.value = getUnicodeIcon(json.icon)
}

async function handleClick() {
  if (props.actionData.arguments && props.actionData.arguments.length > 0) {
	router.push(`/actionBinding/${props.actionData.bindingId}/argumentForm`)
  } else {
	await startAction()
  }
}

function getUniqueId() {
  if (window.isSecureContext) {
	return window.crypto.randomUUID()
  } else {
	return Date.now().toString()
  }
}

async function startAction(actionArgs) {
  buttonClasses.value = [] // Removes old animation classes

  if (actionArgs === undefined) {
	actionArgs = []
  }

  // UUIDs are create client side, so that we can setup a "execution-button"
  // to track the execution before we send the request to the server.
  const startActionArgs = {
	bindingId: props.actionData.bindingId,
	arguments: actionArgs,
	uniqueTrackingId: getUniqueId()
  }

  console.log('Watching buttonResults for', startActionArgs.uniqueTrackingId)

  watch(
	() => buttonResults[startActionArgs.uniqueTrackingId],
	(newResult, oldResult) => {
	  onLogEntryChanged(newResult)
	}
  )

  try {
	await window.client.startAction(startActionArgs)
  } catch (err) {
	console.error('Failed to start action:', err)
  }
}

function onLogEntryChanged(logEntry) {
  if (logEntry.executionFinished) {
	onExecutionFinished(logEntry)
  } else {
	onExecutionStarted(logEntry)
  }
}

function onExecutionStarted(logEntry) {
  if (popupOnStart.value && popupOnStart.value.includes('execution-dialog')) {
	router.push(`/logs/${logEntry.executionTrackingId}`)
  }

  isDisabled.value = true
}

function onExecutionFinished(logEntry) {
  if (logEntry.timedOut) {
	renderExecutionResult('action-timeout', 'Timed out')
  } else if (logEntry.blocked) {
	renderExecutionResult('action-blocked', 'Blocked!')
  } else if (logEntry.exitCode !== 0) {
	renderExecutionResult('action-nonzero-exit', 'Exit code ' + logEntry.exitCode)
  } else {
	const ellapsed = Math.ceil(new Date(logEntry.datetimeFinished) - new Date(logEntry.datetimeStarted)) / 1000
	renderExecutionResult('action-success', 'Success!')
  }
}

function renderExecutionResult(resultCssClass, temporaryStatusMessage) {
  updateDom(resultCssClass, '[' + temporaryStatusMessage + ']')
  onExecStatusChanged()
}

function updateDom(resultCssClass, newTitle) {
  if (resultCssClass == null) {
	buttonClasses.value = []
  } else {
	buttonClasses.value = [resultCssClass]
  }

  displayTitle.value = newTitle
}

function onExecStatusChanged() {
  isDisabled.value = false

  setTimeout(() => {
	updateDom(null, title.value)
  }, 2000)
}

onMounted(() => {
  constructFromJson(props.actionData)
})

watch(
  () => props.actionData,
  (newData) => {
	updateFromJson(newData)
  },
  { deep: true }
)

</script>

<style scoped>
.action-button {
	display: flex;
	flex-direction: column;
	flex-grow: 1;
}

.action-button button {
	display: flex;
	flex-direction: column;
	flex-grow: 1;
	justify-content: center;
	padding: 0.5em;
	border: 1px solid #ccc;
	border-radius: 4px;
	background: #fff;
	cursor: pointer;
	transition: all 0.2s ease;
	box-shadow: 0 0 .6em #aaa;
	font-size: .85em;
	border-radius: .7em;
}

.action-button button:hover:not(:disabled) {
	background: #f5f5f5;
	border-color: #999;
}

.action-button button:disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.action-button button .icon {
	font-size: 3em;
	flex-grow: 1;
	align-content: center;
}

.action-button button .title {
	font-weight: 500;

	padding: 0.2em;
}

/* Animation classes */
.action-button button.action-timeout {
	background: #fff3cd;
	border-color: #ffeaa7;
	color: #856404;
}

.action-button button.action-blocked {
	background: #f8d7da !important;
	border-color: #f5c6cb;
	color: #721c24;
}

.action-button button.action-nonzero-exit {
	background: #f8d7da !important;
	border-color: #f5c6cb;
	color: #721c24;
}

.action-button button.action-success {
	background: #d4edda !important;
	border-color: #c3e6cb;
	color: #155724;
}

.action-button-footer {
	margin-top: 0.5em;
}

.navigate-on-start-container {
	position: relative;
	margin-left: auto;
	height: 0;
	right: 0;
	top: 0;
}

@media (prefers-color-scheme: dark) {
	.action-button button {
		background: #111;
		border-color: #000;
		box-shadow: 0 0 6px #000;
		color: #fff;
	}

	.action-button button:hover:not(:disabled) {
		background: #222;
		border-color: #000;
		box-shadow: 0 0 6px #444;
		color: #fff;
	}
}

</style>
