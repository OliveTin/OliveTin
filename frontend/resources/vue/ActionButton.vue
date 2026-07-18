<template>
  <div
    :id="`actionButton-${bindingId}`"
    role="none"
    class="action-button"
    @contextmenu.prevent="openActionDetails"
  >
    <span
      v-if="showExecutionIndicator"
      class="execution-indicator"
      :class="executionIndicatorClass"
      :title="executionIndicatorTitle"
      aria-hidden="true"
    />
    <button
      :id="`actionButtonInner-${bindingId}`"
      :title="title"
      :disabled="!canExec || isDisabled"
      :class="combinedClasses"
      @click="handleClick"
    >
      <div
        v-if="showNavigateOnStartIcons"
        class="navigate-on-start-container"
      >
        <div
          v-if="navigateOnStart == 'pop'"
          class="navigate-on-start"
          title="Opens a popup dialog on start"
        >
          <HugeiconsIcon :icon="ComputerTerminal01Icon" />
        </div>
        <div
          v-if="navigateOnStart == 'arg'"
          class="navigate-on-start"
          title="Opens an argument form on start"
        >
          <HugeiconsIcon :icon="TypeCursorIcon" />
        </div>
        <div
          v-if="navigateOnStart == 'hist'"
          class="navigate-on-start"
          title="Opens action execution history on start"
        >
          <HugeiconsIcon :icon="WorkHistoryIcon" />
        </div>
        <div
          v-if="navigateOnStart == ''"
          class="navigate-on-start"
          title="Run in the background"
        >
          <HugeiconsIcon :icon="WorkoutRunIcon" />
        </div>
      </div>

      <ActionIconGlyph
        class="icon"
        :glyph="actionGlyph"
      />
      <span
        class="title"
        aria-live="polite"
      >{{ displayTitle }}
      </span>
      <span
        v-if="rateLimitMessage"
        class="rate-limit-message"
      >{{ rateLimitMessage }}</span>
    </button>
  </div>
</template>

<script setup>
import { buttonResults } from './stores/buttonResults'
import { rateLimits } from './stores/rateLimits'
import {
  bindingExecutionState,
  pendingBindingFlash,
  consumePendingBindingFlash,
  setBindingExecutionState
} from './stores/bindingExecutionState'
import { connectionState } from './stores/connectionState'
import { requestReconnectNow, applyExecutionLogEntry } from '../../js/websocket.js'
import { useRouter } from 'vue-router'
import { needsArgumentForm } from './utils/needsArgumentForm.js'
import { shouldSuppressPopupOnStartNavigation } from './utils/popupOnStartNavigation.js'
import { HugeiconsIcon } from '@hugeicons/vue'
import { WorkoutRunIcon, TypeCursorIcon, ComputerTerminal01Icon, WorkHistoryIcon } from '@hugeicons/core-free-icons'

import ActionIconGlyph from './components/ActionIconGlyph.vue'

import { ref, watch, onMounted, onUnmounted, computed } from 'vue'

const router = useRouter()
const navigateOnStart = ref('')

const props = defineProps({
  actionData: {
    type: Object,
    required: true
  },
  cssClass: {
    type: String,
    required: false,
    default: ''
  },
  prefilledArguments: {
    type: Object,
    required: false,
    default: () => ({})
  }
})

const bindingId = ref('')
const title = ref('')
const canExec = ref(true)
const popupOnStart = ref('')

// Display properties
const displayTitle = ref('')

// State
const isDisabled = ref(false)

// Rate limiting
const rateLimitExpires = ref(0)
const isRateLimited = ref(false)
const rateLimitMessage = ref('')
const rateLimitInterval = ref(null)
const isComponentMounted = ref(true)

// Animation classes
const buttonClasses = ref([])
const flashedTrackingIds = new Set()

// Show navigate on start icons - defaults to true if not set
const showNavigateOnStartIcons = computed(() => {
  return window.initResponse?.showNavigateOnStartIcons ?? true
})

const actionGlyph = computed(() => props.actionData?.icon ?? '')
const glyph = ref('')

// Combined classes including custom cssClass
const combinedClasses = computed(() => {
  const classes = [...buttonClasses.value]
  if (props.cssClass) {
    classes.push(props.cssClass)
  }
  return classes
})

const hasRunningInstance = computed(() => {
  const id = bindingId.value
  return !!(id && bindingExecutionState[id]?.hasRunning)
})

const hasQueuedInstance = computed(() => {
  const id = bindingId.value
  return !!(id && bindingExecutionState[id]?.hasQueued)
})

const showExecutionIndicator = computed(() => {
  return hasRunningInstance.value || hasQueuedInstance.value
})

const executionIndicatorClass = computed(() => {
  if (hasRunningInstance.value) {
    return 'execution-indicator-running'
  }
  if (hasQueuedInstance.value) {
    return 'execution-indicator-queued'
  }
  return ''
})

