<template>
  <div :id="`actionButton-${actionId}`" role="none" class="action-button">
    <button :id="`actionButtonInner-${actionId}`" :title="title" :disabled="!canExec || isDisabled"
      :class="buttonClasses" @click="handleClick">
      <span class="icon" v-html="unicodeIcon"></span>
      <span class="title" aria-live="polite">{{ displayTitle }}</span>
    </button>

    <ArgumentForm v-if="showArgumentForm" :action-data="actionData" @submit="handleArgumentSubmit"
      @cancel="handleArgumentCancel" @close="handleArgumentClose" />
  </div>
</template>

<script setup>
import ArgumentForm from './ArgumentForm.vue'

import { ref, computed, watch, onMounted, inject } from 'vue'

const executionDialog = inject('executionDialog');

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
    return '&#x1f4a9;'
  } else {
    return unescape(icon)
  }
}

function constructFromJson(json) {
  updateIterationTimestamp.value = 0

  // Class attributes
  updateFromJson(json)

  actionId.value = json.id
  title.value = json.title
  canExec.value = json.canExec
  popupOnStart.value = json.popupOnStart

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
    updateUrlWithAction()
    showArgumentForm.value = true
  } else {
    await startAction()
  }
}

function updateUrlWithAction() {
  // Get the current URL and create a new URL object
  const url = new URL(window.location.href)

  // Set the action parameter
  url.searchParams.set('action', title.value)

  // Update the URL without reloading the page
  window.history.replaceState({}, '', url.toString())
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
    actionId: actionId.value,
    arguments: actionArgs,
    uniqueTrackingId: getUniqueId()
  }

  onActionStarted(startActionArgs.uniqueTrackingId)

  try {
    await window.client.startAction(startActionArgs)
  } catch (err) {
    console.error('Failed to start action:', err)
  }
}

function onActionStarted(execTrackingId) {
  console.log('onActionStarted', execTrackingId)
  console.log('executionDialog', executionDialog)

  if (popupOnStart.value && popupOnStart.value.includes('execution-dialog')) {
    if (executionDialog.value) {
      executionDialog.value.reset();

      if (popupOnStart.value === 'execution-dialog-stdout-only') {
        executionDialog.value.hideEverythingApartFromOutput();
      }
    }

    executionDialog.value.executionTrackingId = execTrackingId;
    executionDialog.value.show()
  }

  isDisabled.value = true
}

function handleArgumentSubmit(args) {
  startAction(args)
  showArgumentForm.value = false
}

function handleArgumentCancel() {
  showArgumentForm.value = false
}

function handleArgumentClose() {
  showArgumentForm.value = false
}

// ExecutionFeedbackButton methods
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

defineExpose({
  onExecutionFinished
})
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
  gap: 0.5em;
  padding: 0.5em 1em;
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
}

.action-button button .title {
  font-weight: 500;
}

/* Animation classes */
.action-button button.action-timeout {
  background: #fff3cd;
  border-color: #ffeaa7;
  color: #856404;
}

.action-button button.action-blocked {
  background: #f8d7da;
  border-color: #f5c6cb;
  color: #721c24;
}

.action-button button.action-nonzero-exit {
  background: #f8d7da;
  border-color: #f5c6cb;
  color: #721c24;
}

.action-button button.action-success {
  background: #d4edda;
  border-color: #c3e6cb;
  color: #155724;
}

.action-button-footer {
  margin-top: 0.5em;
}
</style>