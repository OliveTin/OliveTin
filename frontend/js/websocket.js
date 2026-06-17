import { buttonResults } from '../resources/vue/stores/buttonResults.js'
import { rateLimits } from '../resources/vue/stores/rateLimits.js'
import { connectionState } from '../resources/vue/stores/connectionState.js'
import {
  applyExecutionFinishedBindingState,
  applyExecutionStartedBindingState
} from '../resources/vue/stores/bindingExecutionState.js'
import { cloneLogEntry } from '../resources/vue/utils/executionLogEvents.js'

const RECONNECT_DELAYS_MS = [200, 1000, 2000, 4000, 8000, 16000, 32000]
const BANNER_DELAY_MS = 2000

let reconnectAttempt = 0
let reconnectTimer = null
let listenersInitialized = false
let eventStreamGeneration = 0
let eventStreamAbortController = null

function shouldConnectEventStream () {
  return window.initResponse && !window.initResponse.loginRequired
}

export function stopEventStream () {
  eventStreamGeneration++
  if (eventStreamAbortController != null) {
    eventStreamAbortController.abort()
    eventStreamAbortController = null
  }

  if (reconnectTimer != null) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  reconnectAttempt = 0
  connectionState.connected = false
  connectionState.reconnecting = false
  connectionState.scheduledReconnectDelayMs = 0
  connectionState.nextReconnectAt = null
  connectionState.showDisconnectedBanner = false
  window.websocketAvailable = false
}

export function connectEventStreamIfNeeded () {
  if (!shouldConnectEventStream()) {
    stopEventStream()
    return
  }

  if (connectionState.connected || reconnectTimer != null) {
    return
  }

  reconnectWebsocket()
}

export function initWebsocket () {
  if (!listenersInitialized) {
    window.addEventListener('EventOutputChunk', onOutputChunk)
    window.addEventListener('EventExecutionStarted', onExecutionStarted)
    window.addEventListener('EventExecutionFinished', onExecutionFinished)
    window.addEventListener('pagehide', stopEventStream)
    listenersInitialized = true
  }

  connectEventStreamIfNeeded()
}

window.websocketAvailable = false

export function requestReconnectNow () {
  if (!shouldConnectEventStream()) {
    return
  }

  if (connectionState.connected) {
    return
  }

  if (reconnectTimer != null) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  reconnectAttempt = 0
  scheduleReconnect(RECONNECT_DELAYS_MS[0])
}

function scheduleReconnect (delayMs) {
  if (reconnectTimer != null) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  connectionState.scheduledReconnectDelayMs = delayMs
  connectionState.nextReconnectAt = delayMs > 0 ? Date.now() + delayMs : null
  updateBannerVisibility()
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    reconnectWebsocket()
  }, delayMs)
}

function updateBannerVisibility () {
  if (connectionState.connected) {
    connectionState.showDisconnectedBanner = false
    return
  }

  connectionState.showDisconnectedBanner = connectionState.scheduledReconnectDelayMs >= BANNER_DELAY_MS
}