const executionIndicatorTitle = computed(() => {
  if (hasRunningInstance.value) {
    return 'Running'
  }
  if (hasQueuedInstance.value) {
    return 'Queued'
  }
  return ''
})

function consumeAndFlashPendingResult () {
  const id = bindingId.value
  if (!id) {
    return
  }

  const pending = consumePendingBindingFlash(id)
  if (pending) {
    onExecutionFinished(pending)
  }
}

// Timestamps
const updateIterationTimestamp = ref(0)

function constructFromJson (json) {
  updateIterationTimestamp.value = 0

  updateFromJson(json)

  bindingId.value = json.bindingId
  title.value = json.title
  canExec.value = json.canExec
  popupOnStart.value = json.popupOnStart

  if (popupOnStart.value.includes('execution-dialog')) {
    navigateOnStart.value = 'pop'
  } else if (popupOnStart.value === 'history') {
    navigateOnStart.value = 'hist'
  } else if (needsArgumentForm(props.actionData)) {
    navigateOnStart.value = 'arg'
  }

  isDisabled.value = !json.canExec
  displayTitle.value = title.value
  glyph.value = json.icon ?? ''
  // Initialize rate limit from action data (parse datetime string)
  if (json.datetimeRateLimitExpires) {
    const date = new Date(json.datetimeRateLimitExpires.replace(' ', 'T'))
    rateLimitExpires.value = date.getTime() / 1000
  } else {
    rateLimitExpires.value = 0
  }
  // Also initialize the store so the watch picks it up
  if (bindingId.value) {
    rateLimits[bindingId.value] = rateLimitExpires.value
    setBindingExecutionState(
	  bindingId.value,
	  !!json.hasRunningInstance,
	  !!json.hasQueuedInstance
    )
  }
  updateRateLimitStatus()
}

function updateFromJson (json) {
  // Fields that should not be updated
  // title - as the callback URL relies on it

  // Update rate limiting if changed (parse datetime string)
  if (json.datetimeRateLimitExpires) {
    const date = new Date(json.datetimeRateLimitExpires.replace(' ', 'T'))
    rateLimitExpires.value = date.getTime() / 1000
    updateRateLimitStatus()
  } else if (json.datetimeRateLimitExpires === '') {
    // Explicitly clear if empty string
    rateLimitExpires.value = 0
    updateRateLimitStatus()
  }
}

function updateRateLimitStatus () {
  if (rateLimitExpires.value === 0) {
    isRateLimited.value = false
    rateLimitMessage.value = ''
    if (rateLimitInterval.value) {
	  clearInterval(rateLimitInterval.value)
	  rateLimitInterval.value = null
    }
    return
  }

  const now = Math.floor(Date.now() / 1000)
  const expires = rateLimitExpires.value

  if (now >= expires) {
    // Rate limit has expired
    isRateLimited.value = false
    rateLimitMessage.value = ''
    rateLimitExpires.value = 0
    if (rateLimitInterval.value) {
	  clearInterval(rateLimitInterval.value)
	  rateLimitInterval.value = null
    }
  } else {
    // Still rate limited
    isRateLimited.value = true
    const secondsRemaining = expires - now
    rateLimitMessage.value = `Rate limited, available in ${secondsRemaining} second${secondsRemaining !== 1 ? 's' : ''}`

    // Set up interval to update every second
    if (!rateLimitInterval.value) {
	  rateLimitInterval.value = setInterval(() => {
        updateRateLimitStatus()
	  }, 1000)
    }
  }
}

function openActionDetails () {
  const id = props.actionData?.bindingId
  if (!id) {
    return
  }
  router.push(`/action/${id}`)
}

async function handleClick () {
  if (popupOnStart.value === 'history') {
    openActionDetails()
    return
  }
  if (needsArgumentForm(props.actionData)) {
    const bindingId = props.actionData.bindingId
    const prefilled = props.prefilledArguments || {}
    if (Object.keys(prefilled).length > 0) {
	  router.push({
        path: `/actionBinding/${bindingId}/argumentForm`,
        state: { prefilledArguments: prefilled }
	  })
    } else {
	  router.push(`/actionBinding/${bindingId}/argumentForm`)
    }
  } else {
    await startAction()
  }
}

function getUniqueId () {
  if (window.isSecureContext) {
    return window.crypto.randomUUID()
  } else {
    return Date.now().toString()
  }
}

async function pollExecutionUntilDone (trackingId) {
  const pollIntervalMs = 500
  const pollTimeoutMs = 10 * 60 * 1000
  const deadline = Date.now() + pollTimeoutMs

  while (Date.now() < deadline && isComponentMounted.value) {
    try {
      const result = await window.client.executionStatus({ executionTrackingId: trackingId })
      if (!isComponentMounted.value) {
        return
      }
      if (result.logEntry) {
        applyExecutionLogEntry(result.logEntry)
        if (result.logEntry.executionFinished) {
          return
        }
      }
    } catch (err) {
      console.error('Failed to poll execution status:', err)
    }

    if (!isComponentMounted.value) {
      return
    }

    await new Promise(resolve => setTimeout(resolve, pollIntervalMs))
  }
}

