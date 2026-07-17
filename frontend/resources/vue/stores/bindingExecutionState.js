import { reactive } from 'vue'
import { buttonResults } from './buttonResults.js'

const INDICATOR_SHOW_DELAY_MS = 1000
const PENDING_FLASH_TTL_MS = 5000

export const bindingExecutionState = reactive({})
export const pendingBindingFlash = reactive({})

const pendingShowTimers = {}
const pendingFlashExpireTimers = {}

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
  recordPendingBindingFlash(logEntry)
}

function cancelPendingFlashExpireTimer (bindingId) {
  const timer = pendingFlashExpireTimers[bindingId]
  if (timer != null) {
    clearTimeout(timer)
    delete pendingFlashExpireTimers[bindingId]
  }
}

export function recordPendingBindingFlash (logEntry) {
  if (!logEntry?.bindingId || !logEntry.executionFinished) {
    return
  }

  const bindingId = logEntry.bindingId
  cancelPendingFlashExpireTimer(bindingId)

  pendingBindingFlash[bindingId] = {
    executionTrackingId: logEntry.executionTrackingId,
    timedOut: logEntry.timedOut,
    blocked: logEntry.blocked,
    exitCode: logEntry.exitCode,
    datetimeStarted: logEntry.datetimeStarted,
    datetimeFinished: logEntry.datetimeFinished,
    recordedAt: Date.now()
  }

  pendingFlashExpireTimers[bindingId] = setTimeout(() => {
    delete pendingFlashExpireTimers[bindingId]
    const current = pendingBindingFlash[bindingId]
    if (current?.executionTrackingId === logEntry.executionTrackingId) {
      delete pendingBindingFlash[bindingId]
    }
  }, PENDING_FLASH_TTL_MS)
}

export function consumePendingBindingFlash (bindingId) {
  if (!bindingId || pendingBindingFlash[bindingId] === undefined) {
    return null
  }

  const result = pendingBindingFlash[bindingId]
  cancelPendingFlashExpireTimer(bindingId)
  delete pendingBindingFlash[bindingId]

  if (result.recordedAt && (Date.now() - result.recordedAt) > PENDING_FLASH_TTL_MS) {
    return null
  }

  return result
}