async function reconnectWebsocket () {
  if (!shouldConnectEventStream()) {
    return
  }

  if (connectionState.connected) {
    return
  }

  const streamGeneration = ++eventStreamGeneration
  if (eventStreamAbortController != null) {
    eventStreamAbortController.abort()
  }
  eventStreamAbortController = new AbortController()

  connectionState.reconnecting = true
  connectionState.connected = false
  if (connectionState.disconnectedAt == null) {
    connectionState.disconnectedAt = Date.now()
  }
  connectionState.nextReconnectAt = null
  connectionState.scheduledReconnectDelayMs = 0

  try {
    window.websocketAvailable = true
    const stream = window.client.eventStream({}, { signal: eventStreamAbortController.signal })
    connectionState.connected = true
    connectionState.reconnecting = false
    connectionState.disconnectedAt = null
    connectionState.nextReconnectAt = null
    connectionState.scheduledReconnectDelayMs = 0
    connectionState.showDisconnectedBanner = false
    for await (const e of stream) {
      if (streamGeneration !== eventStreamGeneration) {
        return
      }
      if (reconnectAttempt !== 0) {
        reconnectAttempt = 0
      }
      handleEvent(e)
    }
  } catch (err) {
    if (streamGeneration !== eventStreamGeneration) {
      return
    }
    console.error('Websocket connection failed: ', err)
  }

  if (streamGeneration !== eventStreamGeneration) {
    return
  }

  window.websocketAvailable = false
  connectionState.connected = false
  connectionState.reconnecting = false
  connectionState.disconnectedAt = connectionState.disconnectedAt ?? Date.now()

  const delay = RECONNECT_DELAYS_MS[Math.min(reconnectAttempt, RECONNECT_DELAYS_MS.length - 1)]
  reconnectAttempt++
  console.log('Reconnecting websocket in ' + delay + 'ms...')

  if (!shouldConnectEventStream()) {
    return
  }

  scheduleReconnect(delay)
}

async function refreshInitAfterConfigChange () {
  if (!window.client) {
    return
  }

  try {
    window.initResponse = await window.client.init({})

    if (typeof window.updateHeaderFromInit === 'function') {
      window.updateHeaderFromInit()
    }
  } catch (err) {
    console.error('Failed to refresh config from server after EventConfigChanged:', err)
  }
}

async function handleConfigChangedEvent (j) {
  await refreshInitAfterConfigChange()
  window.dispatchEvent(j)
}

const eventCaseToTypeName = {
  entityChanged: 'EventEntityChanged',
  configChanged: 'EventConfigChanged',
  executionFinished: 'EventExecutionFinished',
  executionStarted: 'EventExecutionStarted',
  outputChunk: 'EventOutputChunk',
  heartbeat: 'EventHeartbeat'
}

function handleEvent (msg) {
  const eventCase = msg?.event?.case
  const eventValue = msg?.event?.value
  const typeName = eventCaseToTypeName[eventCase]

  if (!typeName || !eventValue) {
    console.warn('Skipping websocket event with no payload:', msg)
    return
  }

  const j = new Event(typeName)
  j.payload = eventValue

  switch (typeName) {
    case 'EventConfigChanged':
      handleConfigChangedEvent(j).catch((err) => {
        console.error('EventConfigChanged handler failed:', err)
      })
      break
    case 'EventHeartbeat':
      break
    case 'EventOutputChunk':
    case 'EventEntityChanged':
      window.dispatchEvent(j)
      break
    case 'EventExecutionFinished':
    case 'EventExecutionStarted':
      window.dispatchEvent(j)
      break
    default:
      console.warn('Unhandled websocket message type from server: ', typeName)

      window.showBigError('ws-unhandled-message', 'handling websocket message', 'Unhandled websocket message type from server: ' + typeName, true)
  }
}

function onOutputChunk (evt) {
  const chunk = evt.payload

  if (window.terminal) {
    if (chunk.executionTrackingId === window.terminal.executionTrackingId) {
      window.terminal.write(chunk.output)
    }
  }
}

export function applyExecutionLogEntry (logEntry) {
  const entry = cloneLogEntry(logEntry)
  if (!entry?.executionTrackingId) {
    return
  }

  buttonResults[entry.executionTrackingId] = entry

  if (entry.datetimeRateLimitExpires && entry.bindingId) {
    const date = new Date(entry.datetimeRateLimitExpires.replace(' ', 'T') + 'Z')
    rateLimits[entry.bindingId] = date.getTime() / 1000
  } else if (entry.bindingId) {
    rateLimits[entry.bindingId] = 0
  }
}

function onExecutionStarted (evt) {
  applyExecutionLogEntry(evt.payload.logEntry)
  applyExecutionStartedBindingState(evt.payload.logEntry)
}

function onExecutionFinished (evt) {
  applyExecutionLogEntry(evt.payload.logEntry)
  applyExecutionFinishedBindingState(evt.payload.logEntry)
}