async function startAction (actionArgs) {
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

  requestReconnectNow()

  try {
    const response = await window.client.startAction(startActionArgs)
    const trackingId = response.executionTrackingId || startActionArgs.uniqueTrackingId

    if (popupOnStart.value && popupOnStart.value.includes('execution-dialog')) {
	  router.push(`/logs/${trackingId}`)
    }

    if (!connectionState.connected) {
	  await pollExecutionUntilDone(trackingId)
    }
  } catch (err) {
    console.error('Failed to start action:', err)
  }
}

function onLogEntryChanged (logEntry) {
  if (logEntry.executionFinished) {
    onExecutionFinished(logEntry)
  } else if (logEntry.queued && !logEntry.executionStarted) {
    onExecutionQueued(logEntry)
  } else {
    onExecutionStarted(logEntry)
  }
}

function onExecutionQueued (_logEntry) {
  isDisabled.value = true
  updateDom('action-queued', '[Queued]')
}

function onExecutionStarted (logEntry) {
  if (
    popupOnStart.value &&
	popupOnStart.value.includes('execution-dialog') &&
	!shouldSuppressPopupOnStartNavigation(router)
  ) {
    router.push(`/logs/${logEntry.executionTrackingId}`)
  }

  isDisabled.value = true
  updateDom(null, title.value)
}

function onExecutionFinished (logEntry) {
  const trackingId = logEntry.executionTrackingId
  if (trackingId) {
    if (flashedTrackingIds.has(trackingId)) {
      return
    }
    flashedTrackingIds.add(trackingId)
  }

  // Local no-arg watches and the binding-scoped pending flash can both
  // observe the same finished execution; consume so only one path flashes.
  if (bindingId.value) {
    consumePendingBindingFlash(bindingId.value)
  }

  if (logEntry.timedOut) {
    renderExecutionResult('action-timeout', 'Timed out')
  } else if (logEntry.blocked) {
    renderExecutionResult('action-blocked', 'Blocked!')
  } else if (logEntry.exitCode !== 0) {
    renderExecutionResult('action-nonzero-exit', 'Exit code ' + logEntry.exitCode)
  } else {
    renderExecutionResult('action-success', 'Success!')
  }
}

function renderExecutionResult (resultCssClass, temporaryStatusMessage) {
  updateDom(resultCssClass, '[' + temporaryStatusMessage + ']')
  onExecStatusChanged()
}

function updateDom (resultCssClass, newTitle) {
  if (resultCssClass == null) {
    buttonClasses.value = []
  } else {
    buttonClasses.value = [resultCssClass]
  }

  displayTitle.value = newTitle
}

function onExecStatusChanged () {
  isDisabled.value = false

  setTimeout(() => {
    updateDom(null, title.value)
  }, 2000)
}

onMounted(() => {
  constructFromJson(props.actionData)

  // Watch the central rate limit store for updates to this button's bindingId
  // Watch the entire rateLimits object to ensure reactivity with dynamic keys
  watch(
    rateLimits,
    () => {
	  const id = bindingId.value
	  if (id && rateLimits[id] !== undefined) {
        const newExpires = rateLimits[id]
        if (newExpires !== rateLimitExpires.value) {
		  rateLimitExpires.value = newExpires
		  updateRateLimitStatus()
        }
	  }
    },
    { deep: true }
  )

  // Binding-scoped flash survives argument-form navigation/remount (#920).
  watch(
    () => pendingBindingFlash[bindingId.value],
    (pending) => {
	  if (pending) {
        consumeAndFlashPendingResult()
	  }
    },
    { immediate: true }
  )
})

onUnmounted(() => {
  isComponentMounted.value = false
  if (rateLimitInterval.value) {
    clearInterval(rateLimitInterval.value)
    rateLimitInterval.value = null
  }
})

watch(
  () => props.actionData,
  (newData) => {
    updateFromJson(newData)
    if (newData?.icon !== undefined) {
	  glyph.value = newData.icon ?? ''
    }
  },
  { deep: true }
)

defineExpose({
  glyph
})

</script>

<style>

@layer components {
	.action-button {
		display: flex;
		flex-direction: column;
		flex-grow: 1;
		position: relative;
	}

	.execution-indicator {
		position: absolute;
		top: 0.45em;
		left: 0.45em;
		width: 0.65em;
		height: 0.65em;
		border-radius: 50%;
		z-index: 1;
		pointer-events: none;
	}

	.execution-indicator-running {
		background: #28a745;
	}

	.execution-indicator-queued {
		background: #0d6efd;
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

	.action-button button .rate-limit-message {
		font-size: 0.75em;
		color: #856404;
		padding: 0.2em;
		font-weight: normal;
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

	.action-button button.action-queued {
		background: #e7f1ff !important;
		border-color: #9ec5fe;
		color: #084298;
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
}
</style>
