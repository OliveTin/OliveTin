import { reactive } from 'vue'
import { buttonResults } from './buttonResults.js'

const INDICATOR_SHOW_DELAY_MS = 1000

export const bindingExecutionState = reactive({})

const pendingShowTimers = {}

function cancelPendingShowTimer (bindingId) {
  const timer = pendingShowTimers[bindingId]
  if (timer != null) {
    clearTimeout(timer)
    delete pendingShowTimers[bindingId]
  }
}

function scheduleIndicatorShow (bindingId) {
  cancelPendingShowTimer(bindingId)

  pendingShowTimers[bindingId] = setTimeout(() => {
    delete pendingShowTimers[bindingId]
    recomputeBindingExecutionState(bindingId)
  }, INDICATOR_SHOW_DELAY_MS)
}

export function setBindingExecutionState (bindingId, hasRunning, hasQueued) {
  if (!bindingId) {
    return
  }

  if (!hasRunning && !hasQueued) {
    delete bindingExecutionState[bindingId]
    return
  }

  bindingExecutionState[bindingId] = { hasRunning, hasQueued }
}

export function recomputeBindingExecutionState (bindingId) {
  if (!bindingId) {
    return
  }

  let hasRunning = false
  let hasQueued = false

  for (const trackingId in buttonResults) {
    const entry = buttonResults[trackingId]
    if (!entry || entry.bindingId !== bindingId || entry.executionFinished) {
      continue
    }

    if (entry.executionStarted) {
      hasRunning = true
    } else {
      hasQueued = true
    }
  }

  setBindingExecutionState(bindingId, hasRunning, hasQueued)
}

export function applyExecutionStartedBindingState (logEntry) {
  if (!logEntry?.bindingId || logEntry.executionFinished) {
    return
  }

  scheduleIndicatorShow(logEntry.bindingId)
}

export function applyExecutionFinishedBindingState (logEntry) {
  if (!logEntry?.bindingId) {
    return
  }

  cancelPendingShowTimer(logEntry.bindingId)
  recomputeBindingExecutionState(logEntry.bindingId)
}
